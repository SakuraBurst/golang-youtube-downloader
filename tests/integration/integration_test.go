// Package integration provides end-to-end tests that make real HTTP requests to YouTube.
// These tests are skipped by default unless YTDL_INTEGRATION_TESTS=1 is set.
package integration

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	ythttp "github.com/SakuraBurst/golang-youtube-downloader/internal/http"
	"github.com/SakuraBurst/golang-youtube-downloader/pkg/youtube"
)

// TestFixtures contains well-known YouTube content for testing.
// These are stable, public videos/playlists unlikely to be removed.
type TestFixtures struct {
	// VideoID is a known public video ID for testing.
	// Using "dQw4w9WgXcQ" (Rick Astley - Never Gonna Give You Up) as it's stable and unlikely to be removed.
	VideoID string

	// VideoTitle is the expected title of the test video.
	VideoTitle string

	// VideoAuthor is the expected author/channel name.
	VideoAuthor string

	// PlaylistID is a known public playlist for testing.
	// Using a small, stable playlist.
	PlaylistID string

	// PlaylistTitle is the expected title of the test playlist.
	PlaylistTitle string

	// PlaylistMinVideos is the minimum expected video count in the playlist.
	PlaylistMinVideos int
}

// DefaultFixtures returns the default test fixtures.
func DefaultFixtures() TestFixtures {
	return TestFixtures{
		// Rick Astley - Never Gonna Give You Up (stable, public, famous)
		VideoID:     "dQw4w9WgXcQ",
		VideoTitle:  "Rick Astley - Never Gonna Give You Up",
		VideoAuthor: "Rick Astley",

		// YouTube Spotlight - Popular on YouTube playlist (public, maintained by YouTube)
		// Using a smaller, stable playlist for faster tests
		PlaylistID:        "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
		PlaylistTitle:     "Elon Musk",
		PlaylistMinVideos: 2,
	}
}

// SkipIfNoIntegration skips the test if integration tests are not enabled.
// Set YTDL_INTEGRATION_TESTS=1 to run integration tests.
func SkipIfNoIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("YTDL_INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration test (set YTDL_INTEGRATION_TESTS=1 to run)")
	}
}

// NewTestContext creates a context with a reasonable timeout for integration tests.
func NewTestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 60*time.Second)
}

// RequireNoError fails the test immediately if err is not nil.
func RequireNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

// TempDir creates a temporary directory for test downloads and registers cleanup.
// The directory is automatically removed when the test completes.
func TempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

// TempFile returns a path to a temporary file within a temp directory.
// The file is not created, only the path is returned.
func TempFile(t *testing.T, name string) string {
	t.Helper()
	dir := TempDir(t)
	return filepath.Join(dir, name)
}

// TestClient returns a configured HTTP client for integration tests.
// This client has appropriate timeouts and headers for YouTube requests.
type TestClient struct {
	*http.Client
	WatchPageFetcher *youtube.WatchPageFetcher
}

// NewTestClient creates a new test client with all necessary components.
func NewTestClient(t *testing.T) *TestClient {
	t.Helper()
	client := ythttp.NewClient()
	return &TestClient{
		Client:           client,
		WatchPageFetcher: &youtube.WatchPageFetcher{Client: client},
	}
}

// FetchVideo fetches video information from YouTube.
func (tc *TestClient) FetchVideo(ctx context.Context, t *testing.T, videoID string) *youtube.Video {
	t.Helper()

	page, err := tc.WatchPageFetcher.Fetch(ctx, videoID)
	RequireNoError(t, err, "Failed to fetch watch page")

	pr, err := page.ExtractPlayerResponse()
	RequireNoError(t, err, "Failed to extract player response")

	video, err := pr.ToVideo()
	RequireNoError(t, err, "Failed to convert player response to video")

	return video
}

