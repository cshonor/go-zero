package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
)

type resp[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

func main() {
	var addr = flag.String("addr", "127.0.0.1:8060", "listen addr")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, resp[string]{Code: 0, Message: "ok", Data: "pong"})
	})
	mux.HandleFunc("/api/fail", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, resp[any]{Code: 20001, Message: "something went wrong", Data: nil})
	})

	log.Println("listening on", *addr, "GET /api/ping or /api/fail")
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

