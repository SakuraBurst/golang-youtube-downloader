package youtube

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// ErrInvalidChannelID is returned when a channel identifier cannot be parsed from the input.
var ErrInvalidChannelID = errors.New("invalid channel identifier")

// ChannelType represents the type of channel identifier.
type ChannelType string

const (
	// ChannelTypeID is a raw channel ID (e.g., UCuAXFkgsw1L7xaCfnd5JJOw)
	ChannelTypeID ChannelType = "id"
	// ChannelTypeHandle is a channel handle (e.g., @MrBeast)
	ChannelTypeHandle ChannelType = "handle"
	// ChannelTypeCustom is a custom channel URL (e.g., /c/MrBeast)
	ChannelTypeCustom ChannelType = "custom"
	// ChannelTypeUser is a legacy user URL (e.g., /user/PewDiePie)
	ChannelTypeUser ChannelType = "user"
)

// ChannelIdentifier represents a parsed channel identifier with its type.
type ChannelIdentifier struct {
	Type  ChannelType
	Value string
}

// channelIDRegex matches a valid YouTube channel ID (24 characters starting with UC).
var channelIDRegex = regexp.MustCompile(`^UC[a-zA-Z0-9_-]{22}$`)

// IsValidChannelID checks if the given string is a valid YouTube channel ID.
// Valid channel IDs are 24 characters starting with "UC".
func IsValidChannelID(id string) bool {
	return channelIDRegex.MatchString(id)
}

// ParseChannelIdentifier extracts the channel identifier from a YouTube URL or validates a raw channel ID.
// Supported URL formats:
//   - https://www.youtube.com/channel/CHANNEL_ID
//   - https://www.youtube.com/@handle
//   - https://www.youtube.com/c/customname
//   - https://www.youtube.com/user/username
//   - CHANNEL_ID (raw 24-character ID starting with UC)
func ParseChannelIdentifier(input string) (ChannelIdentifier, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return ChannelIdentifier{}, ErrInvalidChannelID
	}

	// Check if input is already a valid channel ID
	if IsValidChannelID(input) {
		return ChannelIdentifier{Type: ChannelTypeID, Value: input}, nil
	}

	// Try to parse as URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return ChannelIdentifier{}, ErrInvalidChannelID
	}

	// Check if it's a YouTube URL
	if !isYouTubeHost(parsedURL.Host) {
		return ChannelIdentifier{}, ErrInvalidChannelID
	}

	path := strings.TrimSuffix(parsedURL.Path, "/")

	// Check for /channel/ID format
	if strings.HasPrefix(path, "/channel/") {
		channelID := strings.TrimPrefix(path, "/channel/")
		channelID = extractFirstPathSegment(channelID)
		if IsValidChannelID(channelID) {
			return ChannelIdentifier{Type: ChannelTypeID, Value: channelID}, nil
		}
		return ChannelIdentifier{}, ErrInvalidChannelID
	}

	// Check for /@handle format
	if strings.HasPrefix(path, "/@") {
		handle := strings.TrimPrefix(path, "/@")
		handle = extractFirstPathSegment(handle)
		if handle != "" {
			return ChannelIdentifier{Type: ChannelTypeHandle, Value: handle}, nil
		}
		return ChannelIdentifier{}, ErrInvalidChannelID
	}

	// Check for /c/customname format
	if strings.HasPrefix(path, "/c/") {
		customName := strings.TrimPrefix(path, "/c/")
		customName = extractFirstPathSegment(customName)
		if customName != "" {
			return ChannelIdentifier{Type: ChannelTypeCustom, Value: customName}, nil
		}
		return ChannelIdentifier{}, ErrInvalidChannelID
	}

	// Check for /user/username format
	if strings.HasPrefix(path, "/user/") {
		username := strings.TrimPrefix(path, "/user/")
		username = extractFirstPathSegment(username)
		if username != "" {
			return ChannelIdentifier{Type: ChannelTypeUser, Value: username}, nil
		}
		return ChannelIdentifier{}, ErrInvalidChannelID
	}

	return ChannelIdentifier{}, ErrInvalidChannelID
}

// extractFirstPathSegment extracts the first segment from a path (before any /).
func extractFirstPathSegment(path string) string {
	if idx := strings.Index(path, "/"); idx != -1 {
		return path[:idx]
	}
	return path
}
