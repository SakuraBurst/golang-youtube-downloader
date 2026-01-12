// Package filename provides utilities for generating filenames from templates.
package filename

import (
	"strings"

	"github.com/SakuraBurst/golang-youtube-downloader/pkg/youtube"
)

// invalidChars contains characters that are not allowed in filenames across platforms.
const invalidChars = `<>:"/\|?*`

// SanitizeFilename replaces invalid filename characters with underscores and trims spaces.
func SanitizeFilename(name string) string {
	var sb strings.Builder
	sb.Grow(len(name))

	for _, r := range name {
		if strings.ContainsRune(invalidChars, r) {
			sb.WriteRune('_')
		} else {
			sb.WriteRune(r)
		}
	}

	return strings.TrimSpace(sb.String())
}

// ApplyTemplate applies a template to generate a filename from video metadata.
// Supported placeholders:
//   - $title: Video title
//   - $author: Channel/author name
//   - $id: Video ID
//   - $uploadDate: Upload date in YYYY-MM-DD format
//   - $num: Playlist number in brackets [N] (empty if not provided)
//   - $numc: Playlist number without brackets (empty if not provided)
//
// The container extension is automatically appended.
// All placeholders are sanitized to remove invalid filename characters.
func ApplyTemplate(template string, video *youtube.Video, container, number string) string {
	result := template

	// Replace number placeholders first (they need special handling)
	if number != "" {
		result = strings.ReplaceAll(result, "$numc", number)
		result = strings.ReplaceAll(result, "$num", "["+number+"]")
	} else {
		result = strings.ReplaceAll(result, "$numc", "")
		result = strings.ReplaceAll(result, "$num", "")
	}

	// Replace video metadata placeholders
	result = strings.ReplaceAll(result, "$id", SanitizeFilename(video.ID))
	result = strings.ReplaceAll(result, "$title", SanitizeFilename(video.Title))
	result = strings.ReplaceAll(result, "$author", SanitizeFilename(video.Author.Name))

	// Format upload date
	uploadDate := ""
	if !video.UploadDate.IsZero() {
		uploadDate = video.UploadDate.Format("2006-01-02")
	}
	result = strings.ReplaceAll(result, "$uploadDate", uploadDate)

	// Trim and append extension
	result = strings.TrimSpace(result)
	return result + "." + container
}

// DefaultTemplate is the default filename template.
const DefaultTemplate = "$title"
