// Package http provides HTTP client utilities for YouTube requests.
package http

import (
	"net/http"
	"sync"
	"time"
)

// Version is the application version, can be set via ldflags.
var Version = "dev"

// defaultTimeout is the default timeout for HTTP requests.
const defaultTimeout = 30 * time.Second

var (
	defaultClient     *http.Client
	defaultClientOnce sync.Once
)

// UserAgent returns the User-Agent string for HTTP requests.
func UserAgent() string {
	return "ytdl/" + Version
}

// NewClient creates a new HTTP client with custom settings for YouTube requests.
// The client is configured with:
//   - Custom User-Agent header
//   - Accept-Language header
//   - Reasonable timeout
func NewClient() *http.Client {
	return &http.Client{
		Timeout:   defaultTimeout,
		Transport: &transport{base: http.DefaultTransport},
	}
}

// DefaultClient returns a shared HTTP client instance.
// This is the recommended way to make HTTP requests to YouTube.
func DefaultClient() *http.Client {
	defaultClientOnce.Do(func() {
		defaultClient = NewClient()
	})
	return defaultClient
}

// transport is a custom http.RoundTripper that adds required headers.
type transport struct {
	base http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqCopy := req.Clone(req.Context())

	// Set required headers
	if reqCopy.Header.Get("User-Agent") == "" {
		reqCopy.Header.Set("User-Agent", UserAgent())
	}
	if reqCopy.Header.Get("Accept-Language") == "" {
		reqCopy.Header.Set("Accept-Language", "en-US,en;q=0.9")
	}

	return t.base.RoundTrip(reqCopy)
}
