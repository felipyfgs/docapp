package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"docapp/core/internal/model"
	"docapp/core/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type EmpresaHandler struct {
	svc *service.EmpresaService
	log zerolog.Logger
}

func NewEmpresaHandler(svc *service.EmpresaService, log zerolog.Logger) *EmpresaHandler {
	return &EmpresaHandler{svc: svc, log: log}
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
