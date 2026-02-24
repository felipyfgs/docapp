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
	StatusProcessamento string       `json:"StatusProcessamento"`
	LoteDFe             []NFSeDocDFe `json:"LoteDFe"`
}

func (r *NFSeDistResponse) HasDocuments() bool {
	return r.StatusProcessamento == "DOCUMENTOS_LOCALIZADOS" && len(r.LoteDFe) > 0
}

type NFSeDocDFe struct {
	NSU            int    `json:"NSU"`
	ChaveAcesso    string `json:"ChaveAcesso"`
	TipoDocumento  string `json:"TipoDocumento"`
	ArquivoXml     string `json:"ArquivoXml"`
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
		return &NFSeDistResponse{StatusProcessamento: "NENHUM_DOCUMENTO_LOCALIZADO"}, rawBody, nil
	}

	if status >= 400 {
		return nil, rawBody, fmt.Errorf("ADN API error: status %d, body: %s", status, rawBody)
	}

	if len(body) == 0 {
		return &NFSeDistResponse{StatusProcessamento: "NENHUM_DOCUMENTO_LOCALIZADO"}, "", nil
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
