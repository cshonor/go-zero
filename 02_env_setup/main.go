package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"time"
)

func main() {
	var (
		onlyPrint = flag.Bool("print-only", true, "print checklist and exit")
	)
	flag.Parse()

	if !*onlyPrint {
		log.Println("This demo is intended as a checklist tool; run with -print-only.")
	}

	fmt.Println("Environment Checklist (learning tool)")
	fmt.Println("--------------------------------------")
	fmt.Println("OS:", runtime.GOOS, "ARCH:", runtime.GOARCH)
	fmt.Println("Go Version:", runtime.Version())
	fmt.Println("Go Env Sample:")
	fmt.Println("  GOPATH:", getenv("GOPATH", ""))
	fmt.Println("  GOMOD:", os.Getenv("GOMOD"))
	fmt.Println("  GOMODCACHE:", getenv("GOMODCACHE", ""))

	// Ports quick scan (just educational)
	ports := []string{"8000", "8001", "8002", "8070", "9012", "9013", "9014", "9015"}
	fmt.Println()
	fmt.Println("Suggested learning ports (may already be in use):")
	for _, p := range ports {
		inUse := portInUse(p)
		fmt.Printf("  - %s: %s\n", p, map[bool]string{true: "IN USE", false: "free/unknown"}[inUse])
	}

	fmt.Println()
	fmt.Println("Common environment variables (for later chapters):")
	fmt.Println("  JWT_SECRET (default: secret)")
	fmt.Println("  ETCD_ENDPOINT (default: 127.0.0.1:2379) - mock in this repo step")
	fmt.Println("  MYSQL_DSN (optional) - this learning module uses an in-memory mock")
	fmt.Println("  RPC_ADDR (default: 127.0.0.1:9015)")
	fmt.Println()
	fmt.Println("Run examples:")
	fmt.Println("  go run ./01_microservices")
	fmt.Println("  go run ./04_first_microservice_demo")
	fmt.Println("  go run ./07_jwt_authentication -mode server")
	fmt.Println()
	fmt.Printf("Generated at %s (%s)\n", time.Now().Format(time.RFC3339), hostname())
}

func getenv(k, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return fallback
}

func hostname() string {
	h, _ := os.Hostname()
	if h == "" {
		return "unknown-host"
	}
	return h
}

func portInUse(port string) bool {
	ln, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		return true
	}
	_ = ln.Close()
	return false
}

