package youtube

import (
	"testing"
)

func TestParseChannelID_ChannelURL(t *testing.T) {
	tests := []struct {
		url      string
		expected ChannelIdentifier
	}{
		{
			"https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw",
			ChannelIdentifier{Type: ChannelTypeID, Value: "UCuAXFkgsw1L7xaCfnd5JJOw"},
		},
		{
			"http://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw",
			ChannelIdentifier{Type: ChannelTypeID, Value: "UCuAXFkgsw1L7xaCfnd5JJOw"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			id, err := ParseChannelIdentifier(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %+v, got %+v", tt.expected, id)
			}
		})
	}
}

func TestParseChannelID_HandleURL(t *testing.T) {
	tests := []struct {
		url      string
		expected ChannelIdentifier
	}{
		{
			"https://www.youtube.com/@MrBeast",
			ChannelIdentifier{Type: ChannelTypeHandle, Value: "MrBeast"},
		},
		{
			"https://youtube.com/@pewdiepie",
			ChannelIdentifier{Type: ChannelTypeHandle, Value: "pewdiepie"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			id, err := ParseChannelIdentifier(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %+v, got %+v", tt.expected, id)
			}
		})
	}
}

func TestParseChannelID_CustomURL(t *testing.T) {
	tests := []struct {
		url      string
		expected ChannelIdentifier
	}{
		{
			"https://www.youtube.com/c/MrBeast",
			ChannelIdentifier{Type: ChannelTypeCustom, Value: "MrBeast"},
		},
		{
			"https://youtube.com/c/pewdiepie",
			ChannelIdentifier{Type: ChannelTypeCustom, Value: "pewdiepie"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			id, err := ParseChannelIdentifier(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %+v, got %+v", tt.expected, id)
			}
		})
	}
}

func TestParseChannelID_UserURL(t *testing.T) {
	tests := []struct {
		url      string
		expected ChannelIdentifier
	}{
		{
			"https://www.youtube.com/user/PewDiePie",
			ChannelIdentifier{Type: ChannelTypeUser, Value: "PewDiePie"},
		},
		{
			"https://youtube.com/user/Google",
			ChannelIdentifier{Type: ChannelTypeUser, Value: "Google"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			id, err := ParseChannelIdentifier(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %+v, got %+v", tt.expected, id)
			}
		})
	}
}

func TestParseChannelID_RawChannelID(t *testing.T) {
	tests := []struct {
		input    string
		expected ChannelIdentifier
	}{
		{
			"UCuAXFkgsw1L7xaCfnd5JJOw",
			ChannelIdentifier{Type: ChannelTypeID, Value: "UCuAXFkgsw1L7xaCfnd5JJOw"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			id, err := ParseChannelIdentifier(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.expected {
				t.Errorf("expected %+v, got %+v", tt.expected, id)
			}
		})
	}
}

func TestParseChannelID_Invalid(t *testing.T) {
	tests := []string{
		"",
		"https://www.google.com",
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"https://www.youtube.com/",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			_, err := ParseChannelIdentifier(tt)
			if err == nil {
				t.Errorf("expected error for input %q", tt)
			}
		})
	}
}

func TestIsValidChannelID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"UCuAXFkgsw1L7xaCfnd5JJOw", true},
		{"UC-lHJZR3Gqxm24_Vd_AJ5Yw", true},
		{"", false},
		{"short", false},
		{"invalid!char", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			result := IsValidChannelID(tt.id)
			if result != tt.valid {
				t.Errorf("IsValidChannelID(%q) = %v, want %v", tt.id, result, tt.valid)
			}
		})
	}
}

func TestChannelToUploadsPlaylistID(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		want      string
	}{
		{
			name:      "typical channel ID",
			channelID: "UCuAXFkgsw1L7xaCfnd5JJOw",
			want:      "UUuAXFkgsw1L7xaCfnd5JJOw",
		},
		{
			name:      "channel with dashes",
			channelID: "UC-lHJZR3Gqxm24_Vd_AJ5Yw",
			want:      "UU-lHJZR3Gqxm24_Vd_AJ5Yw",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChannelToUploadsPlaylistID(tt.channelID)
			if got != tt.want {
				t.Errorf("ChannelToUploadsPlaylistID(%q) = %q, want %q", tt.channelID, got, tt.want)
			}
		})
	}
}

func TestChannelToUploadsPlaylistID_InvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
	}{
		{
			name:      "empty string",
			channelID: "",
		},
		{
			name:      "too short",
			channelID: "UC123",
		},
		{
			name:      "doesn't start with UC",
			channelID: "ABuAXFkgsw1L7xaCfnd5JJOw",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChannelToUploadsPlaylistID(tt.channelID)
			if got != "" {
				t.Errorf("ChannelToUploadsPlaylistID(%q) = %q, want empty string for invalid input", tt.channelID, got)
			}
		})
	}
}

func TestChannelIdentifier_UploadsPlaylistID(t *testing.T) {
	tests := []struct {
		name       string
		identifier ChannelIdentifier
		want       string
	}{
		{
			name:       "channel ID type",
			identifier: ChannelIdentifier{Type: ChannelTypeID, Value: "UCuAXFkgsw1L7xaCfnd5JJOw"},
			want:       "UUuAXFkgsw1L7xaCfnd5JJOw",
		},
		{
			name:       "handle type returns empty",
			identifier: ChannelIdentifier{Type: ChannelTypeHandle, Value: "MrBeast"},
			want:       "",
		},
		{
			name:       "custom type returns empty",
			identifier: ChannelIdentifier{Type: ChannelTypeCustom, Value: "pewdiepie"},
			want:       "",
		},
		{
			name:       "user type returns empty",
			identifier: ChannelIdentifier{Type: ChannelTypeUser, Value: "Google"},
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.identifier.UploadsPlaylistID()
			if got != tt.want {
				t.Errorf("UploadsPlaylistID() = %q, want %q", got, tt.want)
			}
		})
	}
}
