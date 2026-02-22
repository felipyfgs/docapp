package server

import (
	"encoding/json"
	"net/http"
	"time"

	"docapp/core/internal/client"
	"docapp/core/internal/config"
	"docapp/core/internal/proxy"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	cfg        *config.Config
	router     chi.Router
	spedClient *client.Client
}

func New(cfg *config.Config) *Server {
	s := &Server{
		cfg:        cfg,
		spedClient: client.New(cfg.SpedServiceURL, time.Duration(cfg.SpedTimeoutSeconds)*time.Second),
	}
	s.setupRoutes()
	return s
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) setupRoutes() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/health", s.handleHealth)

	r.Route("/api/v1", func(r chi.Router) {
		proxy.RegisterRoutes(r, s.spedClient)
	})

	s.router = r
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
