package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type user struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

// mysqlRepo is an educational "MySQL-like" repository.
// For learning, we keep it in-memory to avoid external DB setup.
type mysqlRepo struct {
	mu     sync.RWMutex
	nextID int64
	users  map[int64]user
}

func newMySQLRepo() *mysqlRepo {
	return &mysqlRepo{
		nextID: 1,
		users:  make(map[int64]user),
	}
}

func (r *mysqlRepo) create(name string) user {
	r.mu.Lock()
	defer r.mu.Unlock()
	u := user{
		ID:        r.nextID,
		Name:      name,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	r.users[u.ID] = u
	r.nextID++
	return u
}

func (r *mysqlRepo) getByID(id int64) (user, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	return u, ok
}

func (r *mysqlRepo) deleteByID(id int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[id]; !ok {
		return false
	}
	delete(r.users, id)
	return true
}

func main() {
	var addr = flag.String("addr", "127.0.0.1:8100", "listen addr")
	flag.Parse()

	mysqlDSN := os.Getenv("MYSQL_DSN")
	if mysqlDSN == "" {
		log.Println("MYSQL_DSN not set: this module uses an in-memory mock store.")
	} else {
		log.Println("MYSQL_DSN is set (not used in mock mode):", mysqlDSN)
	}

	repo := newMySQLRepo()

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
			if req.Name == "" {
				writeErr(w, http.StatusBadRequest, 10002, "missing name")
				return
			}
			u := repo.create(req.Name)
			writeJSON(w, http.StatusOK, map[string]any{"user": u})
		case http.MethodGet:
			idStr := r.URL.Query().Get("id")
			if idStr == "" {
				writeErr(w, http.StatusBadRequest, 10003, "missing query param: id")
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				writeErr(w, http.StatusBadRequest, 10004, "id must be integer")
				return
			}
			u, ok := repo.getByID(id)
			if !ok {
				writeErr(w, http.StatusNotFound, 10005, "user not found")
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"user": u})
		case http.MethodDelete:
			idStr := r.URL.Query().Get("id")
			if idStr == "" {
				writeErr(w, http.StatusBadRequest, 10006, "missing query param: id")
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				writeErr(w, http.StatusBadRequest, 10007, "id must be integer")
				return
			}
			ok := repo.deleteByID(id)
			if !ok {
				writeErr(w, http.StatusNotFound, 10008, "user not found")
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
		default:
			writeErr(w, http.StatusMethodNotAllowed, 10009, "method not allowed")
		}
	})

	log.Println("listening on", *addr)
	log.Println("POST   /api/users {\"name\":\"alice\"}")
	log.Println("GET    /api/users?id=1")
	log.Println("DELETE /api/users?id=1")
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

