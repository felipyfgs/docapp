package server

import (
	"context"
	"net/http"
	"time"

	"docapp/core/internal/client"
	"docapp/core/internal/config"
	"docapp/core/internal/handler"
	"docapp/core/internal/repository"
	"docapp/core/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
)

type Server struct {
	router http.Handler
}

func New(cfg *config.Config, db *bun.DB, log zerolog.Logger) *Server {
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

	empresaRepo := repository.NewEmpresaRepository(db)
	documentoRepo := repository.NewDocumentoRepository(db)

	empresaService := service.NewEmpresaService(empresaRepo)
	syncService := service.NewSyncService(empresaRepo, documentoRepo, c, storage, log)
	documentoService := service.NewDocumentoService(documentoRepo, storage, c, log)
	empresaHandler := handler.NewEmpresaHandler(empresaService, syncService, log)
	documentoHandler := handler.NewDocumentoHandler(documentoService, log)
	cnpjHandler := handler.NewCNPJHandler(log)

	r := chi.NewRouter()
	r.Use(requestLogger(log))
	r.Use(middleware.Recoverer)

	RegisterRoutes(r, c, empresaHandler, cnpjHandler, documentoHandler)

	return &Server{router: r}
}

func (s *Server) Router() http.Handler {
	return s.router
}

func requestLogger(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			defer func() {
				log.Info().
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Int("status", ww.Status()).
					Dur("latency", time.Since(t1)).
					Msg("request")
			}()
			next.ServeHTTP(ww, r)
		})
	}
}
