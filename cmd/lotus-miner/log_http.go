package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

func startMinerLogHTTPServer(addr string) (func(context.Context) error, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/miner_info.log", serveStaticLogFile("miner_info.log"))
	mux.HandleFunc("/proof_info.log", serveStaticLogFile("proof_info.log"))

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("miner log http server failed: %s", err)
		}
	}()

	log.Infof("miner log http server listening on %s", addr)

	return srv.Shutdown, nil
}

func serveStaticLogFile(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, fmt.Sprintf("%s not found", path), http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("read %s failed", path), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}
