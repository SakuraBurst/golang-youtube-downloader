package download

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadStream_WritesToFile(t *testing.T) {
	// Setup test server that returns some content
	content := []byte("test video content - this is fake stream data")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		_, _ = w.Write(content)
	}))
	defer server.Close()

	// Create temp file for output
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.mp4")

	// Download the stream
	downloader := NewDownloader(http.DefaultClient)
	err := downloader.DownloadStream(context.Background(), server.URL, outputPath, nil)
	if err != nil {
		t.Fatalf("DownloadStream failed: %v", err)
	}

	// Verify file was written
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !bytes.Equal(data, content) {
		t.Errorf("Content mismatch: got %q, want %q", data, content)
	}
}

func TestDownloadStream_ReportsProgress(t *testing.T) {
	// Setup test server with known content size
	content := make([]byte, 1000) // 1KB of data
	for i := range content {
		content[i] = byte(i % 256)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		_, _ = w.Write(content)
	}))
	defer server.Close()

	// Create temp file for output
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.mp4")

	// Track progress callbacks
	var progressUpdates []Progress
	progressCallback := func(p Progress) {
		progressUpdates = append(progressUpdates, p)
	}

	// Download the stream
	downloader := NewDownloader(http.DefaultClient)
	err := downloader.DownloadStream(context.Background(), server.URL, outputPath, progressCallback)
	if err != nil {
		t.Fatalf("DownloadStream failed: %v", err)
	}

	// Verify progress was reported
	if len(progressUpdates) == 0 {
		t.Fatal("Expected progress updates, got none")
	}

	// Verify final progress shows completion
	lastProgress := progressUpdates[len(progressUpdates)-1]
	if lastProgress.Downloaded != lastProgress.Total {
		t.Errorf("Final progress incomplete: downloaded %d of %d", lastProgress.Downloaded, lastProgress.Total)
	}
}

func TestDownloadStream_HandlesContextCancellation(t *testing.T) {
	// Setup test server that writes slowly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000000")
		// Write some data then stall
		_, _ = w.Write([]byte("start"))
		// The client should cancel before we get here
		<-r.Context().Done()
	}))
	defer server.Close()

	// Create temp file for output
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.mp4")

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Download should fail due to canceled context
	downloader := NewDownloader(http.DefaultClient)
	err := downloader.DownloadStream(ctx, server.URL, outputPath, nil)
	if err == nil {
		t.Fatal("Expected error for canceled context, got nil")
	}
}

func TestDownloadStream_CreatesParentDirectory(t *testing.T) {
	// Setup test server
	content := []byte("test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "12")
		_, _ = w.Write(content)
	}))
	defer server.Close()

	// Create path with non-existent parent directory
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "subdir", "nested", "output.mp4")

	// Download the stream
	downloader := NewDownloader(http.DefaultClient)
	err := downloader.DownloadStream(context.Background(), server.URL, outputPath, nil)
	if err != nil {
		t.Fatalf("DownloadStream failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Expected output file to be created")
	}
}

func TestDownloadStream_HandlesHTTPError(t *testing.T) {
	// Setup test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer server.Close()

	// Create temp file for output
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.mp4")

	// Download should fail
	downloader := NewDownloader(http.DefaultClient)
	err := downloader.DownloadStream(context.Background(), server.URL, outputPath, nil)
	if err == nil {
		t.Fatal("Expected error for HTTP 404, got nil")
	}
}

func TestProgress_Percentage(t *testing.T) {
	tests := []struct {
		name     string
		progress Progress
		wantPct  float64
	}{
		{
			name:     "zero total",
			progress: Progress{Downloaded: 100, Total: 0},
			wantPct:  0,
		},
		{
			name:     "half done",
			progress: Progress{Downloaded: 50, Total: 100},
			wantPct:  50,
		},
		{
			name:     "complete",
			progress: Progress{Downloaded: 100, Total: 100},
			wantPct:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.progress.Percentage()
			if got != tt.wantPct {
				t.Errorf("Percentage() = %v, want %v", got, tt.wantPct)
			}
		})
	}
}
