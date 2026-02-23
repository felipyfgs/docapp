package main

import (
	"context"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"docapp/core/internal/client"
	"docapp/core/internal/config"
	"docapp/core/internal/db"
	"docapp/core/internal/repository"
	"docapp/core/internal/service"

	"github.com/rs/zerolog"
)

func main() {
	cfg := config.Load()
	log := config.NewLogger(cfg.Env)

	bunDB, err := db.Connect(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	c := client.New(cfg.SpedServiceURL, cfg.SpedTimeoutSeconds)

	var storage service.DocumentStorage
	minioStorage, err := service.NewMinioStorage(cfg)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize minio storage")
	} else {
		if err := minioStorage.EnsureBucket(context.Background()); err != nil {
			log.Error().Err(err).Msg("failed to ensure storage bucket")
		} else {
			storage = minioStorage
		}
	}

	empresaRepo := repository.NewEmpresaRepository(bunDB)
	documentoRepo := repository.NewDocumentoRepository(bunDB)
	empresaService := service.NewEmpresaService(empresaRepo)
	syncService := service.NewSyncService(empresaRepo, documentoRepo, c, storage, log)

	interval := time.Duration(cfg.WorkerIntervalMinutes) * time.Minute
	log.Info().Dur("interval", interval).Str("sped_url", cfg.SpedServiceURL).Msg("worker starting")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	run(log, empresaService, syncService)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			run(log, empresaService, syncService)
		case sig := <-sigChan:
			log.Info().Str("signal", sig.String()).Msg("shutting down worker")
			return
		}
	}
}

func run(log zerolog.Logger, empresaService *service.EmpresaService, syncService *service.SyncService) {
	start := time.Now()

	empresas, err := empresaService.ListAtivasComCertificado()
	if err != nil {
		log.Error().Err(err).Msg("worker: list empresas failed")
		return
	}

	if len(empresas) == 0 {
		log.Info().Msg("worker: no empresas with certificates to sync")
		return
	}

	sort.SliceStable(empresas, func(i, j int) bool {
		iNew := empresas[i].UltimaSincronizacao == nil
		jNew := empresas[j].UltimaSincronizacao == nil
		return iNew && !jNew
	})

	log.Info().Int("total", len(empresas)).Msg("worker: starting sync cycle")

	successCount := 0
	errorCount := 0

	for _, empresa := range empresas {
		if err := syncService.SyncEmpresa(empresa); err != nil {
			log.Error().Err(err).Uint("empresa_id", empresa.ID).Str("cnpj", empresa.CNPJ).Msg("worker: sync failed")
			errorCount++
		} else {
			successCount++
		}
	}

	log.Info().
		Int("success", successCount).
		Int("errors", errorCount).
		Dur("duration", time.Since(start)).
		Msg("worker: sync cycle completed")
}
