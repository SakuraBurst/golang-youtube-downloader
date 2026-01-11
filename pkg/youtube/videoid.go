// Package youtube provides utilities for parsing YouTube URLs and video information.
package youtube

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// ErrInvalidVideoID is returned when a video ID cannot be parsed from the input.
var ErrInvalidVideoID = errors.New("invalid video ID")

// videoIDRegex matches a valid YouTube video ID (11 characters: alphanumeric, underscore, hyphen).
var videoIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`)

// IsValidVideoID checks if the given string is a valid YouTube video ID.
// Valid video IDs are exactly 11 characters consisting of letters, numbers, underscores, and hyphens.
func IsValidVideoID(id string) bool {
	return videoIDRegex.MatchString(id)
}

// ParseVideoID extracts the video ID from a YouTube URL or validates a raw video ID.
// Supported URL formats:
//   - https://www.youtube.com/watch?v=VIDEO_ID
//   - https://youtu.be/VIDEO_ID
//   - https://www.youtube.com/embed/VIDEO_ID
//   - https://www.youtube.com/v/VIDEO_ID
//   - VIDEO_ID (raw 11-character ID)
func ParseVideoID(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", ErrInvalidVideoID
	}

	// Check if input is already a valid video ID
	if IsValidVideoID(input) {
		return input, nil
	}

	// Try to parse as URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return "", ErrInvalidVideoID
	}

	var videoID string

	switch {
	case isYouTubeWatchURL(parsedURL):
		// youtube.com/watch?v=VIDEO_ID
		videoID = parsedURL.Query().Get("v")

	case isYouTubeShortURL(parsedURL):
		// youtu.be/VIDEO_ID
		videoID = strings.TrimPrefix(parsedURL.Path, "/")

	case isYouTubeEmbedURL(parsedURL):
		// youtube.com/embed/VIDEO_ID
		videoID = extractPathID(parsedURL.Path, "/embed/")

	case isYouTubeVURL(parsedURL):
		// youtube.com/v/VIDEO_ID
		videoID = extractPathID(parsedURL.Path, "/v/")

	default:
		return "", ErrInvalidVideoID
	}

	// Validate the extracted ID
	if !IsValidVideoID(videoID) {
		return "", ErrInvalidVideoID
	}

	return videoID, nil
}

// isYouTubeWatchURL checks if the URL is a standard YouTube watch URL.
func isYouTubeWatchURL(u *url.URL) bool {
	host := strings.ToLower(u.Host)
	return (host == "youtube.com" || host == "www.youtube.com" || host == "m.youtube.com") &&
		u.Path == "/watch" &&
		u.Query().Get("v") != ""
}

// isYouTubeShortURL checks if the URL is a youtu.be short URL.
func isYouTubeShortURL(u *url.URL) bool {
	host := strings.ToLower(u.Host)
	return host == "youtu.be" && len(u.Path) > 1
}

// isYouTubeEmbedURL checks if the URL is a YouTube embed URL.
func isYouTubeEmbedURL(u *url.URL) bool {
	host := strings.ToLower(u.Host)
	return (host == "youtube.com" || host == "www.youtube.com") &&
		strings.HasPrefix(u.Path, "/embed/")
}

// isYouTubeVURL checks if the URL is a YouTube /v/ URL.
func isYouTubeVURL(u *url.URL) bool {
	host := strings.ToLower(u.Host)
	return (host == "youtube.com" || host == "www.youtube.com") &&
		strings.HasPrefix(u.Path, "/v/")
}

// extractPathID extracts the video ID from a path with a given prefix.
func extractPathID(path, prefix string) string {
	id := strings.TrimPrefix(path, prefix)
	// Remove any query parameters or trailing slashes
	if idx := strings.IndexAny(id, "?/"); idx != -1 {
		id = id[:idx]
	}
	return id
}
