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

func TestWatchPage_ExtractPlayerResponse_Success(t *testing.T) {
	// Sample HTML with ytInitialPlayerResponse embedded
	html := `<!DOCTYPE html>
<html>
<head><title>Test Video</title></head>
<body>
<script>var ytInitialPlayerResponse = {"videoDetails":{"videoId":"dQw4w9WgXcQ","title":"Test Video","author":"Test Channel"},"playabilityStatus":{"status":"OK"}};</script>
</body>
</html>`

	page := &WatchPage{
		VideoID: "dQw4w9WgXcQ",
		HTML:    html,
	}

	playerResponse, err := page.ExtractPlayerResponse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if playerResponse == nil {
		t.Fatal("expected player response to be non-nil")
	}
	if playerResponse.VideoDetails.VideoID != "dQw4w9WgXcQ" {
		t.Errorf("expected videoId %q, got %q", "dQw4w9WgXcQ", playerResponse.VideoDetails.VideoID)
	}
	if playerResponse.VideoDetails.Title != "Test Video" {
		t.Errorf("expected title %q, got %q", "Test Video", playerResponse.VideoDetails.Title)
	}
	if playerResponse.VideoDetails.Author != "Test Channel" {
		t.Errorf("expected author %q, got %q", "Test Channel", playerResponse.VideoDetails.Author)
	}
}

func TestWatchPage_ExtractPlayerResponse_WithWhitespace(t *testing.T) {
	// ytInitialPlayerResponse with various whitespace patterns
	html := `<!DOCTYPE html>
<script>
var   ytInitialPlayerResponse   =   {"videoDetails":{"videoId":"abc123XYZ90","title":"Whitespace Test"},"playabilityStatus":{"status":"OK"}}  ;
</script>`

	page := &WatchPage{
		VideoID: "abc123XYZ90",
		HTML:    html,
	}

	playerResponse, err := page.ExtractPlayerResponse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if playerResponse.VideoDetails.VideoID != "abc123XYZ90" {
		t.Errorf("expected videoId %q, got %q", "abc123XYZ90", playerResponse.VideoDetails.VideoID)
	}
}

func TestWatchPage_ExtractPlayerResponse_NotFound(t *testing.T) {
	// HTML without ytInitialPlayerResponse
	html := `<!DOCTYPE html>
<html>
<head><title>Test Video</title></head>
<body>
<script>var someOtherVar = {};</script>
</body>
</html>`

	page := &WatchPage{
		VideoID: "dQw4w9WgXcQ",
		HTML:    html,
	}

	_, err := page.ExtractPlayerResponse()
	if err == nil {
		t.Error("expected error when player response is not found")
	}
}

func TestWatchPage_ExtractPlayerResponse_InvalidJSON(t *testing.T) {
	// HTML with malformed JSON in ytInitialPlayerResponse
	html := `<!DOCTYPE html>
<script>var ytInitialPlayerResponse = {invalid json here};</script>`

	page := &WatchPage{
		VideoID: "dQw4w9WgXcQ",
		HTML:    html,
	}

	_, err := page.ExtractPlayerResponse()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestWatchPage_ExtractPlayerResponse_WithPlayabilityStatus(t *testing.T) {
	// Test playability status extraction
	html := `<!DOCTYPE html>
<script>var ytInitialPlayerResponse = {"videoDetails":{"videoId":"test123"},"playabilityStatus":{"status":"ERROR","reason":"Video unavailable"}};</script>`

	page := &WatchPage{
		VideoID: "test123",
		HTML:    html,
	}

	playerResponse, err := page.ExtractPlayerResponse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if playerResponse.PlayabilityStatus.Status != "ERROR" {
		t.Errorf("expected status %q, got %q", "ERROR", playerResponse.PlayabilityStatus.Status)
	}
	if playerResponse.PlayabilityStatus.Reason != "Video unavailable" {
		t.Errorf("expected reason %q, got %q", "Video unavailable", playerResponse.PlayabilityStatus.Reason)
	}
}

func TestWatchPage_ExtractPlayerResponse_NestedJSON(t *testing.T) {
	// More complex JSON structure with nested objects
	html := `<!DOCTYPE html>
<script>var ytInitialPlayerResponse = {"videoDetails":{"videoId":"nested123","title":"Nested Test","lengthSeconds":"120","viewCount":"1000","shortDescription":"A test video"},"playabilityStatus":{"status":"OK"},"streamingData":{"formats":[{"itag":18,"url":"https://example.com/stream"}]}};</script>`

	page := &WatchPage{
		VideoID: "nested123",
		HTML:    html,
	}

	playerResponse, err := page.ExtractPlayerResponse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if playerResponse.VideoDetails.VideoID != "nested123" {
		t.Errorf("expected videoId %q, got %q", "nested123", playerResponse.VideoDetails.VideoID)
	}
	if playerResponse.VideoDetails.LengthSeconds != "120" {
		t.Errorf("expected lengthSeconds %q, got %q", "120", playerResponse.VideoDetails.LengthSeconds)
	}
	if playerResponse.VideoDetails.ViewCount != "1000" {
		t.Errorf("expected viewCount %q, got %q", "1000", playerResponse.VideoDetails.ViewCount)
	}
	if playerResponse.StreamingData == nil {
		t.Error("expected streaming data to be non-nil")
	}
}
