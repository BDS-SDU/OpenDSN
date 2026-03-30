package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	repoRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("get working directory: %v", err)
	}

	if v := os.Getenv("OPENDSN_REPO_ROOT"); v != "" {
		repoRoot = v
	}

	addr := "0.0.0.0:8080"
	if v := os.Getenv("OPENDSN_API_ADDR"); v != "" {
		addr = v
	}

	minerHTTPPort := 18080
	if v := os.Getenv("OPENDSN_MINER_HTTP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			minerHTTPPort = p
		}
	}

	cfg := Config{
		Addr:          addr,
		RepoRoot:      repoRoot,
		LotusBin:      "./lotus",
		MinerHTTPPort: minerHTTPPort,
		DealPrice:     "0.026",
		DealDuration:  "518400",
	}

	srv := NewServer(cfg)

	log.Printf("opendsn-api listening on %s", cfg.Addr)
	log.Printf("repo root: %s", cfg.RepoRoot)

	if err := http.ListenAndServe(cfg.Addr, srv.Routes()); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}
