package main

import (
	"fmt"
	"log"
	"net/http"

	"docapp/core/internal/config"
	"docapp/core/internal/server"
)

func main() {
	cfg := config.Load()

	srv := server.New(cfg)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)

	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
