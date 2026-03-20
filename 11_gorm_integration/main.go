package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type user struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

type gormRepo struct {
	mu     sync.RWMutex
	nextID int64
	users  map[int64]user
}

func newGormRepo() *gormRepo {
	return &gormRepo{nextID: 1, users: make(map[int64]user)}
}

func (r *gormRepo) Create(name string) user {
	r.mu.Lock()
	defer r.mu.Unlock()
	u := user{ID: r.nextID, Name: name, CreatedAt: time.Now().Format(time.RFC3339)}
	r.users[u.ID] = u
	r.nextID++
	return u
}

type Query struct {
	repo      *gormRepo
	nameLike  string
}

func (r *gormRepo) WhereNameLike(substr string) *Query {
	return &Query{repo: r, nameLike: substr}
}

// First returns the first matched record (like gorm's First()).
func (q *Query) First() (user, bool) {
	q.repo.mu.RLock()
	defer q.repo.mu.RUnlock()
	for _, u := range q.repo.users {
		if q.nameLike == "" || strings.Contains(strings.ToLower(u.Name), strings.ToLower(q.nameLike)) {
			return u, true
		}
	}
	return user{}, false
}

// Find returns all matched records (like gorm's Find()).
func (q *Query) Find() []user {
	q.repo.mu.RLock()
	defer q.repo.mu.RUnlock()
	var out []user
	for _, u := range q.repo.users {
		if q.nameLike == "" || strings.Contains(strings.ToLower(u.Name), strings.ToLower(q.nameLike)) {
			out = append(out, u)
		}
	}
	return out
}

func main() {
	var addr = flag.String("addr", "127.0.0.1:8110", "listen addr")
	flag.Parse()

	repo := newGormRepo()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var req struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErr(w, http.StatusBadRequest, 10001, "bad json: "+err.Error())
				return
			}
			if strings.TrimSpace(req.Name) == "" {
				writeErr(w, http.StatusBadRequest, 10002, "missing name")
				return
			}
			u := repo.Create(req.Name)
			writeJSON(w, http.StatusOK, map[string]any{"user": u})
		case http.MethodGet:
			like := r.URL.Query().Get("name")
			q := repo.WhereNameLike(like)

			// If caller sets first=true, behave like First().
			first := r.URL.Query().Get("first")
			if strings.EqualFold(first, "true") {
				u, ok := q.First()
				if !ok {
					writeErr(w, http.StatusNotFound, 10003, "not found")
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"user": u})
				return
			}

			users := q.Find()
			writeJSON(w, http.StatusOK, map[string]any{"users": users})
		default:
			writeErr(w, http.StatusMethodNotAllowed, 10009, "method not allowed")
		}
	})

	log.Println("listening on", *addr)
	log.Println("POST /api/users {\"name\":\"alice\"}")
	log.Println("GET  /api/users?name=ali&first=true")
	log.Println("GET  /api/users?name=ali")
	log.Fatal(http.ListenAndServe(*addr, mux))
}

type apiErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func writeErr(w http.ResponseWriter, httpStatus, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(apiErr{Code: code, Message: msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

