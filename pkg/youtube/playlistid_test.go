package youtube

import (
	"testing"
)

func TestParsePlaylistID_PlaylistURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"},
		{"http://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"},
		{"https://youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			id, err := ParsePlaylistID(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, id)
			}
		})
	}
}

func TestParsePlaylistID_WatchURLWithList(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"},
		{"https://www.youtube.com/watch?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf&v=dQw4w9WgXcQ", "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			id, err := ParsePlaylistID(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, id)
			}
		})
	}
}

func TestParsePlaylistID_SpecialPlaylists(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		// Watch Later
		{"https://www.youtube.com/playlist?list=WL", "WL"},
		// Liked Videos
		{"https://www.youtube.com/playlist?list=LL", "LL"},
		// Mix/Auto-generated playlists (start with RD)
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=RDdQw4w9WgXcQ", "RDdQw4w9WgXcQ"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			id, err := ParsePlaylistID(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, id)
			}
		})
	}
}

func TestParsePlaylistID_RawID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"},
		{"WL", "WL"},
		{"LL", "LL"},
		{"RDdQw4w9WgXcQ", "RDdQw4w9WgXcQ"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			id, err := ParsePlaylistID(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, id)
			}
		})
	}
}

func TestParsePlaylistID_Invalid(t *testing.T) {
	tests := []string{
		"",
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"https://www.youtube.com/playlist",
		"https://www.youtube.com/playlist?list=",
		"https://www.google.com",
		"not-a-playlist-id!",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			_, err := ParsePlaylistID(tt)
			if err == nil {
				t.Errorf("expected error for input %q", tt)
			}
		})
	}
}

func TestIsValidPlaylistID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", true},
		{"WL", true},
		{"LL", true},
		{"LM", true},
		{"RDdQw4w9WgXcQ", true},
		{"OLAK5uy_test123", true},
		{"", false},
		{"invalid!char", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			result := IsValidPlaylistID(tt.id)
			if result != tt.valid {
				t.Errorf("IsValidPlaylistID(%q) = %v, want %v", tt.id, result, tt.valid)
			}
		})
	}
}
