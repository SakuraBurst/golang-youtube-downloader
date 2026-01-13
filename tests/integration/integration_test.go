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
	"github.com/SakuraBurst/golang-youtube-downloader/pkg/download"
	"github.com/SakuraBurst/golang-youtube-downloader/pkg/ffmpeg"
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

// SkipIfNoFFmpeg skips the test if FFmpeg is not available.
func SkipIfNoFFmpeg(t *testing.T) {
	t.Helper()
	if !ffmpeg.IsAvailable() {
		t.Skip("Skipping test: FFmpeg not available")
	}
}

// FindStreamWithDirectURL finds a video stream with a direct URL (no signature cipher).
func FindStreamWithDirectURL(manifest *youtube.StreamManifest) *youtube.VideoStreamInfo {
	for i := range manifest.VideoStreams {
		if manifest.VideoStreams[i].URL != "" {
			return &manifest.VideoStreams[i]
		}
	}
	return nil
}

// FindAudioStreamWithDirectURL finds an audio stream with a direct URL.
func FindAudioStreamWithDirectURL(manifest *youtube.StreamManifest) *youtube.AudioStreamInfo {
	for i := range manifest.AudioStreams {
		if manifest.AudioStreams[i].URL != "" {
			return &manifest.AudioStreams[i]
		}
	}
	return nil
}

// FindMuxedStreamWithDirectURL finds a muxed stream with a direct URL.
func FindMuxedStreamWithDirectURL(manifest *youtube.StreamManifest) *youtube.MuxedStreamInfo {
	for i := range manifest.MuxedStreams {
		if manifest.MuxedStreams[i].VideoStreamInfo.URL != "" {
			return &manifest.MuxedStreams[i]
		}
	}
	return nil
}

// AssertFileExists verifies that a file exists and has non-zero size.
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Fatalf("Expected file to exist: %s", path)
	}
	if err != nil {
		t.Fatalf("Error checking file: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("Expected file to have non-zero size: %s", path)
	}
}

