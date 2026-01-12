// Package download provides functionality to download streams from YouTube.
package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// Progress represents the current download progress.
type Progress struct {
	// Downloaded is the number of bytes downloaded so far.
	Downloaded int64

	// Total is the total size in bytes. May be 0 if unknown.
	Total int64
}

// Percentage returns the download completion percentage (0-100).
// Returns 0 if total size is unknown.
func (p Progress) Percentage() float64 {
	if p.Total == 0 {
		return 0
	}
	return float64(p.Downloaded) / float64(p.Total) * 100
}

// ProgressCallback is a function called to report download progress.
type ProgressCallback func(Progress)

// Downloader handles downloading streams to files.
type Downloader struct {
	client *http.Client
}

// NewDownloader creates a new Downloader with the given HTTP client.
func NewDownloader(client *http.Client) *Downloader {
	if client == nil {
		client = http.DefaultClient
	}
	return &Downloader{client: client}
}

// DownloadStream downloads a stream from the given URL to the specified file path.
// Progress is reported via the optional callback function.
func (d *Downloader) DownloadStream(ctx context.Context, url, filePath string, progress ProgressCallback) error {
	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	// Execute request
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}
	}

	// Create output file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Get content length for progress tracking
	totalSize := resp.ContentLength

	// Create progress-tracking reader if callback is provided
	var reader io.Reader = resp.Body
	if progress != nil {
		reader = &progressReader{
			reader:   resp.Body,
			total:    totalSize,
			callback: progress,
		}
	}

	// Copy data to file
	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	return nil
}

// progressReader wraps an io.Reader to track and report progress.
type progressReader struct {
	reader     io.Reader
	downloaded int64
	total      int64
	callback   ProgressCallback
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.downloaded += int64(n)
		pr.callback(Progress{
			Downloaded: pr.downloaded,
			Total:      pr.total,
		})
	}
	return n, err
}

// StreamDownload represents a single stream to download.
type StreamDownload struct {
	// URL is the stream URL to download from.
	URL string

	// FilePath is the destination file path.
	FilePath string
}

// DownloadResult represents the result of a download operation.
type DownloadResult struct {
	// FilePath is the destination file path.
	FilePath string

	// Error is any error that occurred during download (nil if successful).
	Error error
}

// DownloadStreamsParallel downloads multiple streams in parallel using goroutines.
// Progress is reported as an aggregate of all downloads via the optional callback.
// Returns a slice of DownloadResult in the same order as the input streams.
func (d *Downloader) DownloadStreamsParallel(ctx context.Context, streams []StreamDownload, progress ProgressCallback) []DownloadResult {
	if len(streams) == 0 {
		return nil
	}

	results := make([]DownloadResult, len(streams))
	var wg sync.WaitGroup

	// Create aggregate progress tracker
	var tracker *aggregateProgressTracker
	if progress != nil {
		tracker = newAggregateProgressTracker(len(streams), progress)
	}

	for i, stream := range streams {
		wg.Add(1)
		go func(idx int, s StreamDownload) {
			defer wg.Done()

			var streamProgress ProgressCallback
			if tracker != nil {
				streamProgress = tracker.progressCallbackFor(idx)
			}

			err := d.DownloadStream(ctx, s.URL, s.FilePath, streamProgress)
			results[idx] = DownloadResult{
				FilePath: s.FilePath,
				Error:    err,
			}
		}(i, stream)
	}

	wg.Wait()
	return results
}

// aggregateProgressTracker tracks progress across multiple parallel downloads.
type aggregateProgressTracker struct {
	mu         sync.Mutex
	progresses []Progress // Per-stream progress
	callback   ProgressCallback
}

func newAggregateProgressTracker(count int, callback ProgressCallback) *aggregateProgressTracker {
	return &aggregateProgressTracker{
		progresses: make([]Progress, count),
		callback:   callback,
	}
}

func (apt *aggregateProgressTracker) progressCallbackFor(index int) ProgressCallback {
	return func(p Progress) {
		apt.mu.Lock()
		apt.progresses[index] = p

		// Calculate aggregate progress
		var totalDownloaded, totalSize int64
		for _, sp := range apt.progresses {
			totalDownloaded += sp.Downloaded
			totalSize += sp.Total
		}
		apt.mu.Unlock()

		apt.callback(Progress{
			Downloaded: totalDownloaded,
			Total:      totalSize,
		})
	}
}
