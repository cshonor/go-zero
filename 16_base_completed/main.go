package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	var only = flag.String("only", "", "optional: 01~16 (e.g. -only 07)")
	flag.Parse()

	steps := []struct {
		dir   string
		title string
		run   string
	}{
		{"01_microservices", "microservices demo (HTTP gateway + 2 services)", "go run ./01_microservices"},
		{"02_env_setup", "environment checklist", "go run ./02_env_setup"},
		{"03_etcd_installation", "etcd-like in-memory KV with TTL", "go run ./03_etcd_installation"},
		{"04_first_microservice_demo", "hello API", "go run ./04_first_microservice_demo"},
		{"05_api_syntax", "request parsing + validation + error JSON", "go run ./05_api_syntax"},
		{"06_api_response_integration", "unified response wrapper", "go run ./06_api_response_integration"},
		{"07_jwt_authentication", "JWT login + protected endpoint", "go run ./07_jwt_authentication"},
		{"08_jwt_failure_response", "JWT failure unified error code", "go run ./08_jwt_failure_response"},
		{"09_api_documentation", "manual OpenAPI JSON", "go run ./09_api_documentation"},
		{"10_mysql_integration", "MySQL-like CRUD (in-memory mock)", "go run ./10_mysql_integration"},
		{"11_gorm_integration", "gorm-like query style (in-memory mock)", "go run ./11_gorm_integration"},
		{"12_rpc_service", "RPC server/client (net/rpc)", "go run ./12_rpc_service -mode server   ; go run ./12_rpc_service -mode client"},
		{"13_rpc_grouping", "RPC multiple services (user/math)", "go run ./13_rpc_grouping -mode server ; go run ./13_rpc_grouping -mode client"},
		{"14_rpc_gorm_integration", "RPC method uses gorm-like repo", "go run ./14_rpc_gorm_integration -mode server ; go run ./14_rpc_gorm_integration -mode client"},
		{"15_rpc_api_integration", "API calls RPC (HTTP -> net/rpc)", "go run ./15_rpc_api_integration -mode rpc ; go run ./15_rpc_api_integration -mode api"},
		{"16_base_completed", "this summary tool", "go run ./16_base_completed"},
	}

	if strings.TrimSpace(*only) != "" {
		printOne(steps, *only)
		return
	}

	fmt.Println("Go-Zero Learning: Base Modules Summary")
	fmt.Println("========================================")
	fmt.Println("Tip: run each module from repo root with `go run ./<dir>`")
	fmt.Println()

	for _, s := range steps {
		if s.dir == "16_base_completed" {
			continue
		}
		fmt.Printf("- %-22s : %s\n", s.dir, s.title)
		fmt.Println("  Run:", s.run)
		fmt.Println()
	}

	log.Println("Done. If you want, tell me which chapter you want to make 'real' (etcd/mysql/gorm/proto), and I will upgrade the mock to real integrations.")
	_ = os.Getenv("DUMMY") // keep os imported explicitly for clarity
}

func printOne(steps []struct {
	dir   string
	title string
	run   string
}, only string) {
	// only can be "07" etc
	target := "0" + only
	if len(only) == 2 {
		target = "0" + only
	} else {
		target = only
	}
	for _, s := range steps {
		if strings.HasPrefix(s.dir, target) || s.dir == "16_base_completed" && only == "16" {
			fmt.Printf("%s (%s)\nRun: %s\n", s.dir, s.title, s.run)
			return
		}
	}
	fmt.Println("No match for -only:", only)
}

