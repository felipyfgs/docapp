package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"docapp/core/internal/repository"
	"docapp/core/internal/service"

	"github.com/rs/zerolog"
)

type DocumentoHandler struct {
	svc           *service.DocumentoService
	importService *service.ImportService
	syncService   *service.SyncService
	docRepo       *repository.DocumentoRepository
	log           zerolog.Logger
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

func NewDocumentoHandler(svc *service.DocumentoService, importService *service.ImportService, syncService *service.SyncService, docRepo *repository.DocumentoRepository, log zerolog.Logger) *DocumentoHandler {
	return &DocumentoHandler{svc: svc, importService: importService, syncService: syncService, docRepo: docRepo, log: log}
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

func (h *DocumentoHandler) Itens(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "ID inválido."})
		return
	}

	itens, err := h.svc.ListItens(r.Context(), id)
	if err != nil {
		h.log.Error().Err(err).Uint("id", id).Msg("list itens failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Erro ao listar itens."})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": itens,
		"total": len(itens),
	})
}

func (h *DocumentoHandler) BackfillItens(w http.ResponseWriter, r *http.Request) {
	var req backfillRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Payload inválido."})
			return
		}
	}

	result, err := h.svc.BackfillItens(r.Context(), req.Limit)
	if err != nil {
		h.log.Error().Err(err).Msg("backfill itens failed")
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
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

func (h *DocumentoHandler) Import(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(200 << 20); err != nil {
		writeJSONDoc(w, http.StatusBadRequest, map[string]string{"message": "Falha ao ler form-data."})
		return
	}

	fhs := r.MultipartForm.File["files"]
	if len(fhs) == 0 {
		writeJSONDoc(w, http.StatusBadRequest, map[string]string{"message": "Campo 'files' não encontrado."})
		return
	}

	var allFiles []service.ImportFile
	for _, fh := range fhs {
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if ext != ".xml" && ext != ".zip" {
			continue
		}
		f, err := fh.Open()
		if err != nil {
			continue
		}
		content, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			continue
		}
		extracted, err := service.ExtractFiles(fh.Filename, content)
		if err != nil {
			h.log.Warn().Err(err).Str("filename", fh.Filename).Msg("import: extract failed")
			continue
		}
		allFiles = append(allFiles, extracted...)
	}

	if len(allFiles) == 0 {
		writeJSONDoc(w, http.StatusBadRequest, map[string]string{"message": "Nenhum arquivo XML encontrado."})
		return
	}

	result := h.importService.ImportDocumentosAuto(r.Context(), allFiles)
	writeJSONDoc(w, http.StatusOK, result)
}

func (h *DocumentoHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	fromStr := strings.TrimSpace(r.URL.Query().Get("from"))
	toStr := strings.TrimSpace(r.URL.Query().Get("to"))
	groupBy := strings.TrimSpace(r.URL.Query().Get("group_by"))

	if fromStr == "" || toStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Parâmetros 'from' e 'to' são obrigatórios (YYYY-MM-DD)."})
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Formato de data 'from' inválido. Use YYYY-MM-DD."})
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Formato de data 'to' inválido. Use YYYY-MM-DD."})
		return
	}

	to = to.Add(24*time.Hour - time.Second)

	if groupBy == "" {
		groupBy = "daily"
	}

	ctx := r.Context()

	stats, err := h.docRepo.DashboardStats(ctx, from, to)
	if err != nil {
		h.log.Error().Err(err).Msg("dashboard stats failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Erro ao calcular estatísticas."})
		return
	}

	duration := to.Sub(from)
	prevTo := from.Add(-time.Second)
	prevFrom := prevTo.Add(-duration)
	prevStats, err := h.docRepo.DashboardPreviousPeriodStats(ctx, prevFrom, prevTo)
	if err != nil {
		h.log.Error().Err(err).Msg("dashboard previous period stats failed")
		prevStats = &repository.DashboardPeriodStats{}
	}

	chart, err := h.docRepo.DashboardChart(ctx, from, to, groupBy)
	if err != nil {
		h.log.Error().Err(err).Msg("dashboard chart failed")
		chart = nil
	}

	recentes, err := h.docRepo.DashboardRecentes(ctx, 10)
	if err != nil {
		h.log.Error().Err(err).Msg("dashboard recentes failed")
		recentes = nil
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stats":          stats,
		"previous_stats": prevStats,
		"chart":          chart,
		"recentes":       recentes,
	})
}

type manifestarRequest struct {
	IDs           []uint `json:"ids"`
	TipoEvento    string `json:"tipo_evento"`
	Justificativa string `json:"justificativa"`
}

func (h *DocumentoHandler) Manifestar(w http.ResponseWriter, r *http.Request) {
	var req manifestarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Payload inválido."})
		return
	}

	if len(req.IDs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Nenhum documento selecionado."})
		return
	}

	validEvents := map[string]bool{"210210": true, "210200": true, "210220": true, "210240": true}
	if !validEvents[req.TipoEvento] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Tipo de evento inválido."})
		return
	}

	if req.TipoEvento == "210240" && strings.TrimSpace(req.Justificativa) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Justificativa é obrigatória para 'Operação Não Realizada'."})
		return
	}

	result, err := h.syncService.ManifestarEmLote(r.Context(), service.ManifestacaoRequest{
		IDs:           req.IDs,
		TipoEvento:    req.TipoEvento,
		Justificativa: req.Justificativa,
	})
	if err != nil {
		h.log.Error().Err(err).Msg("manifestar em lote failed")
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func writeJSONDoc(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
