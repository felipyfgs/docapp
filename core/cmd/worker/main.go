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
	nfseSyncService := service.NewNFSeSyncService(empresaRepo, documentoRepo, storage, log, cfg.ADNBaseURL)

	interval := time.Duration(cfg.WorkerIntervalMinutes) * time.Minute
	log.Info().Dur("interval", interval).Str("sped_url", cfg.SpedServiceURL).Msg("worker starting")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	run(log, empresaService, syncService, nfseSyncService)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			run(log, empresaService, syncService, nfseSyncService)
		case sig := <-sigChan:
			log.Info().Str("signal", sig.String()).Msg("shutting down worker")
			return
		}
	}
}

func run(log zerolog.Logger, empresaService *service.EmpresaService, syncService *service.SyncService, nfseSyncService *service.NFSeSyncService) {
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
			log.Error().Err(err).Uint("empresa_id", empresa.ID).Str("cnpj", empresa.CNPJ).Msg("worker: nfe sync failed")
			errorCount++
		} else {
			successCount++
		}

		if empresa.NFSeHabilitada {
			if err := nfseSyncService.SyncEmpresaNFSe(empresa); err != nil {
				log.Error().Err(err).Uint("empresa_id", empresa.ID).Str("cnpj", empresa.CNPJ).Msg("worker: nfse sync failed")
			}
		}
	}

	log.Info().
		Int("success", successCount).
		Int("errors", errorCount).
		Dur("duration", time.Since(start)).
		Msg("worker: sync cycle completed")
}
