package main

import (
	"fmt"
	"net/http"

	"docapp/core/internal/config"
	"docapp/core/internal/database"
	"docapp/core/internal/logger"
	"docapp/core/internal/server"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.Env)

	db, err := database.Connect(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	srv := server.New(cfg, db, log)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Info().Str("addr", addr).Msg("server starting")

	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		log.Fatal().Err(err).Msg("server failed to start")
	}
}
