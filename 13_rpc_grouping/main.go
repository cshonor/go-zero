package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"
)

// --- user service ---

type UserReq struct {
	ID int
}

type UserResp struct {
	Name string
}

type UserService struct{}

func (s *UserService) GetName(req *UserReq, resp *UserResp) error {
	if req == nil || req.ID <= 0 {
		return fmt.Errorf("invalid id")
	}
	resp.Name = fmt.Sprintf("user-%d", req.ID)
	return nil
}

// --- math service ---

type AddReq struct {
	A int
	B int
}

type AddResp struct {
	Sum int
}

type MathService struct{}

func (s *MathService) Add(req *AddReq, resp *AddResp) error {
	if req == nil {
		return fmt.Errorf("nil request")
	}
	resp.Sum = req.A + req.B
	return nil
}

func main() {
	var (
		mode = flag.String("mode", "server", "server|client")
		addr = flag.String("addr", "127.0.0.1:9013", "rpc listen/dial addr")
		id   = flag.Int("id", 2, "client: user id")
		a    = flag.Int("a", 10, "client: add a")
		b    = flag.Int("b", 20, "client: add b")
	)
	flag.Parse()

	switch *mode {
	case "server":
		startServer(*addr)
	case "client":
		startClient(*addr, *id, *a, *b)
	default:
		log.Fatal("unknown -mode: " + *mode)
	}
}

func startServer(addr string) {
	if err := rpc.RegisterName("user", new(UserService)); err != nil {
		log.Fatal("register user service failed:", err)
	}
	if err := rpc.RegisterName("math", new(MathService)); err != nil {
		log.Fatal("register math service failed:", err)
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("listen failed:", err)
	}
	log.Println("rpc server listening on", addr, "(services: user, math)")
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}
}

func startClient(addr string, id, a, b int) {
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		log.Fatal("dial failed:", err)
	}
	defer func() { _ = client.Close() }()

	var ur UserResp
	if err := client.Call("user.GetName", &UserReq{ID: id}, &ur); err != nil {
		log.Fatal("call user.GetName failed:", err)
	}

	var ar AddResp
	if err := client.Call("math.Add", &AddReq{A: a, B: b}, &ar); err != nil {
		log.Fatal("call math.Add failed:", err)
	}

	log.Println("user.GetName result:", ur.Name)
	log.Println("math.Add result:", ar.Sum)
	time.Sleep(100 * time.Millisecond)
}

