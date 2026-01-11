package youtube

import (
	"testing"
)

func TestResolveQuery_Video(t *testing.T) {
	tests := []string{
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"https://youtu.be/dQw4w9WgXcQ",
		"https://www.youtube.com/embed/dQw4w9WgXcQ",
		"dQw4w9WgXcQ",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			result, err := ResolveQuery(tt)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Type != QueryTypeVideo {
				t.Errorf("expected QueryTypeVideo, got %v", result.Type)
			}
			if result.VideoID != "dQw4w9WgXcQ" {
				t.Errorf("expected video ID 'dQw4w9WgXcQ', got %q", result.VideoID)
			}
		})
	}
}

func TestResolveQuery_Playlist(t *testing.T) {
	tests := []struct {
		input      string
		playlistID string
	}{
		{"https://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"},
		{"PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ResolveQuery(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Type != QueryTypePlaylist {
				t.Errorf("expected QueryTypePlaylist, got %v", result.Type)
			}
			if result.PlaylistID != tt.playlistID {
				t.Errorf("expected playlist ID %q, got %q", tt.playlistID, result.PlaylistID)
			}
		})
	}
}

func TestResolveQuery_Channel(t *testing.T) {
	tests := []struct {
		input       string
		channelType ChannelType
		channelVal  string
	}{
		{"https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw", ChannelTypeID, "UCuAXFkgsw1L7xaCfnd5JJOw"},
		{"https://www.youtube.com/@MrBeast", ChannelTypeHandle, "MrBeast"},
		{"https://www.youtube.com/c/MrBeast", ChannelTypeCustom, "MrBeast"},
		{"https://www.youtube.com/user/PewDiePie", ChannelTypeUser, "PewDiePie"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ResolveQuery(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Type != QueryTypeChannel {
				t.Errorf("expected QueryTypeChannel, got %v", result.Type)
			}
			if result.Channel.Type != tt.channelType {
				t.Errorf("expected channel type %v, got %v", tt.channelType, result.Channel.Type)
			}
			if result.Channel.Value != tt.channelVal {
				t.Errorf("expected channel value %q, got %q", tt.channelVal, result.Channel.Value)
			}
		})
	}
}

func TestResolveQuery_Search(t *testing.T) {
	tests := []struct {
		input       string
		searchQuery string
	}{
		{"?rick astley", "rick astley"},
		{"?how to program", "how to program"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ResolveQuery(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Type != QueryTypeSearch {
				t.Errorf("expected QueryTypeSearch, got %v", result.Type)
			}
			if result.SearchQuery != tt.searchQuery {
				t.Errorf("expected search query %q, got %q", tt.searchQuery, result.SearchQuery)
			}
		})
	}
}

func TestResolveQuery_VideoWithPlaylist(t *testing.T) {
	// When a URL contains both video and playlist, prioritize video
	input := "https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"

	result, err := ResolveQuery(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should detect as video with playlist context
	if result.Type != QueryTypeVideo {
		t.Errorf("expected QueryTypeVideo, got %v", result.Type)
	}
	if result.VideoID != "dQw4w9WgXcQ" {
		t.Errorf("expected video ID 'dQw4w9WgXcQ', got %q", result.VideoID)
	}
	if result.PlaylistID != "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf" {
		t.Errorf("expected playlist ID in context, got %q", result.PlaylistID)
	}
}

func TestResolveQuery_Invalid(t *testing.T) {
	tests := []string{
		"",
		"https://www.google.com",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			_, err := ResolveQuery(tt)
			if err == nil {
				t.Errorf("expected error for input %q", tt)
			}
		})
	}
}
