package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient_ReturnsNonNil(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient should return a non-nil client")
	}
}

func TestNewClient_SetsUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent := r.Header.Get("User-Agent")
		if !strings.Contains(userAgent, "ytdl") {
			t.Errorf("User-Agent should contain 'ytdl', got: %s", userAgent)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestNewClient_SetsAcceptLanguage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptLang := r.Header.Get("Accept-Language")
		if acceptLang == "" {
			t.Error("Accept-Language header should be set")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestDefaultClient_IsSingleton(t *testing.T) {
	client1 := DefaultClient()
	client2 := DefaultClient()
	if client1 != client2 {
		t.Error("DefaultClient should return the same instance")
	}
}

func TestClient_HasTimeout(t *testing.T) {
	client := NewClient()
	if client.Timeout == 0 {
		t.Error("client should have a timeout set")
	}
}

func TestUserAgent_ContainsVersion(t *testing.T) {
	ua := UserAgent()
	if !strings.Contains(ua, "ytdl/") {
		t.Errorf("UserAgent should contain 'ytdl/', got: %s", ua)
	}
}
