package client

import (
	"bytes"
	"context"
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

func (c *Client) Health(ctx context.Context) (int, []byte, error) {
	return c.get(ctx, "/v1/nfe/health")
}

func (c *Client) DistDFe(ctx context.Context, payload []byte) (int, []byte, error) {
	return c.post(ctx, "/v1/nfe/distdfe", payload)
}

func (c *Client) DownloadByKey(ctx context.Context, payload []byte) (int, []byte, error) {
	return c.post(ctx, "/v1/nfe/download", payload)
}

func (c *Client) ConsultaChave(ctx context.Context, payload []byte) (int, []byte, error) {
	return c.post(ctx, "/v1/nfe/consulta-chave", payload)
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
