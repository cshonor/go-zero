package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"
)

type addResp struct {
	Sum int64 `json:"sum"`
}

type errResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	var addr = flag.String("addr", "127.0.0.1:8050", "listen addr")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeErr(w, http.StatusMethodNotAllowed, 10001, "method not allowed")
			return
		}
		aStr := r.URL.Query().Get("a")
		bStr := r.URL.Query().Get("b")
		if aStr == "" || bStr == "" {
			writeErr(w, http.StatusBadRequest, 10002, "missing query params: a and/or b")
			return
		}
		a, err1 := strconv.ParseInt(aStr, 10, 64)
		b, err2 := strconv.ParseInt(bStr, 10, 64)
		if err1 != nil || err2 != nil {
			writeErr(w, http.StatusBadRequest, 10003, "a/b must be integers")
			return
		}
		writeJSON(w, http.StatusOK, addResp{Sum: a + b})
	})

	log.Println("listening on", *addr, "GET /api/add?a=1&b=2")
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func writeErr(w http.ResponseWriter, httpStatus, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(errResp{Code: code, Message: msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

