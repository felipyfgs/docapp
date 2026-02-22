package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string, timeoutSeconds int) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

type DistDFeRequest struct {
	CertificadoPFX string `json:"certificado_pfx"`
	Senha          string `json:"senha"`
	CNPJ           string `json:"cnpj"`
	RazaoSocial    string `json:"razao_social"`
	SiglaUF        string `json:"sigla_uf"`
	TpAmb          int    `json:"tp_amb"`
	UltNSU         string `json:"ult_nsu"`
}

type DownloadRequest struct {
	CertificadoPFX string `json:"certificado_pfx"`
	Senha          string `json:"senha"`
	CNPJ           string `json:"cnpj"`
	RazaoSocial    string `json:"razao_social"`
	SiglaUF        string `json:"sigla_uf"`
	TpAmb          int    `json:"tp_amb"`
	Chave          string `json:"chave"`
}

type ConsultaChaveRequest struct {
	CertificadoPFX string `json:"certificado_pfx"`
	Senha          string `json:"senha"`
	CNPJ           string `json:"cnpj"`
	RazaoSocial    string `json:"razao_social"`
	SiglaUF        string `json:"sigla_uf"`
	TpAmb          int    `json:"tp_amb"`
	Chave          string `json:"chave"`
}

type SefazResponse struct {
	RawXML     string `json:"raw_xml"`
	CStat      string `json:"cstat,omitempty"`
	XMotivo    string `json:"xmotivo,omitempty"`
	RetryAfter int    `json:"retry_after,omitempty"`
	Error      string `json:"error,omitempty"`
}

func (c *Client) Health(ctx context.Context) (int, []byte, error) {
	return c.get(ctx, "/v1/health")
}

func (c *Client) DistDFe(ctx context.Context, pfx []byte, senha, cnpj, razaoSocial, siglaUF string, tpAmb int, ultNSU string) (*SefazResponse, error) {
	req := DistDFeRequest{
		CertificadoPFX: base64.StdEncoding.EncodeToString(pfx),
		Senha:          senha,
		CNPJ:           cnpj,
		RazaoSocial:    razaoSocial,
		SiglaUF:        siglaUF,
		TpAmb:          tpAmb,
		UltNSU:         ultNSU,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	status, body, err := c.post(ctx, "/v1/sefaz/distdfe", payload)
	if err != nil {
		return nil, err
	}

	var resp SefazResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if status == 429 {
		return &resp, fmt.Errorf("throttled: %s", resp.XMotivo)
	}

	if status >= 400 {
		return &resp, fmt.Errorf("sefaz error: %s", resp.Error)
	}

	return &resp, nil
}

func (c *Client) DownloadByKey(ctx context.Context, pfx []byte, senha, cnpj, razaoSocial, siglaUF string, tpAmb int, chave string) (*SefazResponse, error) {
	req := DownloadRequest{
		CertificadoPFX: base64.StdEncoding.EncodeToString(pfx),
		Senha:          senha,
		CNPJ:           cnpj,
		RazaoSocial:    razaoSocial,
		SiglaUF:        siglaUF,
		TpAmb:          tpAmb,
		Chave:          chave,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	status, body, err := c.post(ctx, "/v1/sefaz/download", payload)
	if err != nil {
		return nil, err
	}

	var resp SefazResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if status >= 400 {
		return &resp, fmt.Errorf("sefaz error: %s", resp.Error)
	}

	return &resp, nil
}

func (c *Client) ConsultaChave(ctx context.Context, pfx []byte, senha, cnpj, razaoSocial, siglaUF string, tpAmb int, chave string) (*SefazResponse, error) {
	req := ConsultaChaveRequest{
		CertificadoPFX: base64.StdEncoding.EncodeToString(pfx),
		Senha:          senha,
		CNPJ:           cnpj,
		RazaoSocial:    razaoSocial,
		SiglaUF:        siglaUF,
		TpAmb:          tpAmb,
		Chave:          chave,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	status, body, err := c.post(ctx, "/v1/sefaz/consulta", payload)
	if err != nil {
		return nil, err
	}

	var resp SefazResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if status >= 400 {
		return &resp, fmt.Errorf("sefaz error: %s", resp.Error)
	}

	return &resp, nil
}

func (c *Client) get(ctx context.Context, path string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("building request: %w", err)
	}

	return c.do(req)
}

func (c *Client) post(ctx context.Context, path string, body []byte) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return 0, nil, fmt.Errorf("building request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return c.do(req)
}

func (c *Client) do(req *http.Request) (int, []byte, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("reading response body: %w", err)
	}

	return resp.StatusCode, body, nil
}
