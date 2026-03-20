package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"
)

type UserReq struct {
	ID int
}

type UserResp struct {
	Name string
}

type UserService struct{}

func (s *UserService) GetName(req *UserReq, resp *UserResp) error {
	if req == nil || req.ID <= 0 {
		return fmt.Errorf("invalid id: %d", req.GetIDSafe())
	}
	resp.Name = fmt.Sprintf("user-%d", req.ID)
	return nil
}

func (r *UserReq) GetIDSafe() int {
	if r == nil {
		return 0
	}
	return r.ID
}

func main() {
	var (
		mode = flag.String("mode", "server", "server|client")
		addr = flag.String("addr", "127.0.0.1:9012", "rpc listen/dial addr")
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
	svc := new(UserService)
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

	var out UserResp
	req := UserReq{ID: id}
	if err := client.Call("user.GetName", &req, &out); err != nil {
		log.Fatal("call failed:", err)
	}
	log.Println("rpc result:", out.Name)
	time.Sleep(100 * time.Millisecond)
}

