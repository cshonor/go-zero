package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"
)

// --- RPC types ---

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

// --- gorm-like repo layer (in-memory mock) ---

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

type UserQuery struct {
	repo *userRepo
	id   int
}

func (r *userRepo) WhereID(id int) *UserQuery {
	return &UserQuery{repo: r, id: id}
}

func (q *UserQuery) First() (User, error) {
	u, ok := q.repo.data[q.id]
	if !ok {
		return User{}, errors.New("record not found")
	}
	return u, nil
}

// --- RPC service ---

type UserService struct {
	repo *userRepo
}

func (s *UserService) GetUser(req *GetUserReq, resp *GetUserResp) error {
	if req == nil {
		return fmt.Errorf("nil request")
	}
	if req.ID <= 0 {
		return fmt.Errorf("invalid id: %d", req.ID)
	}
	u, err := s.repo.WhereID(req.ID).First()
	if err != nil {
		return err
	}
	resp.User = u
	return nil
}

func main() {
	var (
		mode = flag.String("mode", "server", "server|client")
		addr = flag.String("addr", "127.0.0.1:9014", "rpc listen/dial addr")
		id   = flag.Int("id", 1, "client: user id")
	)
	flag.Parse()

	switch *mode {
	case "server":
		startServer(*addr)
	case "client":
		startClient(*addr, *id)
	default:
		log.Fatal("unknown -mode: " + *mode)
	}
}

func startServer(addr string) {
	repo := newUserRepo()
	svc := &UserService{repo: repo}

	if err := rpc.RegisterName("user", svc); err != nil {
		log.Fatal("register rpc service failed:", err)
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("listen failed:", err)
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

func startClient(addr string, id int) {
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Fatal("dial failed:", err)
	}
	defer func() { _ = client.Close() }()

	var out GetUserResp
	req := GetUserReq{ID: id}
	if err := client.Call("user.GetUser", &req, &out); err != nil {
		log.Fatal("call failed:", err)
	}
	log.Println("rpc result:", out.User)
	time.Sleep(100 * time.Millisecond)
}

