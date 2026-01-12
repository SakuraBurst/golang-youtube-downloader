package youtube

import (
	"testing"
)

func TestPlaylist_Fields(t *testing.T) {
	playlist := Playlist{
		ID:         "PLtest123",
		Title:      "Test Playlist",
		Author:     Author{Name: "Test Channel", ChannelID: "UCtest123"},
		VideoCount: 10,
	}

	if playlist.ID != "PLtest123" {
		t.Errorf("ID = %q, want %q", playlist.ID, "PLtest123")
	}
	if playlist.Title != "Test Playlist" {
		t.Errorf("Title = %q, want %q", playlist.Title, "Test Playlist")
	}
	if playlist.Author.Name != "Test Channel" {
		t.Errorf("Author.Name = %q, want %q", playlist.Author.Name, "Test Channel")
	}
	if playlist.VideoCount != 10 {
		t.Errorf("VideoCount = %d, want %d", playlist.VideoCount, 10)
	}
}

func TestParsePlaylistMetadata_ExtractsTitle(t *testing.T) {
	// This is a simplified test with mock JSON data
	jsonData := `{"header":{"playlistHeaderRenderer":{"title":{"simpleText":"Test Playlist Title"}}}}`

	title, err := parsePlaylistTitle(jsonData)
	if err != nil {
		t.Fatalf("parsePlaylistTitle failed: %v", err)
	}
	if title != "Test Playlist Title" {
		t.Errorf("title = %q, want %q", title, "Test Playlist Title")
	}
}

func TestParsePlaylistMetadata_ExtractsVideoCount(t *testing.T) {
	tests := []struct {
		name  string
		json  string
		want  int
		error bool
	}{
		{
			name:  "simple count",
			json:  `{"header":{"playlistHeaderRenderer":{"numVideosText":{"runs":[{"text":"42"}]}}}}`,
			want:  42,
			error: false,
		},
		{
			name:  "count with text",
			json:  `{"header":{"playlistHeaderRenderer":{"numVideosText":{"runs":[{"text":"100 videos"}]}}}}`,
			want:  100,
			error: false,
		},
		{
			name:  "missing count",
			json:  `{"header":{"playlistHeaderRenderer":{}}}`,
			want:  0,
			error: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := parsePlaylistVideoCount(tt.json)
			if tt.error && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.error && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if count != tt.want {
				t.Errorf("count = %d, want %d", count, tt.want)
			}
		})
	}
}

func TestParsePlaylistMetadata_ExtractsAuthor(t *testing.T) {
	jsonData := `{"header":{"playlistHeaderRenderer":{"ownerText":{"runs":[{"text":"Channel Name","navigationEndpoint":{"browseEndpoint":{"browseId":"UCtest123"}}}]}}}}`

	author, err := parsePlaylistAuthor(jsonData)
	if err != nil {
		t.Fatalf("parsePlaylistAuthor failed: %v", err)
	}
	if author.Name != "Channel Name" {
		t.Errorf("author.Name = %q, want %q", author.Name, "Channel Name")
	}
	if author.ChannelID != "UCtest123" {
		t.Errorf("author.ChannelID = %q, want %q", author.ChannelID, "UCtest123")
	}
}
