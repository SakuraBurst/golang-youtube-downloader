package filename

import (
	"testing"
	"time"

	"github.com/SakuraBurst/golang-youtube-downloader/pkg/youtube"
)

func TestApplyTemplate_BasicPlaceholders(t *testing.T) {
	video := youtube.Video{
		ID:         "dQw4w9WgXcQ",
		Title:      "Test Video Title",
		Author:     youtube.Author{Name: "Test Author"},
		UploadDate: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{
			name:     "title only",
			template: "$title",
			want:     "Test Video Title.mp4",
		},
		{
			name:     "id only",
			template: "$id",
			want:     "dQw4w9WgXcQ.mp4",
		},
		{
			name:     "author only",
			template: "$author",
			want:     "Test Author.mp4",
		},
		{
			name:     "upload date only",
			template: "$uploadDate",
			want:     "2024-03-15.mp4",
		},
		{
			name:     "title and author",
			template: "$title - $author",
			want:     "Test Video Title - Test Author.mp4",
		},
		{
			name:     "complex template",
			template: "$author - $title ($id)",
			want:     "Test Author - Test Video Title (dQw4w9WgXcQ).mp4",
		},
		{
			name:     "literal text only",
			template: "video",
			want:     "video.mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyTemplate(tt.template, &video, "mp4", "")
			if got != tt.want {
				t.Errorf("ApplyTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApplyTemplate_NumberPlaceholders(t *testing.T) {
	video := youtube.Video{
		ID:    "abc123",
		Title: "Test",
	}

	tests := []struct {
		name     string
		template string
		number   string
		want     string
	}{
		{
			name:     "num with number",
			template: "$num $title",
			number:   "5",
			want:     "[5] Test.mp4",
		},
		{
			name:     "num without number",
			template: "$num $title",
			number:   "",
			want:     "Test.mp4",
		},
		{
			name:     "numc with number",
			template: "$numc - $title",
			number:   "05",
			want:     "05 - Test.mp4",
		},
		{
			name:     "numc without number",
			template: "$numc - $title",
			number:   "",
			want:     "- Test.mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyTemplate(tt.template, &video, "mp4", tt.number)
			if got != tt.want {
				t.Errorf("ApplyTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal filename",
			input: "Test Video",
			want:  "Test Video",
		},
		{
			name:  "filename with slash",
			input: "Test/Video",
			want:  "Test_Video",
		},
		{
			name:  "filename with backslash",
			input: "Test\\Video",
			want:  "Test_Video",
		},
		{
			name:  "filename with colon",
			input: "Test:Video",
			want:  "Test_Video",
		},
		{
			name:  "filename with asterisk",
			input: "Test*Video",
			want:  "Test_Video",
		},
		{
			name:  "filename with question mark",
			input: "Test?Video",
			want:  "Test_Video",
		},
		{
			name:  "filename with quotes",
			input: `Test"Video`,
			want:  "Test_Video",
		},
		{
			name:  "filename with angle brackets",
			input: "Test<Video>",
			want:  "Test_Video_",
		},
		{
			name:  "filename with pipe",
			input: "Test|Video",
			want:  "Test_Video",
		},
		{
			name:  "multiple invalid chars",
			input: "Test:Video/\\*?",
			want:  "Test_Video____",
		},
		{
			name:  "leading and trailing spaces",
			input: "  Test Video  ",
			want:  "Test Video",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestApplyTemplate_SanitizesOutput(t *testing.T) {
	video := youtube.Video{
		ID:     "abc123",
		Title:  "Test:Video/Title",
		Author: youtube.Author{Name: "Author<Name>"},
	}

	got := ApplyTemplate("$author - $title", &video, "mp4", "")
	want := "Author_Name_ - Test_Video_Title.mp4"

	if got != want {
		t.Errorf("ApplyTemplate() = %q, want %q", got, want)
	}
}

func TestApplyTemplate_DifferentContainers(t *testing.T) {
	video := youtube.Video{
		ID:    "abc123",
		Title: "Test",
	}

	tests := []struct {
		container string
		want      string
	}{
		{"mp4", "Test.mp4"},
		{"webm", "Test.webm"},
		{"mp3", "Test.mp3"},
		{"m4a", "Test.m4a"},
	}

	for _, tt := range tests {
		t.Run(tt.container, func(t *testing.T) {
			got := ApplyTemplate("$title", &video, tt.container, "")
			if got != tt.want {
				t.Errorf("ApplyTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}
