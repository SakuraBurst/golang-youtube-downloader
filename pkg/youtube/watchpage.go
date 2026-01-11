package youtube

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

const (
	// youtubeBaseURL is the base URL for YouTube.
	youtubeBaseURL = "https://www.youtube.com"

	// bpctr parameter value to bypass content restriction checks.
	bpctrValue = "9999999999"
)

// WatchPage represents a fetched YouTube video watch page.
type WatchPage struct {
	// VideoID is the video ID this page was fetched for.
	VideoID string

	// HTML is the raw HTML content of the page.
	HTML string
}

// WatchPageFetcher fetches YouTube video watch pages.
type WatchPageFetcher struct {
	// Client is the HTTP client to use for requests.
	Client *http.Client

	// BaseURL is the base URL for YouTube (used for testing).
	// If empty, defaults to https://www.youtube.com.
	BaseURL string
}

// WatchPageURL returns the URL for a video's watch page.
// The bpctr parameter is included to bypass content restriction checks.
func WatchPageURL(videoID string) string {
	return fmt.Sprintf("%s/watch?v=%s&bpctr=%s", youtubeBaseURL, videoID, bpctrValue)
}

// Fetch retrieves the watch page HTML for a given video ID.
func (f *WatchPageFetcher) Fetch(ctx context.Context, videoID string) (*WatchPage, error) {
	baseURL := f.BaseURL
	if baseURL == "" {
		baseURL = youtubeBaseURL
	}

	url := fmt.Sprintf("%s/watch?v=%s&bpctr=%s", baseURL, videoID, bpctrValue)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching watch page: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &RateLimitError{Message: "YouTube returned 429 Too Many Requests"}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return &WatchPage{
		VideoID: videoID,
		HTML:    string(body),
	}, nil
}

// RateLimitError is returned when YouTube rate limits the request.
type RateLimitError struct {
	Message string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %s", e.Message)
}

// VideoUnavailableError is returned when a video is not available.
type VideoUnavailableError struct {
	VideoID string
	Reason  string
}

func (e *VideoUnavailableError) Error() string {
	return fmt.Sprintf("video '%s' is unavailable: %s", e.VideoID, e.Reason)
}
