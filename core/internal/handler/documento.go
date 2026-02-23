package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"docapp/core/internal/repository"
	"docapp/core/internal/service"

	"github.com/rs/zerolog"
)

type DocumentoHandler struct {
	svc *service.DocumentoService
	log zerolog.Logger
}

type exportRequest struct {
	IDs          []uint `json:"ids"`
	Format       string `json:"format"`
	Organization string `json:"organization"`
	DeliveryMode string `json:"delivery_mode"`
}

type backfillRequest struct {
	Limit int `json:"limit"`
}

func NewDocumentoHandler(svc *service.DocumentoService, log zerolog.Logger) *DocumentoHandler {
	return &DocumentoHandler{svc: svc, log: log}
}

func (h *DocumentoHandler) List(w http.ResponseWriter, r *http.Request) {
	page := parseQueryInt(r, "page", 1)
	pageSize := parseQueryInt(r, "page_size", 20)

	var empresaID uint
	if raw := strings.TrimSpace(r.URL.Query().Get("empresa_id")); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "empresa_id inválido."})
			return
		}
		empresaID = uint(parsed)
	}

	var xmlResumo *bool
	if raw := strings.TrimSpace(r.URL.Query().Get("xml_resumo")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "xml_resumo inválido."})
			return
		}
		xmlResumo = &parsed
	}

	docs, total, err := h.svc.List(service.DocumentoListFilter{
		Search:    r.URL.Query().Get("search"),
		Tipo:      r.URL.Query().Get("tipo"),
		Status:    r.URL.Query().Get("status"),
		EmpresaID: empresaID,
		XMLResumo: xmlResumo,
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		h.log.Error().Err(err).Msg("list documentos failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Erro ao listar documentos."})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":     docs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *DocumentoHandler) XML(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "ID inválido."})
		return
	}

	doc, err := h.svc.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "Documento não encontrado."})
			return
		}

		h.log.Error().Err(err).Uint("id", id).Msg("get documento failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Erro ao buscar documento."})
		return
	}

	xmlContent, err := h.svc.ReadXML(r.Context(), doc)
	if err != nil {
		h.log.Error().Err(err).Uint("id", id).Msg("read documento xml failed")
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "XML não encontrado."})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":           doc.ID,
		"xml":          xmlContent,
		"xml_resumo":   doc.XMLResumo,
		"chave_acesso": doc.ChaveAcesso,
	})
}

func (h *DocumentoHandler) Export(w http.ResponseWriter, r *http.Request) {
	var req exportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Payload inválido."})
		return
	}

	result, err := h.svc.Export(r.Context(), service.DocumentoExportOptions{
		IDs:          req.IDs,
		Format:       req.Format,
		Organization: req.Organization,
		DeliveryMode: req.DeliveryMode,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	if result.Mode == service.ExportDeliveryPresigned {
		writeJSON(w, http.StatusOK, result)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+result.FileName)
	w.Header().Set("X-Export-Total", strconv.Itoa(result.Total))
	w.Header().Set("X-Export-XML", strconv.Itoa(result.XMLCount))
	w.Header().Set("X-Export-DANFE", strconv.Itoa(result.DanfeCount))
	w.Header().Set("X-Export-Skipped-DANFE", strconv.Itoa(result.SkippedDanfe))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result.Content)
}

func (h *DocumentoHandler) Backfill(w http.ResponseWriter, r *http.Request) {
	var req backfillRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Payload inválido."})
			return
		}
	}

	result, err := h.svc.Backfill(r.Context(), req.Limit)
	if err != nil {
		h.log.Error().Err(err).Msg("documentos backfill failed")
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func parseQueryInt(r *http.Request, key string, fallback int) int {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}

	return value
}
