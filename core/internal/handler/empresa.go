package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"docapp/core/internal/model"
	"docapp/core/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type EmpresaHandler struct {
	svc         *service.EmpresaService
	syncService *service.SyncService
	log         zerolog.Logger
}

func NewEmpresaHandler(svc *service.EmpresaService, syncService *service.SyncService, log zerolog.Logger) *EmpresaHandler {
	return &EmpresaHandler{svc: svc, syncService: syncService, log: log}
}

func (h *EmpresaHandler) List(w http.ResponseWriter, r *http.Request) {
	empresas, err := h.svc.List()
	if err != nil {
		h.log.Error().Err(err).Msg("list empresas failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Erro ao listar empresas."})
		return
	}

	writeJSON(w, http.StatusOK, empresas)
}

func (h *EmpresaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var empresa model.Empresa
	if err := json.NewDecoder(r.Body).Decode(&empresa); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Payload inválido."})
		return
	}

	if empresa.CNPJ == "" || empresa.RazaoSocial == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "CNPJ e Razão Social são obrigatórios."})
		return
	}

	if empresa.LookbackDays <= 0 {
		empresa.LookbackDays = 90
	}

	if err := h.svc.Create(&empresa); err != nil {
		if strings.Contains(err.Error(), "idx_empresas_cnpj") {
			writeJSON(w, http.StatusConflict, map[string]string{"message": "CNPJ já cadastrado."})
			return
		}
		h.log.Error().Err(err).Str("cnpj", empresa.CNPJ).Msg("create empresa failed")
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": "Erro ao criar empresa."})
		return
	}

	writeJSON(w, http.StatusCreated, empresa)
}

func (h *EmpresaHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "ID inválido."})
		return
	}

	empresa, err := h.svc.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "Empresa não encontrada."})
			return
		}

		h.log.Error().Err(err).Uint("id", id).Msg("get empresa failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Erro ao buscar empresa."})
		return
	}

	writeJSON(w, http.StatusOK, empresa)
}

func (h *EmpresaHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "ID inválido."})
		return
	}

	var updates model.Empresa
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Payload inválido."})
		return
	}

	empresa, err := h.svc.Update(id, &updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "Empresa não encontrada."})
			return
		}

		h.log.Error().Err(err).Uint("id", id).Msg("update empresa failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Erro ao atualizar empresa."})
		return
	}

	writeJSON(w, http.StatusOK, empresa)
}

func (h *EmpresaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "ID inválido."})
		return
	}

	if err := h.svc.Delete(id); err != nil {
		h.log.Error().Err(err).Uint("id", id).Msg("delete empresa failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Erro ao deletar empresa."})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *EmpresaHandler) UploadCertificado(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "ID inválido."})
		return
	}

	_, err = h.svc.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "Empresa não encontrada."})
			return
		}
		h.log.Error().Err(err).Uint("id", id).Msg("get empresa for cert failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Erro ao buscar empresa."})
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Formulário multipart inválido."})
		return
	}

	file, _, err := r.FormFile("certificado")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Campo 'certificado' não encontrado."})
		return
	}
	defer file.Close()

	senha := r.FormValue("senha")
	siglaUF := r.FormValue("sigla_uf")
	tpAmb := 1
	if r.FormValue("tp_amb") == "2" {
		tpAmb = 2
	}

	pfxData, err := io.ReadAll(file)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to read cert file")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Erro ao ler certificado."})
		return
	}

	if err := h.svc.UpdateCertificadoPFX(id, pfxData, senha, siglaUF, tpAmb); err != nil {
		h.log.Error().Err(err).Uint("id", id).Msg("update certificado failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Erro ao atualizar certificado na empresa."})
		return
	}

	updated, _ := h.svc.GetByID(id)
	writeJSON(w, http.StatusOK, updated)
}

func (h *EmpresaHandler) Sync(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "ID inválido."})
		return
	}

	empresa, err := h.svc.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "Empresa não encontrada."})
			return
		}
		h.log.Error().Err(err).Uint("id", id).Msg("get empresa for sync failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Erro ao buscar empresa."})
		return
	}

	if len(empresa.CertificadoPFX) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Empresa sem certificado configurado."})
		return
	}

	if err := h.syncService.SyncEmpresa(*empresa); err != nil {
		h.log.Error().Err(err).Uint("id", id).Msg("sync failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"message": "Erro ao sincronizar empresa.",
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Sincronização concluída."})
}

func parseID(r *http.Request) (uint, error) {
	raw := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
