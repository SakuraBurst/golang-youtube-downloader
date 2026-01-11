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
