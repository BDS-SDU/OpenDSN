package main

import (
	"net/http"
	"time"
)

type Server struct {
	cfg        Config
	httpClient *http.Client
}

func NewServer(cfg Config) *Server {
	return &Server{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/miners", s.handleMiners)
	mux.HandleFunc("/api/proofs", s.handleProofs)
	mux.HandleFunc("/api/client/import", s.handleClientImport)
	mux.HandleFunc("/api/client/deal", s.handleClientDeal)
	mux.HandleFunc("/api/client/retrieve-version", s.handleClientRetrieveVersion)
	mux.HandleFunc("/api/client/roots", s.handleClientRoots)
	mux.HandleFunc("/api/client/versions", s.handleClientVersions)

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
