package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type order struct {
	ID     int    `json:"id"`
	UserID int    `json:"userId"`
	Title  string `json:"title"`
}

type profile struct {
	User  user   `json:"user"`
	Orders []order `json:"orders"`
}

func main() {
	var (
		userAddr  = flag.String("user-addr", "127.0.0.1:8001", "user service address")
		orderAddr = flag.String("order-addr", "127.0.0.1:8002", "order service address")
		gwAddr    = flag.String("gw-addr", "127.0.0.1:8000", "gateway address")
	)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	muxUser := http.NewServeMux()
	muxUser.HandleFunc("/users/get", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "missing query param: id", http.StatusBadRequest)
			return
		}
		// Demo data
		u := user{ID: mustAtoi(id), Name: fmt.Sprintf("user-%s", id)}
		writeJSON(w, http.StatusOK, u)
	})

	muxOrder := http.NewServeMux()
	muxOrder.HandleFunc("/orders/list", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("userId")
		if userID == "" {
			http.Error(w, "missing query param: userId", http.StatusBadRequest)
			return
		}
		uid := mustAtoi(userID)
		o1 := order{ID: uid*10 + 1, UserID: uid, Title: fmt.Sprintf("order-%d-A", uid)}
		o2 := order{ID: uid*10 + 2, UserID: uid, Title: fmt.Sprintf("order-%d-B", uid)}
		writeJSON(w, http.StatusOK, map[string]any{"orders": []order{o1, o2}})
	})

	gw := http.NewServeMux()
	client := &http.Client{Timeout: 2 * time.Second}
	gw.HandleFunc("/api/profile", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "missing query param: id", http.StatusBadRequest)
			return
		}

		reqUser, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+*userAddr+"/users/get?id="+id, nil)
		respUser, err := client.Do(reqUser)
		if err != nil {
			http.Error(w, "user service error: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer respUser.Body.Close()
		if respUser.StatusCode != http.StatusOK {
			http.Error(w, "user service non-200", http.StatusBadGateway)
			return
		}

		var u user
		if err := json.NewDecoder(respUser.Body).Decode(&u); err != nil {
			http.Error(w, "decode user failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		reqOrder, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+*orderAddr+"/orders/list?userId="+id, nil)
		respOrder, err := client.Do(reqOrder)
		if err != nil {
			http.Error(w, "order service error: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer respOrder.Body.Close()
		if respOrder.StatusCode != http.StatusOK {
			http.Error(w, "order service non-200", http.StatusBadGateway)
			return
		}

		var parsed struct {
			Orders []order `json:"orders"`
		}
		if err := json.NewDecoder(respOrder.Body).Decode(&parsed); err != nil {
			http.Error(w, "decode orders failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, profile{User: u, Orders: parsed.Orders})
	})

	srvUser := &http.Server{Addr: *userAddr, Handler: muxUser}
	srvOrder := &http.Server{Addr: *orderAddr, Handler: muxOrder}
	srvGW := &http.Server{Addr: *gwAddr, Handler: gw}

	go func() {
		log.Println("user service listening on", *userAddr)
		_ = srvUser.ListenAndServe()
	}()
	go func() {
		log.Println("order service listening on", *orderAddr)
		_ = srvOrder.ListenAndServe()
	}()
	log.Println("gateway listening on", *gwAddr, "-> GET /api/profile?id=1")
	_ = srvGW.ListenAndServe()
}

func mustAtoi(s string) int {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	if err != nil {
		return 0
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