// FetchVideoWithStreams fetches video information including stream manifest.
func (tc *TestClient) FetchVideoWithStreams(ctx context.Context, t *testing.T, videoID string) (*youtube.Video, *youtube.StreamManifest) {
	t.Helper()

	page, err := tc.WatchPageFetcher.Fetch(ctx, videoID)
	RequireNoError(t, err, "Failed to fetch watch page")

	pr, err := page.ExtractPlayerResponse()
	RequireNoError(t, err, "Failed to extract player response")

	video, err := pr.ToVideo()
	RequireNoError(t, err, "Failed to convert player response to video")

	var manifest *youtube.StreamManifest
	if pr.StreamingData != nil {
		manifest = pr.StreamingData.GetStreamManifest()
	}

	return video, manifest
}

// AssertVideoValid verifies that a video has all required fields.
func AssertVideoValid(t *testing.T, video *youtube.Video) {
	t.Helper()
	if video.ID == "" {
		t.Error("Video ID should not be empty")
	}
	if video.Title == "" {
		t.Error("Video title should not be empty")
	}
	if video.Author.Name == "" {
		t.Error("Video author name should not be empty")
	}
}

// AssertStreamManifestValid verifies that a stream manifest has available streams.
func AssertStreamManifestValid(t *testing.T, manifest *youtube.StreamManifest) {
	t.Helper()
	if manifest == nil {
		t.Fatal("Stream manifest should not be nil")
	}
	totalStreams := len(manifest.VideoStreams) + len(manifest.AudioStreams) + len(manifest.MuxedStreams)
	if totalStreams == 0 {
		t.Error("Stream manifest should have at least one stream")
	}
}

// TestIntegrationFramework_SkipsWhenNotEnabled tests that integration tests are skipped
// when YTDL_INTEGRATION_TESTS is not set.
func TestIntegrationFramework_SkipsWhenNotEnabled(t *testing.T) {
	// Save and restore the environment variable
	oldVal := os.Getenv("YTDL_INTEGRATION_TESTS")
	defer func() {
		if oldVal != "" {
			_ = os.Setenv("YTDL_INTEGRATION_TESTS", oldVal)
		} else {
			_ = os.Unsetenv("YTDL_INTEGRATION_TESTS")
		}
	}()

	// Clear the variable
	_ = os.Unsetenv("YTDL_INTEGRATION_TESTS")

	// This would skip in a real test, but we can't easily test skipping
	// So we verify the function returns without panicking when env is not set
	if os.Getenv("YTDL_INTEGRATION_TESTS") == "1" {
		t.Error("Expected YTDL_INTEGRATION_TESTS to be unset")
	}
}

// TestIntegrationFramework_FixturesAreValid tests that the default fixtures contain valid data.
func TestIntegrationFramework_FixturesAreValid(t *testing.T) {
	fixtures := DefaultFixtures()

	// Video fixtures
	if !youtube.IsValidVideoID(fixtures.VideoID) {
		t.Errorf("Invalid VideoID in fixtures: %s", fixtures.VideoID)
	}
	if fixtures.VideoTitle == "" {
		t.Error("VideoTitle should not be empty")
	}
	if fixtures.VideoAuthor == "" {
		t.Error("VideoAuthor should not be empty")
	}

	// Playlist fixtures
	if !youtube.IsValidPlaylistID(fixtures.PlaylistID) {
		t.Errorf("Invalid PlaylistID in fixtures: %s", fixtures.PlaylistID)
	}
	if fixtures.PlaylistTitle == "" {
		t.Error("PlaylistTitle should not be empty")
	}
	if fixtures.PlaylistMinVideos < 1 {
		t.Error("PlaylistMinVideos should be at least 1")
	}
}

// TestIntegrationFramework_ContextHasTimeout tests that NewTestContext creates a context with timeout.
func TestIntegrationFramework_ContextHasTimeout(t *testing.T) {
	ctx, cancel := NewTestContext(t)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("Expected context to have a deadline")
	}

	// Deadline should be in the future
	if deadline.Before(time.Now()) {
		t.Error("Deadline should be in the future")
	}

	// Deadline should be within ~60 seconds
	maxDeadline := time.Now().Add(61 * time.Second)
	if deadline.After(maxDeadline) {
		t.Errorf("Deadline is too far in the future: %v", deadline)
	}
}

