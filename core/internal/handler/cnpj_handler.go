package handler

import (
	"io"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

var cnpjDigitsOnly = regexp.MustCompile(`^\d{14}$`)

type CNPJHandler struct {
	httpClient *http.Client
	log        zerolog.Logger
}

func NewCNPJHandler(log zerolog.Logger) *CNPJHandler {
	return &CNPJHandler{httpClient: &http.Client{}, log: log}
}

func (h *CNPJHandler) Lookup(w http.ResponseWriter, r *http.Request) {
	cnpj := chi.URLParam(r, "cnpj")

	digits := regexp.MustCompile(`\D`).ReplaceAllString(cnpj, "")
	if !cnpjDigitsOnly.MatchString(digits) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "CNPJ inválido."})
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet,
		"https://publica.cnpj.ws/cnpj/"+digits, nil)
	if err != nil {
		h.log.Error().Err(err).Str("cnpj", digits).Msg("cnpj lookup build request failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Erro ao consultar CNPJ."})
		return
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.log.Error().Err(err).Str("cnpj", digits).Msg("cnpj lookup request failed")
		writeJSON(w, http.StatusBadGateway, map[string]string{"message": "Falha ao consultar API de CNPJ."})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.log.Error().Err(err).Str("cnpj", digits).Msg("cnpj lookup read response failed")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Erro ao ler resposta."})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}
