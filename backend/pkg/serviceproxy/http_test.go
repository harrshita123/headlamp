package serviceproxy_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kubernetes-sigs/headlamp/backend/pkg/serviceproxy"
)

//nolint:funlen
func TestHTTPGet(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		statusCode int
		body       string
		wantErr    bool
		wantStatus int
	}{
		{
			name:       "valid URL",
			url:        "http://example.com",
			statusCode: http.StatusOK,
			body:       "Hello, World!",
			wantErr:    false,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid URL",
			url:        " invalid-url",
			statusCode: 0,
			body:       "",
			wantErr:    true,
		},
		{
			name:       "server returns error response",
			url:        "http://example.com/error",
			statusCode: http.StatusInternalServerError,
			body:       "",
			wantErr:    false,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.statusCode != 0 {
					w.WriteHeader(tt.statusCode)
				}

				if _, err := w.Write([]byte(tt.body)); err != nil {
					t.Fatalf("write test: %v", err)
				}
			}))
			defer ts.Close()

			url := ts.URL
			switch tt.url {
			case " invalid-url":
				url = tt.url
			case "http://example.com/error":
				url = ts.URL + "/error"
			}

			resp, err := serviceproxy.HTTPGet(context.Background(), url)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPGet() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && string(resp.Body) != tt.body {
				t.Errorf("HTTPGet() response = %s, want %s", resp.Body, tt.body)
			}

			if !tt.wantErr && resp.StatusCode != tt.wantStatus {
				t.Errorf("HTTPGet() status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestHTTPGetContextCancellation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte("Hello, World!")); err != nil {
			t.Fatalf("write test: %v", err)
		}
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := serviceproxy.HTTPGet(ctx, ts.URL)
	if err == nil {
		t.Errorf("HTTPGet() error = nil, want error")
	}
}

func TestHTTPGetTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(15 * time.Second)

		if _, err := w.Write([]byte("Hello, World!")); err != nil {
			t.Fatalf("write test: %v", err)
		}
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := serviceproxy.HTTPGet(ctx, ts.URL)
	if err == nil {
		t.Errorf("HTTPGet() error = nil, want error")
	}
}
