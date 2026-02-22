package main

import (
	"docapp/core/internal/client"
	"docapp/core/internal/config"
	"docapp/core/internal/database"
	"docapp/core/internal/logger"
	"docapp/core/internal/model"
	"docapp/core/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.Env)

	db, err := database.Connect(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	c := client.New("http://localhost:8000/api", cfg.SpedTimeoutSeconds)
	syncService := service.NewSyncService(db, c, log)

	var empresa model.Empresa
	if err := db.Where("cnpj = ?", "50824718000170").First(&empresa).Error; err != nil {
		log.Fatal().Err(err).Msg("empresa not found")
	}

	log.Info().Str("cnpj", empresa.CNPJ).Msg("starting manual sync")
	if err := syncService.SyncEmpresa(empresa); err != nil {
		log.Fatal().Err(err).Msg("sync failed")
	}

	log.Info().Msg("sync completed successfully")
}
