package router

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"docapp/core/internal/client"
	"docapp/core/internal/handler"

	"github.com/go-chi/chi/v5"
)

type proxyHandler struct {
	client *client.Client
}

func RegisterRoutes(r chi.Router, c *client.Client, empresa *handler.EmpresaHandler, cnpj *handler.CNPJHandler) {
	h := &proxyHandler{client: c}

	r.Route("/fiscal", func(r chi.Router) {
		r.Get("/health", h.handleHealth)
		r.Post("/distdfe", h.handleDistDFe)
		r.Post("/download", h.handleDownloadByKey)
		r.Post("/consulta-chave", h.handleConsultaChave)
	})

	r.Route("/empresas", func(r chi.Router) {
		r.Get("/", empresa.List)
		r.Post("/", empresa.Create)
		r.Get("/{id}", empresa.GetByID)
		r.Put("/{id}", empresa.Update)
		r.Delete("/{id}", empresa.Delete)
		r.Post("/{id}/certificado", empresa.UploadCertificado)
	})

	r.Get("/cnpj/{cnpj}", cnpj.Lookup)
}

func (h *proxyHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	status, body, err := h.client.Health(r.Context())
	if err != nil {
		h.writeProxyError(w, err)
		return
	}

	contentType := "text/plain"
	if isJSON(body) {
		contentType = "application/json"
	}

	h.writeResponse(w, status, body, contentType)
}

func (h *proxyHandler) handleDistDFe(w http.ResponseWriter, r *http.Request) {
	h.forwardPost(w, r, h.client.DistDFe)
}

func (h *proxyHandler) handleDownloadByKey(w http.ResponseWriter, r *http.Request) {
	h.forwardPost(w, r, h.client.DownloadByKey)
}

func (h *proxyHandler) handleConsultaChave(w http.ResponseWriter, r *http.Request) {
	h.forwardPost(w, r, h.client.ConsultaChave)
}

func (h *proxyHandler) forwardPost(
	w http.ResponseWriter,
	r *http.Request,
	fn func(ctx context.Context, payload []byte) (int, []byte, error),
) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeResponse(w, http.StatusBadRequest, mustJSON(map[string]string{
			"message": "Payload inválido.",
		}), "application/json")
		return
	}

	status, body, err := fn(r.Context(), payload)
	if err != nil {
		h.writeProxyError(w, err)
		return
	}

	if !isJSON(body) {
		h.writeResponse(w, status, body, "text/plain")
		return
	}

	h.writeResponse(w, status, body, "application/json")
}

func (h *proxyHandler) writeProxyError(w http.ResponseWriter, err error) {
	h.writeResponse(w, http.StatusBadGateway, mustJSON(map[string]string{
		"message": "Falha ao chamar microserviço SPED.",
		"error":   err.Error(),
	}), "application/json")
}

func (h *proxyHandler) writeResponse(w http.ResponseWriter, status int, body []byte, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func isJSON(body []byte) bool {
	trimmed := strings.TrimSpace(string(body))
	return strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{"message":"erro interno"}`)
	}
	return b
}
