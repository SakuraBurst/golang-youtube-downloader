package youtube

import (
	"testing"
	"time"
)

func TestVideo_HasRequiredFields(t *testing.T) {
	video := Video{
		ID:          "dQw4w9WgXcQ",
		Title:       "Never Gonna Give You Up",
		Author:      Author{Name: "Rick Astley", ChannelID: "UCuAXFkgsw1L7xaCfnd5JJOw"},
		Duration:    3*time.Minute + 33*time.Second,
		Description: "The official video for Never Gonna Give You Up",
		ViewCount:   1500000000,
		UploadDate:  time.Date(2009, 10, 25, 0, 0, 0, 0, time.UTC),
		Thumbnails: []Thumbnail{
			{URL: "https://i.ytimg.com/vi/dQw4w9WgXcQ/default.jpg", Width: 120, Height: 90},
		},
	}

	if video.ID != "dQw4w9WgXcQ" {
		t.Error("ID should be set")
	}
	if video.Title != "Never Gonna Give You Up" {
		t.Error("Title should be set")
	}
	if video.Author.Name != "Rick Astley" {
		t.Error("Author name should be set")
	}
	if video.Duration != 3*time.Minute+33*time.Second {
		t.Error("Duration should be set")
	}
	if video.UploadDate.Year() != 2009 {
		t.Error("UploadDate should be set")
	}
	if len(video.Thumbnails) != 1 {
		t.Error("Thumbnails should be set")
	}
}

func TestVideo_String(t *testing.T) {
	video := &Video{
		ID:       "dQw4w9WgXcQ",
		Title:    "Never Gonna Give You Up",
		Author:   Author{Name: "Rick Astley"},
		Duration: 3*time.Minute + 33*time.Second,
	}

	str := video.String()
	if str == "" {
		t.Error("String() should return non-empty string")
	}
}

func TestAuthor_HasRequiredFields(t *testing.T) {
	author := Author{
		Name:      "Rick Astley",
		ChannelID: "UCuAXFkgsw1L7xaCfnd5JJOw",
		URL:       "https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw",
	}

	if author.Name == "" {
		t.Error("Name should be set")
	}
	if author.ChannelID == "" {
		t.Error("ChannelID should be set")
	}
}

func TestThumbnail_HasRequiredFields(t *testing.T) {
	thumb := Thumbnail{
		URL:    "https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg",
		Width:  1280,
		Height: 720,
	}

	if thumb.URL == "" {
		t.Error("URL should be set")
	}
	if thumb.Width == 0 {
		t.Error("Width should be set")
	}
	if thumb.Height == 0 {
		t.Error("Height should be set")
	}
}

func TestThumbnail_GetBestQuality(t *testing.T) {
	thumbnails := []Thumbnail{
		{URL: "https://i.ytimg.com/vi/dQw4w9WgXcQ/default.jpg", Width: 120, Height: 90},
		{URL: "https://i.ytimg.com/vi/dQw4w9WgXcQ/mqdefault.jpg", Width: 320, Height: 180},
		{URL: "https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg", Width: 1280, Height: 720},
	}

	best := GetBestThumbnail(thumbnails)
	if best == nil {
		t.Fatal("GetBestThumbnail should return a thumbnail")
	}
	if best.Width != 1280 {
		t.Errorf("expected width 1280, got %d", best.Width)
	}
}

func TestThumbnail_GetBestQuality_Empty(t *testing.T) {
	var thumbnails []Thumbnail
	best := GetBestThumbnail(thumbnails)
	if best != nil {
		t.Error("GetBestThumbnail should return nil for empty slice")
	}
}

func TestVideo_DurationString(t *testing.T) {
	video := &Video{
		Duration: 1*time.Hour + 23*time.Minute + 45*time.Second,
	}

	str := video.DurationString()
	if str != "1:23:45" {
		t.Errorf("expected '1:23:45', got %q", str)
	}
}

func TestVideo_DurationString_Short(t *testing.T) {
	video := &Video{
		Duration: 3*time.Minute + 5*time.Second,
	}

	str := video.DurationString()
	if str != "3:05" {
		t.Errorf("expected '3:05', got %q", str)
	}
}

func TestPlayerResponse_ToVideo_Basic(t *testing.T) {
	pr := &PlayerResponse{
		VideoDetails: VideoDetailsResponse{
			VideoID:          "dQw4w9WgXcQ",
			Title:            "Never Gonna Give You Up",
			Author:           "Rick Astley",
			ChannelID:        "UCuAXFkgsw1L7xaCfnd5JJOw",
			LengthSeconds:    "213",
			ViewCount:        "1500000000",
			ShortDescription: "The official video for Never Gonna Give You Up",
			Keywords:         []string{"rick", "astley", "never", "gonna"},
			IsLiveContent:    false,
			IsPrivate:        false,
		},
		PlayabilityStatus: PlayabilityStatusResponse{
			Status: "OK",
		},
	}

	video, err := pr.ToVideo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if video.ID != "dQw4w9WgXcQ" {
		t.Errorf("expected ID %q, got %q", "dQw4w9WgXcQ", video.ID)
	}
	if video.Title != "Never Gonna Give You Up" {
		t.Errorf("expected Title %q, got %q", "Never Gonna Give You Up", video.Title)
	}
	if video.Author.Name != "Rick Astley" {
		t.Errorf("expected Author.Name %q, got %q", "Rick Astley", video.Author.Name)
	}
	if video.Author.ChannelID != "UCuAXFkgsw1L7xaCfnd5JJOw" {
		t.Errorf("expected Author.ChannelID %q, got %q", "UCuAXFkgsw1L7xaCfnd5JJOw", video.Author.ChannelID)
	}
	if video.Duration != 213*time.Second {
		t.Errorf("expected Duration %v, got %v", 213*time.Second, video.Duration)
	}
	if video.ViewCount != 1500000000 {
		t.Errorf("expected ViewCount %d, got %d", int64(1500000000), video.ViewCount)
	}
	if video.Description != "The official video for Never Gonna Give You Up" {
		t.Errorf("expected Description %q, got %q", "The official video for Never Gonna Give You Up", video.Description)
	}
	if len(video.Keywords) != 4 {
		t.Errorf("expected 4 keywords, got %d", len(video.Keywords))
	}
	if video.IsLive != false {
		t.Error("expected IsLive to be false")
	}
	if video.IsPrivate != false {
		t.Error("expected IsPrivate to be false")
	}
}

