package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) DistDFe(ctx context.Context, payload []byte) (int, []byte, error) {
	return c.request(ctx, http.MethodPost, "/api/v1/nfe/distdfe", payload)
}

func (c *Client) DownloadByKey(ctx context.Context, payload []byte) (int, []byte, error) {
	return c.request(ctx, http.MethodPost, "/api/v1/nfe/download", payload)
}

func (c *Client) ConsultaChave(ctx context.Context, payload []byte) (int, []byte, error) {
	return c.request(ctx, http.MethodPost, "/api/v1/nfe/consulta-chave", payload)
}

func (c *Client) Health(ctx context.Context) (int, []byte, error) {
	return c.request(ctx, http.MethodGet, "/health", nil)
}

func (c *Client) request(ctx context.Context, method, path string, payload []byte) (int, []byte, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(payload))
	if err != nil {
		return 0, nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("call sped service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("read sped response: %w", err)
	}

	return resp.StatusCode, body, nil
}