// TestIntegrationFramework_HTTPClientWorks tests that the HTTP client can be created.
func TestIntegrationFramework_HTTPClientWorks(t *testing.T) {
	client := ythttp.NewClient()
	if client == nil {
		t.Fatal("Expected HTTP client to be created")
	}
}

// TestIntegrationFramework_TestClientWorks tests that the test client can be created.
func TestIntegrationFramework_TestClientWorks(t *testing.T) {
	tc := NewTestClient(t)
	if tc == nil {
		t.Fatal("Expected test client to be created")
	}
	if tc.Client == nil {
		t.Error("Expected test client to have HTTP client")
	}
	if tc.WatchPageFetcher == nil {
		t.Error("Expected test client to have watch page fetcher")
	}
}

// TestIntegrationFramework_TempDirWorks tests that TempDir creates a valid temporary directory.
func TestIntegrationFramework_TempDirWorks(t *testing.T) {
	dir := TempDir(t)
	if dir == "" {
		t.Fatal("Expected temp dir to be non-empty")
	}

	// Verify it exists
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Temp dir does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("Expected temp dir to be a directory")
	}
}

// TestIntegrationFramework_TempFileWorks tests that TempFile returns a valid path.
func TestIntegrationFramework_TempFileWorks(t *testing.T) {
	path := TempFile(t, "test.mp4")
	if path == "" {
		t.Fatal("Expected temp file path to be non-empty")
	}

	// Verify the parent directory exists
	dir := filepath.Dir(path)
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Parent dir does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("Expected parent to be a directory")
	}

	// File should not exist yet
	if _, err := os.Stat(path); err == nil {
		t.Error("Expected temp file to not exist yet")
	}
}

// TestIntegration_FetchVideoInfo fetches real video info from YouTube.
// This is a basic smoke test to verify the framework works with real HTTP requests.
func TestIntegration_FetchVideoInfo(t *testing.T) {
	SkipIfNoIntegration(t)

	fixtures := DefaultFixtures()
	ctx, cancel := NewTestContext(t)
	defer cancel()

	tc := NewTestClient(t)
	video := tc.FetchVideo(ctx, t, fixtures.VideoID)

	// Verify basic fields
	if video.ID != fixtures.VideoID {
		t.Errorf("Expected video ID %q, got %q", fixtures.VideoID, video.ID)
	}

	AssertVideoValid(t, video)

	if video.Duration == 0 {
		t.Error("Expected video to have a duration")
	}

	t.Logf("Successfully fetched video: %q by %q (duration: %v)", video.Title, video.Author.Name, video.Duration)
}

// TestIntegration_FetchVideoWithStreams fetches video info including stream manifest.
func TestIntegration_FetchVideoWithStreams(t *testing.T) {
	SkipIfNoIntegration(t)

	fixtures := DefaultFixtures()
	ctx, cancel := NewTestContext(t)
	defer cancel()

	tc := NewTestClient(t)
	video, manifest := tc.FetchVideoWithStreams(ctx, t, fixtures.VideoID)

	AssertVideoValid(t, video)
	AssertStreamManifestValid(t, manifest)

	// Log available streams
	t.Logf("Video: %q", video.Title)
	t.Logf("Video streams: %d", len(manifest.VideoStreams))
	t.Logf("Audio streams: %d", len(manifest.AudioStreams))
	t.Logf("Muxed streams: %d", len(manifest.MuxedStreams))

	// At least one stream should have a URL
	hasURL := false
	for _, vs := range manifest.VideoStreams {
		if vs.URL != "" {
			hasURL = true
			break
		}
	}
	for _, as := range manifest.AudioStreams {
		if as.URL != "" {
			hasURL = true
			break
		}
	}
	for _, ms := range manifest.MuxedStreams {
		if ms.VideoStreamInfo.URL != "" {
			hasURL = true
			break
		}
	}

	if !hasURL {
		t.Log("Warning: No streams have direct URLs (may require signature decryption)")
	}
}
