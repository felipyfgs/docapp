package server

import (
	"net/http"
	"time"

	"docapp/core/internal/client"
	"docapp/core/internal/config"
	"docapp/core/internal/handler"
	"docapp/core/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Server struct {
	router http.Handler
}

func New(cfg *config.Config, db *gorm.DB, log zerolog.Logger) *Server {
	c := client.New(cfg.SpedServiceURL, cfg.SpedTimeoutSeconds)

	empresaService := service.NewEmpresaService(db)
	syncService := service.NewSyncService(db, c, log)
	empresaHandler := handler.NewEmpresaHandler(empresaService, syncService, log)
	cnpjHandler := handler.NewCNPJHandler(log)

	r := chi.NewRouter()
	r.Use(requestLogger(log))
	r.Use(middleware.Recoverer)

	router.RegisterRoutes(r, c, empresaHandler, cnpjHandler)

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
