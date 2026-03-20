package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// This is a learning-friendly "etcd-like" in-memory KV store.
// It demonstrates TTL expiration and key/value APIs without requiring etcd binaries.

type kvValue struct {
	Value     string
	ExpiresAt time.Time // zero => no expiry
}

type kvStore struct {
	mu   sync.RWMutex
	data map[string]kvValue
}

func newKVStore() *kvStore {
	return &kvStore{data: make(map[string]kvValue)}
}

func (s *kvStore) put(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	s.data[key] = kvValue{Value: value, ExpiresAt: exp}
}

func (s *kvStore) get(key string) (string, bool) {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()
	if !ok {
		return "", false
	}
	if !v.ExpiresAt.IsZero() && time.Now().After(v.ExpiresAt) {
		// Lazy delete
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		return "", false
	}
	return v.Value, true
}

func (s *kvStore) expireLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			s.mu.Lock()
			for k, v := range s.data {
				if !v.ExpiresAt.IsZero() && now.After(v.ExpiresAt) {
					delete(s.data, k)
				}
			}
			s.mu.Unlock()
		}
	}
}

func main() {
	var (
		addr = flag.String("addr", "127.0.0.1:8131", "server listen addr")
	)
	flag.Parse()

	store := newKVStore()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go store.expireLoop(ctx)

	mux := http.NewServeMux()

	// PUT /kv  body: {"key":"k1","value":"v1","ttlSeconds":10}
	mux.HandleFunc("/kv", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Key         string `json:"key"`
			Value       string `json:"value"`
			TTLSeconds  int64  `json:"ttlSeconds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
			return
		}
		if req.Key == "" {
			http.Error(w, "missing key", http.StatusBadRequest)
			return
		}
		ttl := time.Duration(req.TTLSeconds) * time.Second
		if req.TTLSeconds <= 0 {
			ttl = 0
		}
		store.put(req.Key, req.Value, ttl)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	// GET /kv?key=k1
	mux.HandleFunc("/kv/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing query param: key", http.StatusBadRequest)
			return
		}
		val, ok := store.get(key)
		if !ok {
			http.Error(w, "not found (or expired)", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"key": key, "value": val})
	})

	log.Println("mock etcd-like server listening on", *addr)
	log.Println("PUT  /kv      {\"key\":\"k1\",\"value\":\"v1\",\"ttlSeconds\":3}")
	log.Println("GET  /kv/get?key=k1")
	log.Println("NOTE: this module is an in-memory mock for learning.")
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
	fmt.Fprintln(w) // keep browsers from caching only the JSON body
}