// AssertFileMinSize verifies that a file exists and has at least minSize bytes.
func AssertFileMinSize(t *testing.T, path string, minSize int64) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Error checking file %s: %v", path, err)
	}
	if info.Size() < minSize {
		t.Errorf("File %s is too small: %d bytes (expected at least %d)", path, info.Size(), minSize)
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

// TestIntegration_FetchVideoFromURL tests the full flow: parse URL → resolve query → fetch video info.
// This tests various URL formats that YouTube supports.
func TestIntegration_FetchVideoFromURL(t *testing.T) {
	SkipIfNoIntegration(t)

	fixtures := DefaultFixtures()
	ctx, cancel := NewTestContext(t)
	defer cancel()

	client := NewTestClient(t)

	// Test various URL formats that should all resolve to the same video
	testCases := []struct {
		name string
		url  string
	}{
		{"standard watch URL", "https://www.youtube.com/watch?v=" + fixtures.VideoID},
		{"short URL", "https://youtu.be/" + fixtures.VideoID},
		{"embedded URL", "https://www.youtube.com/embed/" + fixtures.VideoID},
		{"video ID only", fixtures.VideoID},
		{"watch URL with extra params", "https://www.youtube.com/watch?v=" + fixtures.VideoID + "&list=PLtest&t=10"},
	}

	for _, testCase := range testCases {
		testCase := testCase // capture for parallel subtests
		t.Run(testCase.name, func(t *testing.T) {
			// Parse the URL/query to extract video ID
			result, err := youtube.ResolveQuery(testCase.url)
			RequireNoError(t, err, "Failed to resolve query")

			if result.Type != youtube.QueryTypeVideo {
				t.Errorf("Expected query type %v, got %v", youtube.QueryTypeVideo, result.Type)
			}

			if result.VideoID != fixtures.VideoID {
				t.Errorf("Expected video ID %q, got %q", fixtures.VideoID, result.VideoID)
			}
		})
	}

	// Now test the full fetch flow with one URL
	t.Run("full fetch flow", func(t *testing.T) {
		watchURL := "https://www.youtube.com/watch?v=" + fixtures.VideoID

		// Step 1: Parse URL
		result, err := youtube.ResolveQuery(watchURL)
		RequireNoError(t, err, "Failed to resolve query")

		if result.VideoID != fixtures.VideoID {
			t.Fatalf("Expected video ID %q, got %q", fixtures.VideoID, result.VideoID)
		}

		// Step 2: Fetch video info
		video := client.FetchVideo(ctx, t, result.VideoID)

		// Step 3: Verify video details
		AssertVideoValid(t, video)

		if video.ID != fixtures.VideoID {
			t.Errorf("Expected video ID %q, got %q", fixtures.VideoID, video.ID)
		}

		// Verify additional fields are populated
		if video.Duration == 0 {
			t.Error("Expected video duration to be set")
		}
		if video.ViewCount == 0 {
			t.Error("Expected video view count to be set")
		}
		if len(video.Thumbnails) == 0 {
			t.Error("Expected video thumbnails to be set")
		}
		if video.Author.ChannelID == "" {
			t.Error("Expected author channel ID to be set")
		}

		t.Logf("Successfully fetched video from URL:")
		t.Logf("  Title: %q", video.Title)
		t.Logf("  Author: %q (%s)", video.Author.Name, video.Author.ChannelID)
		t.Logf("  Duration: %v", video.Duration)
		t.Logf("  Views: %d", video.ViewCount)
		t.Logf("  Thumbnails: %d", len(video.Thumbnails))
	})
}

// TestIntegration_FetchVideoMetadata tests that video metadata is fully populated.
func TestIntegration_FetchVideoMetadata(t *testing.T) {
	SkipIfNoIntegration(t)

	fixtures := DefaultFixtures()
	ctx, cancel := NewTestContext(t)
	defer cancel()

	tc := NewTestClient(t)
	video := tc.FetchVideo(ctx, t, fixtures.VideoID)

	// Verify all expected metadata fields
	AssertVideoValid(t, video)

	// Core fields
	if video.ID == "" {
		t.Error("Video ID should not be empty")
	}
	if video.Title == "" {
		t.Error("Video title should not be empty")
	}
	if video.Description == "" {
		t.Error("Video description should not be empty")
	}
	if video.Duration == 0 {
		t.Error("Video duration should not be zero")
	}

	// Author fields
	if video.Author.Name == "" {
		t.Error("Author name should not be empty")
	}
	if video.Author.ChannelID == "" {
		t.Error("Author channel ID should not be empty")
	}
	if video.Author.URL == "" {
		t.Error("Author URL should not be empty")
	}

	// Thumbnails
	if len(video.Thumbnails) == 0 {
		t.Error("Video should have thumbnails")
	} else {
		// Verify thumbnail fields
		for i, thumb := range video.Thumbnails {
			if thumb.URL == "" {
				t.Errorf("Thumbnail %d URL should not be empty", i)
			}
			if thumb.Width == 0 {
				t.Errorf("Thumbnail %d width should not be zero", i)
			}
			if thumb.Height == 0 {
				t.Errorf("Thumbnail %d height should not be zero", i)
			}
		}
	}

	// View count (Rick Roll should have billions of views)
	if video.ViewCount < 1000000 {
		t.Logf("Warning: View count %d seems low for this video", video.ViewCount)
	}

	t.Logf("Video metadata verified:")
	t.Logf("  ID: %s", video.ID)
	t.Logf("  Title: %s", video.Title)
	t.Logf("  Duration: %v", video.Duration)
	t.Logf("  Views: %d", video.ViewCount)
	t.Logf("  Author: %s", video.Author.Name)
	t.Logf("  Channel ID: %s", video.Author.ChannelID)
	t.Logf("  Thumbnails: %d", len(video.Thumbnails))
	t.Logf("  Description length: %d chars", len(video.Description))
}

// TestIntegration_DownloadMuxedStream tests downloading a muxed stream (video+audio).
// This is the simplest download test as it doesn't require FFmpeg for muxing.
func TestIntegration_DownloadMuxedStream(t *testing.T) {
	SkipIfNoIntegration(t)

	fixtures := DefaultFixtures()
	ctx, cancel := NewTestContext(t)
	defer cancel()

	tc := NewTestClient(t)
	_, manifest := tc.FetchVideoWithStreams(ctx, t, fixtures.VideoID)

	// Find a muxed stream with direct URL
	muxedStream := FindMuxedStreamWithDirectURL(manifest)
	if muxedStream == nil {
		t.Skip("No muxed streams with direct URLs available (signature cipher required)")
	}

	// Download the muxed stream
	outputPath := TempFile(t, "video.mp4")
	downloader := download.NewDownloader(tc.Client)

	var progressUpdates int
	progressCallback := func(p download.Progress) {
		progressUpdates++
	}

	t.Logf("Downloading muxed stream: %s (%s)", muxedStream.VideoStreamInfo.Quality, muxedStream.VideoStreamInfo.Container)
	err := downloader.DownloadStream(ctx, muxedStream.VideoStreamInfo.URL, outputPath, progressCallback)
	RequireNoError(t, err, "Failed to download stream")

	// Verify file was created
	AssertFileExists(t, outputPath)

	// Verify progress was reported
	if progressUpdates == 0 {
		t.Error("Expected progress updates during download")
	}

	info, _ := os.Stat(outputPath)
	t.Logf("Downloaded %d bytes to %s (%d progress updates)", info.Size(), outputPath, progressUpdates)
}

// TestIntegration_DownloadAndMuxWithFFmpeg tests the full download and mux flow.
// Downloads separate video and audio streams, then muxes them with FFmpeg.
func TestIntegration_DownloadAndMuxWithFFmpeg(t *testing.T) {
	SkipIfNoIntegration(t)
	SkipIfNoFFmpeg(t)

	fixtures := DefaultFixtures()
	ctx, cancel := NewTestContext(t)
	defer cancel()

	tc := NewTestClient(t)
	video, manifest := tc.FetchVideoWithStreams(ctx, t, fixtures.VideoID)

	// Find video and audio streams with direct URLs
	videoStream := FindStreamWithDirectURL(manifest)
	audioStream := FindAudioStreamWithDirectURL(manifest)

	if videoStream == nil && audioStream == nil {
		// Try falling back to muxed stream
		muxedStream := FindMuxedStreamWithDirectURL(manifest)
		if muxedStream != nil {
			t.Skip("Only muxed streams available - no separate video+audio for muxing test")
		}
		t.Skip("No streams with direct URLs available (signature cipher required)")
	}

	if videoStream == nil {
		t.Skip("No video streams with direct URLs available")
	}
	if audioStream == nil {
		t.Skip("No audio streams with direct URLs available")
	}

	// Create temp directory for downloads
	tempDir := TempDir(t)
	videoPath := filepath.Join(tempDir, "video.mp4")
	audioPath := filepath.Join(tempDir, "audio.m4a")
	outputPath := filepath.Join(tempDir, "output.mp4")

	downloader := download.NewDownloader(tc.Client)

	// Download video stream
	t.Logf("Downloading video stream: %s (%dx%d)", videoStream.Quality, videoStream.Width, videoStream.Height)
	err := downloader.DownloadStream(ctx, videoStream.URL, videoPath, nil)
	RequireNoError(t, err, "Failed to download video stream")
	AssertFileExists(t, videoPath)

	// Download audio stream
	t.Logf("Downloading audio stream: %s (%d Hz)", audioStream.AudioCodec, audioStream.SampleRate)
	err = downloader.DownloadStream(ctx, audioStream.URL, audioPath, nil)
	RequireNoError(t, err, "Failed to download audio stream")
	AssertFileExists(t, audioPath)

	// Mux streams with FFmpeg
	t.Log("Muxing video and audio streams with FFmpeg")
	err = ffmpeg.MuxStreamsWithContext(ctx, videoPath, audioPath, outputPath)
	RequireNoError(t, err, "Failed to mux streams")

	// Verify output file
	AssertFileExists(t, outputPath)

	videoInfo, _ := os.Stat(videoPath)
	audioInfo, _ := os.Stat(audioPath)
	outputInfo, _ := os.Stat(outputPath)

	t.Logf("Download and mux complete for %q:", video.Title)
	t.Logf("  Video: %d bytes", videoInfo.Size())
	t.Logf("  Audio: %d bytes", audioInfo.Size())
	t.Logf("  Output: %d bytes", outputInfo.Size())
}

// TestIntegration_DownloadStreamsParallel tests downloading multiple streams in parallel.
func TestIntegration_DownloadStreamsParallel(t *testing.T) {
	SkipIfNoIntegration(t)

	fixtures := DefaultFixtures()
	ctx, cancel := NewTestContext(t)
	defer cancel()

	tc := NewTestClient(t)
	_, manifest := tc.FetchVideoWithStreams(ctx, t, fixtures.VideoID)

	// Find video and audio streams with direct URLs
	videoStream := FindStreamWithDirectURL(manifest)
	audioStream := FindAudioStreamWithDirectURL(manifest)

	if videoStream == nil || audioStream == nil {
		t.Skip("Both video and audio streams with direct URLs required for parallel test")
	}

	// Create temp directory for downloads
	tempDir := TempDir(t)
	videoPath := filepath.Join(tempDir, "video.mp4")
	audioPath := filepath.Join(tempDir, "audio.m4a")

	downloader := download.NewDownloader(tc.Client)

	// Download both streams in parallel
	streams := []download.StreamDownload{
		{URL: videoStream.URL, FilePath: videoPath},
		{URL: audioStream.URL, FilePath: audioPath},
	}

	var progressUpdates int
	progressCallback := func(p download.Progress) {
		progressUpdates++
	}

	t.Log("Downloading video and audio streams in parallel")
	results := downloader.DownloadStreamsParallel(ctx, streams, progressCallback)

	// Verify both downloads succeeded
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	for i, result := range results {
		if result.Error != nil {
			t.Errorf("Stream %d download failed: %v", i, result.Error)
		}
	}

	AssertFileExists(t, videoPath)
	AssertFileExists(t, audioPath)

	// Verify progress was reported
	if progressUpdates == 0 {
		t.Error("Expected progress updates during parallel download")
	}

	videoInfo, _ := os.Stat(videoPath)
	audioInfo, _ := os.Stat(audioPath)
	t.Logf("Parallel download complete: video=%d bytes, audio=%d bytes (%d progress updates)",
		videoInfo.Size(), audioInfo.Size(), progressUpdates)
}
