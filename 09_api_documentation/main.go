package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

type openAPISpec struct {
	OpenAPI string                          `json:"openapi"`
	Info    map[string]any                  `json:"info"`
	Paths   map[string]map[string]any      `json:"paths"`
}

func main() {
	var addr = flag.String("addr", "127.0.0.1:8090", "listen addr")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "world"
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"message": "hello " + name,
			"time":    time.Now().Format(time.RFC3339),
		})
	})
	mux.HandleFunc("/api/time", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"time": time.Now().Format(time.RFC3339)})
	})

	mux.HandleFunc("/docs/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		spec := openAPISpec{
			OpenAPI: "3.0.0",
			Info: map[string]any{
				"title":   "go-zero learning docs (manual)",
				"version": "1.0.0",
			},
			Paths: map[string]map[string]any{
				"/api/hello": {
					"get": map[string]any{
						"summary": "Hello endpoint",
						"parameters": []map[string]any{
							{
								"name":     "name",
								"in":       "query",
								"required": false,
								"schema":   map[string]any{"type": "string"},
							},
						},
						"responses": map[string]any{
							"200": map[string]any{"description": "ok"},
						},
					},
				},
				"/api/time": {
					"get": map[string]any{
						"summary": "Get current server time",
						"responses": map[string]any{
							"200": map[string]any{"description": "ok"},
						},
					},
				},
			},
		}
		writeJSON(w, http.StatusOK, spec)
	})

	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, "<html><body>")
		fmt.Fprintln(w, "<h3>API Docs (learning)</h3>")
		fmt.Fprintln(w, "<ul>")
		fmt.Fprintln(w, "<li><code>GET /api/hello</code></li>")
		fmt.Fprintln(w, "<li><code>GET /api/time</code></li>")
		fmt.Fprintln(w, "</ul>")
		fmt.Fprintln(w, "<p>OpenAPI JSON: <a href=\"/docs/openapi.json\">/docs/openapi.json</a></p>")
		fmt.Fprintln(w, "</body></html>")
	})

	log.Println("listening on", *addr)
	log.Println("try: GET /api/hello?name=go-zero")
	log.Println("try: GET /docs/openapi.json")
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

