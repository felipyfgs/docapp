package main

import (
	"fmt"
	"net/http"

	"docapp/core/internal/config"
	"docapp/core/internal/db"
	"docapp/core/internal/repository"
	"docapp/core/internal/server"
	"docapp/core/internal/service"
)

func main() {
	cfg := config.Load()
	log := config.NewLogger(cfg.Env)

	bunDB, err := db.Connect(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	empresaRepo := repository.NewEmpresaRepository(bunDB)
	empresaSvc := service.NewEmpresaService(empresaRepo)
	if err := empresaSvc.AtualizarValidadeCertificados(); err != nil {
		log.Warn().Err(err).Msg("failed to update certificate validity")
	}

	srv := server.New(cfg, bunDB, log)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Info().Str("addr", addr).Msg("server starting")

	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		log.Fatal().Err(err).Msg("server failed to start")
	}
}
