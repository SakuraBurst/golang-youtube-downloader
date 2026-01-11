package youtube

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWatchPageURL(t *testing.T) {
	expected := "https://www.youtube.com/watch?v=dQw4w9WgXcQ&bpctr=9999999999"
	got := WatchPageURL("dQw4w9WgXcQ")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestWatchPageURL_DifferentID(t *testing.T) {
	expected := "https://www.youtube.com/watch?v=abc123XYZ90&bpctr=9999999999"
	got := WatchPageURL("abc123XYZ90")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestFetchWatchPage_Success(t *testing.T) {
	// Create a mock server that returns a valid watch page
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path
		if r.URL.Path != "/watch" {
			t.Errorf("expected path /watch, got %s", r.URL.Path)
		}
		// Verify video ID parameter
		if r.URL.Query().Get("v") != "dQw4w9WgXcQ" {
			t.Errorf("expected v=dQw4w9WgXcQ, got v=%s", r.URL.Query().Get("v"))
		}
		// Verify bpctr parameter
		if r.URL.Query().Get("bpctr") != "9999999999" {
			t.Errorf("expected bpctr=9999999999, got bpctr=%s", r.URL.Query().Get("bpctr"))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Test Video</title></head><body></body></html>`))
	}))
	defer server.Close()

	fetcher := &WatchPageFetcher{
		Client:  server.Client(),
		BaseURL: server.URL,
	}

	page, err := fetcher.Fetch(context.Background(), "dQw4w9WgXcQ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page == nil {
		t.Fatal("expected page to be non-nil")
	}
	if page.HTML == "" {
		t.Error("expected HTML to be non-empty")
	}
}

func TestFetchWatchPage_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	fetcher := &WatchPageFetcher{
		Client:  server.Client(),
		BaseURL: server.URL,
	}

	_, err := fetcher.Fetch(context.Background(), "invalidID123")
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestFetchWatchPage_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	fetcher := &WatchPageFetcher{
		Client:  server.Client(),
		BaseURL: server.URL,
	}

	_, err := fetcher.Fetch(context.Background(), "dQw4w9WgXcQ")
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestFetchWatchPage_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	fetcher := &WatchPageFetcher{
		Client:  server.Client(),
		BaseURL: server.URL,
	}

	_, err := fetcher.Fetch(context.Background(), "dQw4w9WgXcQ")
	if err == nil {
		t.Error("expected error for 429 response")
	}
	if _, ok := err.(*RateLimitError); !ok {
		t.Errorf("expected RateLimitError, got %T", err)
	}
}

func TestFetchWatchPage_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should not be reached
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	fetcher := &WatchPageFetcher{
		Client:  server.Client(),
		BaseURL: server.URL,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := fetcher.Fetch(ctx, "dQw4w9WgXcQ")
	if err == nil {
		t.Error("expected error for canceled context")
	}
}

func TestWatchPage_HasHTML(t *testing.T) {
	page := &WatchPage{
		VideoID: "dQw4w9WgXcQ",
		HTML:    "<html></html>",
	}

	if page.VideoID != "dQw4w9WgXcQ" {
		t.Error("VideoID should be set")
	}
	if page.HTML != "<html></html>" {
		t.Error("HTML should be set")
	}
}

func TestRateLimitError_Error(t *testing.T) {
	err := &RateLimitError{Message: "too many requests"}
	expected := "rate limit exceeded: too many requests"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestVideoUnavailableError_Error(t *testing.T) {
	err := &VideoUnavailableError{VideoID: "dQw4w9WgXcQ", Reason: "private"}
	expected := "video 'dQw4w9WgXcQ' is unavailable: private"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}
