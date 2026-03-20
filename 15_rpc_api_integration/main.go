package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

type GetUserReq struct {
	ID int
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type GetUserResp struct {
	User User `json:"user"`
}

type userRepo struct {
	data map[int]User
}

func newUserRepo() *userRepo {
	return &userRepo{
		data: map[int]User{
			1: {ID: 1, Name: "alice"},
			2: {ID: 2, Name: "bob"},
		},
	}
}

type UserService struct {
	repo *userRepo
}

func (s *UserService) GetUser(req *GetUserReq, resp *GetUserResp) error {
	if req == nil {
		return errors.New("nil request")
	}
	if req.ID <= 0 {
		return fmt.Errorf("invalid id: %d", req.ID)
	}
	u, ok := s.repo.data[req.ID]
	if !ok {
		return fmt.Errorf("user not found: %d", req.ID)
	}
	resp.User = u
	return nil
}

func main() {
	var (
		mode      = flag.String("mode", "api", "api|rpc")
		httpAddr  = flag.String("http-addr", "127.0.0.1:8150", "api listen addr")
		rpcAddr   = flag.String("rpc-addr", "127.0.0.1:9015", "rpc server addr")
	)
	flag.Parse()

	switch *mode {
	case "rpc":
		startRPCServer(*rpcAddr)
	case "api":
		startAPIServer(*httpAddr, *rpcAddr)
	default:
		log.Fatal("unknown -mode: " + *mode)
	}
}

func startRPCServer(addr string) {
	repo := newUserRepo()
	svc := &UserService{repo: repo}

	if err := rpc.RegisterName("user", svc); err != nil {
		log.Fatal("register rpc service failed:", err)
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("rpc server listening on", addr, "(service: user)")
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}
}

func startAPIServer(httpAddr, rpcAddr string) {
	mux := http.NewServeMux()
	clientFactory := func() (*rpc.Client, error) {
		return rpc.Dial("tcp", rpcAddr)
	}

	mux.HandleFunc("/api/user", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"code": 10001, "message": "method not allowed"})
			return
		}
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"code": 10002, "message": "missing query param: id"})
			return
		}
		var id int
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil || id <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"code": 10003, "message": "id must be positive int"})
			return
		}

		c, err := clientFactory()
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"code": 10004, "message": "rpc dial failed: " + err.Error()})
			return
		}
		defer c.Close()

		var out GetUserResp
		req := GetUserReq{ID: id}
		err = c.Call("user.GetUser", &req, &out)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"code": 10005, "message": "rpc call failed: " + err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"code": 0, "message": "ok", "data": out.User})
	})

	srv := &http.Server{
		Addr:         httpAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	log.Println("api server listening on", httpAddr, "GET /api/user?id=1 (calls rpc at", rpcAddr, ")")
	log.Fatal(srv.ListenAndServe())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

