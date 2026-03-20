package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"
)

type helloResp struct {
	Message string `json:"message"`
	Time    string `json:"time"`
}

func main() {
	var addr = flag.String("addr", "127.0.0.1:8040", "listen addr")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "world"
		}
		writeJSON(w, http.StatusOK, helloResp{
			Message: "hello " + name,
			Time:    time.Now().Format(time.RFC3339),
		})
	})

	srv := &http.Server{
		Addr:         *addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	log.Println("listening on", *addr, "GET /api/hello?name=go-zero")
	log.Fatal(srv.ListenAndServe())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

