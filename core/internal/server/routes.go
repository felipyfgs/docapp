package server

import (
	"encoding/json"
	"net/http"

	"docapp/core/internal/client"
	"docapp/core/internal/handler"

	"github.com/go-chi/chi/v5"
)

type proxyHandler struct {
	client *client.Client
}

func RegisterRoutes(r chi.Router, c *client.Client, empresa *handler.EmpresaHandler, cnpj *handler.CNPJHandler, documento *handler.DocumentoHandler) {
	h := &proxyHandler{client: c}

	r.Route("/fiscal", func(r chi.Router) {
		r.Get("/health", h.handleHealth)
	})

	r.Route("/empresas", func(r chi.Router) {
		r.Get("/", empresa.List)
		r.Post("/", empresa.Create)
		r.Get("/{id}", empresa.GetByID)
		r.Put("/{id}", empresa.Update)
		r.Delete("/{id}", empresa.Delete)
		r.Post("/{id}/certificado", empresa.UploadCertificado)
		r.Post("/{id}/sync", empresa.Sync)
		r.Post("/{id}/import", empresa.Import)
		r.Get("/{id}/overview", empresa.Overview)
		r.Patch("/{id}/nfse", empresa.ToggleNFSe)
	})

	r.Route("/documentos", func(r chi.Router) {
		r.Get("/", documento.List)
		r.Get("/dashboard", documento.Dashboard)
		r.Get("/{id}/xml", documento.XML)
		r.Post("/import", documento.Import)
		r.Post("/export", documento.Export)
		r.Post("/backfill", documento.Backfill)
	})

	r.Get("/cnpj/{cnpj}", cnpj.Lookup)
}

func (h *proxyHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	status, body, err := h.client.Health(r.Context())
	if err != nil {
		h.writeResponse(w, http.StatusBadGateway, mustJSON(map[string]string{
			"message": "Falha ao verificar serviço SPED.",
			"error":   err.Error(),
		}), "application/json")
		return
	}

	h.writeResponse(w, status, body, "application/json")
}

func (h *proxyHandler) writeResponse(w http.ResponseWriter, status int, body []byte, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{"message":"erro interno"}`)
	}
	return b
}
