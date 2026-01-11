package youtube

import (
	"testing"
)

func TestParseVideoID_StandardWatchURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"http://www.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"https://youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=PLtest", "dQw4w9WgXcQ"},
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=120", "dQw4w9WgXcQ"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			id, err := ParseVideoID(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, id)
			}
		})
	}
}

func TestParseVideoID_ShortURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://youtu.be/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"http://youtu.be/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"https://youtu.be/dQw4w9WgXcQ?t=120", "dQw4w9WgXcQ"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			id, err := ParseVideoID(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, id)
			}
		})
	}
}

func TestParseVideoID_EmbedURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://www.youtube.com/embed/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"http://www.youtube.com/embed/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"https://youtube.com/embed/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			id, err := ParseVideoID(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, id)
			}
		})
	}
}

func TestParseVideoID_VURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://www.youtube.com/v/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"http://www.youtube.com/v/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			id, err := ParseVideoID(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, id)
			}
		})
	}
}

func TestParseVideoID_RawID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"_-AbCdEfGhI", "_-AbCdEfGhI"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			id, err := ParseVideoID(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, id)
			}
		})
	}
}

func TestParseVideoID_Invalid(t *testing.T) {
	tests := []string{
		"",
		"not-a-valid-url",
		"https://www.google.com",
		"https://www.youtube.com/",
		"https://www.youtube.com/watch",
		"https://www.youtube.com/watch?v=",
		"https://www.youtube.com/watch?v=short",
		"https://www.youtube.com/watch?v=toolongtobevalid123",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			_, err := ParseVideoID(tt)
			if err == nil {
				t.Errorf("expected error for input %q", tt)
			}
		})
	}
}

func TestIsValidVideoID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"dQw4w9WgXcQ", true},
		{"_-AbCdEfGhI", true},
		{"12345678901", true},
		{"", false},
		{"short", false},
		{"toolongtobevalid", false},
		{"invalid!char", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			result := IsValidVideoID(tt.id)
			if result != tt.valid {
				t.Errorf("IsValidVideoID(%q) = %v, want %v", tt.id, result, tt.valid)
			}
		})
	}
}
