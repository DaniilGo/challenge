package source

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

type Source interface {
	GetReadCloser(ctx context.Context) (io.ReadCloser, error)
}

type httpPostSource struct {
	url    string
	body   []byte
	client *http.Client
}

func NewHttpPostSource(url string, body []byte, client *http.Client) Source {
	return &httpPostSource{
		client: client,
		url:    url,
		body:   body,
	}
}

func (c *httpPostSource) GetReadCloser(ctx context.Context) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(c.body))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("httpPostSource.Do(req) error: %w", err)
	}

	return resp.Body, nil
}
