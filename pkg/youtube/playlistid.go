package youtube

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// ErrInvalidPlaylistID is returned when a playlist ID cannot be parsed from the input.
var ErrInvalidPlaylistID = errors.New("invalid playlist ID")

// playlistIDRegex matches valid YouTube playlist IDs.
// Playlist IDs can be:
// - PL + 32 characters (user playlists)
// - WL, LL, LM (Watch Later, Liked, Library Music)
// - RD + video ID (auto-generated mix)
// - OL + characters (album playlists)
// - UU + characters (channel uploads)
// - FL + characters (favorites)
var playlistIDRegex = regexp.MustCompile(`^(PL[a-zA-Z0-9_-]{32}|WL|LL|LM|RD[a-zA-Z0-9_-]+|OL[a-zA-Z0-9_-]+|OLAK5uy_[a-zA-Z0-9_-]+|UU[a-zA-Z0-9_-]+|FL[a-zA-Z0-9_-]+)$`)

// IsValidPlaylistID checks if the given string is a valid YouTube playlist ID.
func IsValidPlaylistID(id string) bool {
	return playlistIDRegex.MatchString(id)
}

// ParsePlaylistID extracts the playlist ID from a YouTube URL or validates a raw playlist ID.
// Supported URL formats:
//   - https://www.youtube.com/playlist?list=PLAYLIST_ID
//   - https://www.youtube.com/watch?v=VIDEO_ID&list=PLAYLIST_ID
//   - PLAYLIST_ID (raw playlist ID)
func ParsePlaylistID(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", ErrInvalidPlaylistID
	}

	// Check if input is already a valid playlist ID
	if IsValidPlaylistID(input) {
		return input, nil
	}

	// Try to parse as URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return "", ErrInvalidPlaylistID
	}

	// Check if it's a YouTube URL with a list parameter
	if !isYouTubeHost(parsedURL.Host) {
		return "", ErrInvalidPlaylistID
	}

	playlistID := parsedURL.Query().Get("list")
	if playlistID == "" {
		return "", ErrInvalidPlaylistID
	}

	// Validate the extracted ID
	if !IsValidPlaylistID(playlistID) {
		return "", ErrInvalidPlaylistID
	}

	return playlistID, nil
}

// isYouTubeHost checks if the host is a YouTube domain.
func isYouTubeHost(host string) bool {
	host = strings.ToLower(host)
	return host == "youtube.com" ||
		host == "www.youtube.com" ||
		host == "m.youtube.com" ||
		host == "youtu.be"
}
