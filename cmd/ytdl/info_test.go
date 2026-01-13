package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SakuraBurst/golang-youtube-downloader/pkg/youtube"
)

func TestInfoCommandExists(t *testing.T) {
	rootCmd := newRootCmd()
	infoCmd, _, err := rootCmd.Find([]string{"info"})
	if err != nil {
		t.Fatalf("info command not found: %v", err)
	}
	if infoCmd.Use != "info <url>" {
		t.Errorf("expected Use to be 'info <url>', got %q", infoCmd.Use)
	}
}

func TestInfoCommandRequiresURL(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"info"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("info command should fail without URL argument")
	}
}

func TestInfoCommandAcceptsURL(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"info", "https://www.youtube.com/watch?v=dQw4w9WgXcQ"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("info command failed with valid URL: %v", err)
	}
}

func TestInfoCommandHelp(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"info", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("info help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "info") {
		t.Error("help should mention info")
	}
	if !strings.Contains(output, "metadata") || !strings.Contains(output, "video") {
		t.Error("help should mention video metadata")
	}
}

func TestInfoCommandShortDescription(t *testing.T) {
	rootCmd := newRootCmd()
	infoCmd, _, _ := rootCmd.Find([]string{"info"})

	if infoCmd.Short == "" {
		t.Error("info command should have a short description")
	}
}

// TestInfoCommandDisplaysVideoMetadata tests that the info command properly
// fetches and displays video metadata from YouTube.
func TestInfoCommandDisplaysVideoMetadata(t *testing.T) {
	// Create a mock server that returns a valid watch page with player response
	playerResponseJSON := `{
		"videoDetails": {
			"videoId": "dQw4w9WgXcQ",
			"title": "Rick Astley - Never Gonna Give You Up",
			"author": "Rick Astley",
			"lengthSeconds": "212",
			"viewCount": "1000000000",
			"shortDescription": "Official video for Never Gonna Give You Up"
		},
		"playabilityStatus": {
			"status": "OK"
		},
		"streamingData": {
			"formats": [
				{"itag": 18, "mimeType": "video/mp4; codecs=\"avc1.42001E, mp4a.40.2\"", "width": 640, "height": 360, "qualityLabel": "360p"}
			],
			"adaptiveFormats": [
				{"itag": 137, "mimeType": "video/mp4; codecs=\"avc1.640028\"", "width": 1920, "height": 1080, "qualityLabel": "1080p", "bitrate": 4000000},
				{"itag": 140, "mimeType": "audio/mp4; codecs=\"mp4a.40.2\"", "bitrate": 128000, "audioQuality": "AUDIO_QUALITY_MEDIUM"}
			]
		}
	}`

	html := `<!DOCTYPE html>
<html>
<head><title>Test Video</title></head>
<body>
<script>var ytInitialPlayerResponse = ` + playerResponseJSON + `;</script>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	// Create fetcher with test server
	fetcher := &youtube.WatchPageFetcher{
		Client:  server.Client(),
		BaseURL: server.URL,
	}

	// Run info command with the test fetcher
	buf := new(bytes.Buffer)
	err := runInfoWithFetcher(context.Background(), buf, "https://www.youtube.com/watch?v=dQw4w9WgXcQ", fetcher)
	if err != nil {
		t.Fatalf("runInfoWithFetcher failed: %v", err)
	}

	output := buf.String()

	// Verify output contains expected metadata
	expectedStrings := []string{
		"Rick Astley - Never Gonna Give You Up",
		"Rick Astley",
		"3:32", // 212 seconds = 3:32
		"1080p",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("output should contain %q, got:\n%s", expected, output)
		}
	}
}

// TestInfoCommandInvalidVideoID tests error handling for invalid video IDs.
func TestInfoCommandInvalidVideoID(t *testing.T) {
	buf := new(bytes.Buffer)
	fetcher := &youtube.WatchPageFetcher{
		Client: http.DefaultClient,
	}

	err := runInfoWithFetcher(context.Background(), buf, "not-a-valid-url", fetcher)
	if err == nil {
		t.Error("expected error for invalid video ID")
	}
}

// TestInfoCommandVideoUnavailable tests error handling when video is unavailable.
func TestInfoCommandVideoUnavailable(t *testing.T) {
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

	fetcher := &youtube.WatchPageFetcher{
		Client:  server.Client(),
		BaseURL: server.URL,
	}

	buf := new(bytes.Buffer)
	err := runInfoWithFetcher(context.Background(), buf, "dQw4w9WgXcQ", fetcher)
	if err == nil {
		t.Error("expected error for unavailable video")
	}
	if !strings.Contains(err.Error(), "unavailable") {
		t.Errorf("error should mention unavailable, got: %v", err)
	}
}
