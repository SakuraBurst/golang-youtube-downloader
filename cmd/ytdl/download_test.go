package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SakuraBurst/golang-youtube-downloader/pkg/download"
	"github.com/SakuraBurst/golang-youtube-downloader/pkg/youtube"
)

func TestDownloadCommandExists(t *testing.T) {
	rootCmd := newRootCmd()
	downloadCmd, _, err := rootCmd.Find([]string{"download"})
	if err != nil {
		t.Fatalf("download command not found: %v", err)
	}
	if downloadCmd.Use != "download <url>" {
		t.Errorf("expected Use to be 'download <url>', got %q", downloadCmd.Use)
	}
}

func TestDownloadCommandRequiresURL(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"download"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("download command should fail without URL argument")
	}
}

func TestDownloadCommandAcceptsURL(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"download", "https://www.youtube.com/watch?v=dQw4w9WgXcQ"})

	// Command will fail at runtime (network error or no stream),
	// but should not fail at argument parsing stage
	err := rootCmd.Execute()
	// We expect an error because no mock server is set up,
	// but the error should be from download logic, not argument parsing
	if err != nil {
		// Verify it's a download error, not an argument error
		if strings.Contains(err.Error(), "accepts 1 arg") || strings.Contains(err.Error(), "requires") {
			t.Errorf("unexpected argument parsing error: %v", err)
		}
		// Download errors are expected (no network/mock) - test passes
	}
}

func TestDownloadCommandHasOutputFlag(t *testing.T) {
	rootCmd := newRootCmd()
	downloadCmd, _, _ := rootCmd.Find([]string{"download"})

	flag := downloadCmd.Flags().Lookup("output")
	if flag == nil {
		t.Error("download command should have --output flag")
	}
}

func TestDownloadCommandHasQualityFlag(t *testing.T) {
	rootCmd := newRootCmd()
	downloadCmd, _, _ := rootCmd.Find([]string{"download"})

	flag := downloadCmd.Flags().Lookup("quality")
	if flag == nil {
		t.Error("download command should have --quality flag")
	}
}

func TestDownloadCommandHasFormatFlag(t *testing.T) {
	rootCmd := newRootCmd()
	downloadCmd, _, _ := rootCmd.Find([]string{"download"})

	flag := downloadCmd.Flags().Lookup("format")
	if flag == nil {
		t.Error("download command should have --format flag")
	}
}

func TestDownloadCommandHelp(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"download", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("download help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "download") {
		t.Error("help should mention download")
	}
	if !strings.Contains(output, "output") {
		t.Error("help should mention output flag")
	}
}

// TestDownloadCommandInvalidVideoID tests error handling for invalid video IDs.
func TestDownloadCommandInvalidVideoID(t *testing.T) {
	opts := &downloadOptions{
		output:  t.TempDir(),
		quality: "best",
		format:  "mp4",
	}

	fetcher := &youtube.WatchPageFetcher{
		Client: http.DefaultClient,
	}
	downloader := download.NewDownloader(http.DefaultClient)

	buf := new(bytes.Buffer)
	err := runDownloadWithDeps(context.Background(), buf, "not-a-valid-url", opts, fetcher, downloader, nil)
	if err == nil {
		t.Error("expected error for invalid video ID")
	}
}

// TestDownloadCommandVideoUnavailable tests error handling when video is unavailable.
func TestDownloadCommandVideoUnavailable(t *testing.T) {
	playerResponseJSON := `{
		"videoDetails": {
			"videoId": "dQw4w9WgXcQ",
			"title": ""
		},
		"playabilityStatus": {
			"status": "ERROR",
			"reason": "Video unavailable"
		}
	}`

	html := `<!DOCTYPE html>
<script>var ytInitialPlayerResponse = ` + playerResponseJSON + `;</script>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	opts := &downloadOptions{
		output:  t.TempDir(),
		quality: "best",
		format:  "mp4",
	}

	fetcher := &youtube.WatchPageFetcher{
		Client:  server.Client(),
		BaseURL: server.URL,
	}
	downloader := download.NewDownloader(server.Client())

	buf := new(bytes.Buffer)
	err := runDownloadWithDeps(context.Background(), buf, "dQw4w9WgXcQ", opts, fetcher, downloader, nil)
	if err == nil {
		t.Error("expected error for unavailable video")
	}
	if !strings.Contains(err.Error(), "unavailable") {
		t.Errorf("error should mention unavailable, got: %v", err)
	}
}

// TestDownloadCommandWithMuxedStream tests downloading a muxed stream (video+audio combined).
func TestDownloadCommandWithMuxedStream(t *testing.T) {
	// Create player response with muxed stream
	playerResponseJSON := `{
		"videoDetails": {
			"videoId": "dQw4w9WgXcQ",
			"title": "Test Video",
			"author": "Test Channel",
			"lengthSeconds": "120",
			"viewCount": "1000"
		},
		"playabilityStatus": {
			"status": "OK"
		},
		"streamingData": {
			"formats": [
				{"itag": 18, "url": "STREAM_URL", "mimeType": "video/mp4; codecs=\"avc1.42001E, mp4a.40.2\"", "width": 640, "height": 360, "qualityLabel": "360p", "contentLength": "100"}
			]
		}
	}`

	html := `<!DOCTYPE html>
<script>var ytInitialPlayerResponse = ` + playerResponseJSON + `;</script>`

	// Create a mock stream content
	streamContent := []byte("fake video content for testing")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/watch" {
			// Return watch page
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(html))
		} else {
			// Return stream content
			w.Header().Set("Content-Length", "30")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(streamContent)
		}
	}))
	defer server.Close()

	// Replace STREAM_URL in JSON with actual server URL
	html = strings.ReplaceAll(html, "STREAM_URL", server.URL+"/stream")

	// Re-create server with updated HTML
	server.Close()
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/watch" {
			updatedHTML := strings.ReplaceAll(`<!DOCTYPE html>
<script>var ytInitialPlayerResponse = `+playerResponseJSON+`;</script>`, "STREAM_URL", server.URL+"/stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(updatedHTML))
		} else {
			w.Header().Set("Content-Length", "30")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(streamContent)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	opts := &downloadOptions{
		output:  tempDir,
		quality: "best",
		format:  "mp4",
	}

	fetcher := &youtube.WatchPageFetcher{
		Client:  server.Client(),
		BaseURL: server.URL,
	}
	downloader := download.NewDownloader(server.Client())

	buf := new(bytes.Buffer)
	err := runDownloadWithDeps(context.Background(), buf, "dQw4w9WgXcQ", opts, fetcher, downloader, nil)
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	// Verify output file was created
	outputFile := filepath.Join(tempDir, "Test Video.mp4")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("expected output file to exist: %s", outputFile)
	}
}

// TestDownloadCommandQualityParsing tests quality preference parsing.
func TestDownloadQualityParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected youtube.VideoQualityPreference
	}{
		{"best", youtube.QualityHighest},
		{"1080p", youtube.QualityUpTo1080p},
		{"720p", youtube.QualityUpTo720p},
		{"480p", youtube.QualityUpTo480p},
		{"360p", youtube.QualityUpTo360p},
		{"worst", youtube.QualityLowest},
		{"audio", youtube.QualityLowest}, // audio-only defaults to lowest video quality (will be handled separately)
	}

	for _, tt := range tests {
		got := parseQualityPreference(tt.input)
		if got != tt.expected {
			t.Errorf("parseQualityPreference(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
