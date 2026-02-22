package main

import (
	"time"

	"docapp/core/internal/client"
	"docapp/core/internal/config"
	"docapp/core/internal/database"
	"docapp/core/internal/logger"
	"docapp/core/internal/service"

	"github.com/rs/zerolog"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.Env)

	db, err := database.Connect(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	c := client.New(cfg.SpedServiceURL, cfg.SpedTimeoutSeconds)
	empresaService := service.NewEmpresaService(db)
	syncService := service.NewSyncService(db, c, log)

	interval := time.Duration(cfg.WorkerIntervalMinutes) * time.Minute
	log.Info().Dur("interval", interval).Msg("worker starting")

	run(log, empresaService, syncService)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		run(log, empresaService, syncService)
	}
}

func run(log zerolog.Logger, empresaService *service.EmpresaService, syncService *service.SyncService) {
	empresas, err := empresaService.ListAtivas()
	if err != nil {
		log.Error().Err(err).Msg("worker: list empresas failed")
		return
	}

	log.Info().Int("total", len(empresas)).Msg("worker: syncing empresas")

	for _, empresa := range empresas {
		if err := syncService.SyncEmpresa(empresa); err != nil {
			log.Error().Err(err).Uint("empresa_id", empresa.ID).Str("cnpj", empresa.CNPJ).Msg("worker: sync failed")
		}
	}
}

