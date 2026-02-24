package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"software.sslmate.com/src/go-pkcs12"
)

type NFSeClient struct {
	baseURL    string
	httpClient *http.Client
}

type NFSeDistResponse struct {
	UltNSU    string        `json:"ultNSU"`
	MaxNSU    string        `json:"maxNSU"`
	CStat     string        `json:"cStat"`
	XMotivo   string        `json:"xMotivo"`
	Documentos []NFSeDocDFe `json:"loteDistDFeInt"`
}

type NFSeDocDFe struct {
	NSU        string `json:"NSU"`
	Schema     string `json:"schema"`
	XMLBase64  string `json:"docZip"`
	ChNFSe     string `json:"chNFSe"`
}

type NFSeConsultaResponse struct {
	XML     string `json:"xml"`
	CStat   string `json:"cStat"`
	XMotivo string `json:"xMotivo"`
}

type NFSeEventosResponse struct {
	Eventos []NFSeEvento `json:"eventos"`
	CStat   string       `json:"cStat"`
	XMotivo string       `json:"xMotivo"`
}

type NFSeEvento struct {
	TpEvento    string `json:"tpEvento"`
	NSU         string `json:"NSU"`
	ChNFSe      string `json:"chNFSe"`
	DHEvento    string `json:"dhEvento"`
	XMLBase64   string `json:"docZip"`
}

func NewNFSeClient(baseURL string, pfxData []byte, senha string) (*NFSeClient, error) {
	privateKey, certificate, caCerts, err := pkcs12.DecodeChain(pfxData, senha)
	if err != nil {
		return nil, fmt.Errorf("decoding PFX certificate: %w", err)
	}

	tlsCert := tls.Certificate{
		Certificate: make([][]byte, 0, 1+len(caCerts)),
		PrivateKey:  privateKey,
	}
	tlsCert.Certificate = append(tlsCert.Certificate, certificate.Raw)
	for _, ca := range caCerts {
		tlsCert.Certificate = append(tlsCert.Certificate, ca.Raw)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS12,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &NFSeClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}, nil
}

func (c *NFSeClient) DistDFe(ctx context.Context, ultNSU string) (*NFSeDistResponse, string, error) {
	url := fmt.Sprintf("%s/DFe/%s", c.baseURL, ultNSU)

	status, body, err := c.get(ctx, url)
	if err != nil {
		return nil, "", err
	}

	rawBody := string(body)

	if status == 404 {
		return &NFSeDistResponse{CStat: "137", XMotivo: "Nenhum documento localizado"}, rawBody, nil
	}

	if status >= 400 {
		return nil, rawBody, fmt.Errorf("ADN API error: status %d, body: %s", status, rawBody)
	}

	if len(body) == 0 {
		return &NFSeDistResponse{CStat: "137", XMotivo: "Resposta vazia do ADN"}, "", nil
	}

	var resp NFSeDistResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, rawBody, fmt.Errorf("parsing ADN response (body=%s): %w", truncateBody(rawBody, 500), err)
	}

	return &resp, rawBody, nil
}

func truncateBody(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func (c *NFSeClient) ConsultaNFSe(ctx context.Context, chaveAcesso string) (*NFSeConsultaResponse, error) {
	url := fmt.Sprintf("%s/NFSe/%s", c.baseURL, chaveAcesso)

	status, body, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}

	if status >= 400 {
		return nil, fmt.Errorf("ADN consulta error: status %d", status)
	}

	var resp NFSeConsultaResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing consulta response: %w", err)
	}

	return &resp, nil
}

func (c *NFSeClient) Eventos(ctx context.Context, chaveAcesso string) (*NFSeEventosResponse, error) {
	url := fmt.Sprintf("%s/NFSe/%s/Eventos", c.baseURL, chaveAcesso)

	status, body, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}

	if status >= 400 {
		return nil, fmt.Errorf("ADN eventos error: status %d", status)
	}

	var resp NFSeEventosResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing eventos response: %w", err)
	}

	return &resp, nil
}

func (c *NFSeClient) get(ctx context.Context, url string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("executing request to ADN: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("reading ADN response: %w", err)
	}

	return resp.StatusCode, body, nil
}