func TestPlayerResponse_ToVideo_WithThumbnails(t *testing.T) {
	pr := &PlayerResponse{
		VideoDetails: VideoDetailsResponse{
			VideoID:       "dQw4w9WgXcQ",
			Title:         "Test Video",
			Author:        "Test Author",
			ChannelID:     "UC123",
			LengthSeconds: "100",
			ViewCount:     "1000",
		},
		PlayabilityStatus: PlayabilityStatusResponse{
			Status: "OK",
		},
	}
	pr.VideoDetails.Thumbnail.Thumbnails = []ThumbnailResponse{
		{URL: "https://i.ytimg.com/vi/dQw4w9WgXcQ/default.jpg", Width: 120, Height: 90},
		{URL: "https://i.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg", Width: 1280, Height: 720},
	}

	video, err := pr.ToVideo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(video.Thumbnails) != 2 {
		t.Errorf("expected 2 thumbnails, got %d", len(video.Thumbnails))
	}
	if video.Thumbnails[0].Width != 120 {
		t.Errorf("expected first thumbnail width 120, got %d", video.Thumbnails[0].Width)
	}
	if video.Thumbnails[1].Width != 1280 {
		t.Errorf("expected second thumbnail width 1280, got %d", video.Thumbnails[1].Width)
	}
}

func TestPlayerResponse_ToVideo_InvalidDuration(t *testing.T) {
	pr := &PlayerResponse{
		VideoDetails: VideoDetailsResponse{
			VideoID:       "dQw4w9WgXcQ",
			Title:         "Test Video",
			Author:        "Test Author",
			ChannelID:     "UC123",
			LengthSeconds: "invalid",
			ViewCount:     "1000",
		},
		PlayabilityStatus: PlayabilityStatusResponse{
			Status: "OK",
		},
	}

	_, err := pr.ToVideo()
	if err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestPlayerResponse_ToVideo_InvalidViewCount(t *testing.T) {
	pr := &PlayerResponse{
		VideoDetails: VideoDetailsResponse{
			VideoID:       "dQw4w9WgXcQ",
			Title:         "Test Video",
			Author:        "Test Author",
			ChannelID:     "UC123",
			LengthSeconds: "100",
			ViewCount:     "invalid",
		},
		PlayabilityStatus: PlayabilityStatusResponse{
			Status: "OK",
		},
	}

	// Empty or invalid view count should default to 0, not error
	video, err := pr.ToVideo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if video.ViewCount != 0 {
		t.Errorf("expected ViewCount 0 for invalid value, got %d", video.ViewCount)
	}
}

func TestPlayerResponse_ToVideo_LiveVideo(t *testing.T) {
	pr := &PlayerResponse{
		VideoDetails: VideoDetailsResponse{
			VideoID:       "live123",
			Title:         "Live Stream",
			Author:        "Streamer",
			ChannelID:     "UC123",
			LengthSeconds: "0",
			ViewCount:     "5000",
			IsLiveContent: true,
		},
		PlayabilityStatus: PlayabilityStatusResponse{
			Status: "OK",
		},
	}

	video, err := pr.ToVideo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !video.IsLive {
		t.Error("expected IsLive to be true")
	}
}

func TestPlayerResponse_ToVideo_PrivateVideo(t *testing.T) {
	pr := &PlayerResponse{
		VideoDetails: VideoDetailsResponse{
			VideoID:       "private123",
			Title:         "Private Video",
			Author:        "Author",
			ChannelID:     "UC123",
			LengthSeconds: "60",
			ViewCount:     "0",
			IsPrivate:     true,
		},
		PlayabilityStatus: PlayabilityStatusResponse{
			Status: "OK",
		},
	}

	video, err := pr.ToVideo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !video.IsPrivate {
		t.Error("expected IsPrivate to be true")
	}
}

func TestPlayerResponse_ToVideo_AuthorURL(t *testing.T) {
	pr := &PlayerResponse{
		VideoDetails: VideoDetailsResponse{
			VideoID:       "test123",
			Title:         "Test",
			Author:        "Author",
			ChannelID:     "UCuAXFkgsw1L7xaCfnd5JJOw",
			LengthSeconds: "60",
			ViewCount:     "100",
		},
		PlayabilityStatus: PlayabilityStatusResponse{
			Status: "OK",
		},
	}

	video, err := pr.ToVideo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedURL := "https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw"
	if video.Author.URL != expectedURL {
		t.Errorf("expected Author.URL %q, got %q", expectedURL, video.Author.URL)
	}
}
