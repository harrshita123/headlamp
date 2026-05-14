package serviceproxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kubernetes-sigs/headlamp/backend/pkg/logger"
)

// UpstreamResponse contains the upstream service response data forwarded by the proxy.
type UpstreamResponse struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

// HTTPGet sends an HTTP GET request to the specified URI.
func HTTPGet(ctx context.Context, uri string) (*UpstreamResponse, error) {
	cli := &http.Client{Timeout: 10 * time.Second}

	logger.Log(logger.LevelInfo, nil, nil, fmt.Sprintf("make request to %s", uri))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	resp, err := cli.Do(req) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed HTTP GET: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &UpstreamResponse{
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Body:        body,
	}, nil
}
