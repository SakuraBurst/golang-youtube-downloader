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

func TestDownloadStreamsParallel_DownloadsBothStreams(t *testing.T) {
	// Setup test servers for video and audio
	videoContent := []byte("video stream data - fake video content")
	audioContent := []byte("audio stream data - fake audio content")

	videoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(videoContent)))
		_, _ = w.Write(videoContent)
	}))
	defer videoServer.Close()

	audioServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(audioContent)))
		_, _ = w.Write(audioContent)
	}))
	defer audioServer.Close()

	// Create temp files for output
	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "video.mp4")
	audioPath := filepath.Join(tmpDir, "audio.m4a")

	// Download both streams in parallel
	downloader := NewDownloader(http.DefaultClient)
	results := downloader.DownloadStreamsParallel(context.Background(), []StreamDownload{
		{URL: videoServer.URL, FilePath: videoPath},
		{URL: audioServer.URL, FilePath: audioPath},
	}, nil)

	// Verify both downloads succeeded
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Error != nil {
			t.Errorf("Download failed for %s: %v", result.FilePath, result.Error)
		}
	}

	// Verify files were written correctly
	videoData, err := os.ReadFile(videoPath)
	if err != nil {
		t.Fatalf("Failed to read video file: %v", err)
	}
	if !bytes.Equal(videoData, videoContent) {
		t.Errorf("Video content mismatch")
	}

	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		t.Fatalf("Failed to read audio file: %v", err)
	}
	if !bytes.Equal(audioData, audioContent) {
		t.Errorf("Audio content mismatch")
	}
}

func TestDownloadStreamsParallel_HandlesPartialFailure(t *testing.T) {
	// Setup one working server and one failing server
	content := []byte("working stream")
	workingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		_, _ = w.Write(content)
	}))
	defer workingServer.Close()

	failingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer failingServer.Close()

	// Create temp files for output
	tmpDir := t.TempDir()
	workingPath := filepath.Join(tmpDir, "working.mp4")
	failingPath := filepath.Join(tmpDir, "failing.mp4")

	// Download both streams in parallel
	downloader := NewDownloader(http.DefaultClient)
	results := downloader.DownloadStreamsParallel(context.Background(), []StreamDownload{
		{URL: workingServer.URL, FilePath: workingPath},
		{URL: failingServer.URL, FilePath: failingPath},
	}, nil)

	// Verify we got both results
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Find results by path
	var workingResult, failingResult *DownloadResult
	for i := range results {
		switch results[i].FilePath {
		case workingPath:
			workingResult = &results[i]
		case failingPath:
			failingResult = &results[i]
		}
	}

	// Verify working download succeeded
	if workingResult == nil || workingResult.Error != nil {
		t.Errorf("Expected working download to succeed")
	}

	// Verify failing download failed
	if failingResult == nil || failingResult.Error == nil {
		t.Errorf("Expected failing download to fail")
	}
}

func TestDownloadStreamsParallel_ReportsAggregateProgress(t *testing.T) {
	// Setup test servers
	content1 := make([]byte, 500)
	content2 := make([]byte, 500)

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "500")
		_, _ = w.Write(content1)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "500")
		_, _ = w.Write(content2)
	}))
	defer server2.Close()

	// Create temp files for output
	tmpDir := t.TempDir()
	path1 := filepath.Join(tmpDir, "file1.mp4")
	path2 := filepath.Join(tmpDir, "file2.mp4")

	// Track aggregate progress
	var progressUpdates []Progress
	progressCallback := func(p Progress) {
		progressUpdates = append(progressUpdates, p)
	}

	// Download both streams in parallel
	downloader := NewDownloader(http.DefaultClient)
	results := downloader.DownloadStreamsParallel(context.Background(), []StreamDownload{
		{URL: server1.URL, FilePath: path1},
		{URL: server2.URL, FilePath: path2},
	}, progressCallback)

	// Verify downloads succeeded
	for _, result := range results {
		if result.Error != nil {
			t.Errorf("Download failed: %v", result.Error)
		}
	}

	// Verify progress was reported
	if len(progressUpdates) == 0 {
		t.Fatal("Expected progress updates, got none")
	}

	// Verify final progress shows total of both streams (1000 bytes)
	lastProgress := progressUpdates[len(progressUpdates)-1]
	if lastProgress.Total != 1000 {
		t.Errorf("Expected total of 1000 bytes, got %d", lastProgress.Total)
	}
}

func TestDownloadStreamsParallel_HandlesContextCancellation(t *testing.T) {
	// Setup test server that blocks
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000000")
		_, _ = w.Write([]byte("start"))
		<-r.Context().Done()
	}))
	defer server.Close()

	// Create temp files for output
	tmpDir := t.TempDir()
	path1 := filepath.Join(tmpDir, "file1.mp4")
	path2 := filepath.Join(tmpDir, "file2.mp4")

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Download should fail
	downloader := NewDownloader(http.DefaultClient)
	results := downloader.DownloadStreamsParallel(ctx, []StreamDownload{
		{URL: server.URL, FilePath: path1},
		{URL: server.URL, FilePath: path2},
	}, nil)

	// Verify all downloads failed
	for _, result := range results {
		if result.Error == nil {
			t.Errorf("Expected download to fail for %s", result.FilePath)
		}
	}
}
