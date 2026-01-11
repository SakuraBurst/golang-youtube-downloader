package youtube

import (
	"errors"
	"net/url"
	"strings"
)

// ErrUnresolvableQuery is returned when the input cannot be resolved to any known type.
var ErrUnresolvableQuery = errors.New("unresolvable query")

// QueryType represents the type of resolved query.
type QueryType string

const (
	// QueryTypeVideo indicates the query resolved to a video.
	QueryTypeVideo QueryType = "video"
	// QueryTypePlaylist indicates the query resolved to a playlist.
	QueryTypePlaylist QueryType = "playlist"
	// QueryTypeChannel indicates the query resolved to a channel.
	QueryTypeChannel QueryType = "channel"
	// QueryTypeSearch indicates the query should be treated as a search.
	QueryTypeSearch QueryType = "search"
)

// QueryResult contains the resolved query information.
type QueryResult struct {
	Type        QueryType
	VideoID     string
	PlaylistID  string
	Channel     ChannelIdentifier
	SearchQuery string
}

// ResolveQuery analyzes the input and determines what type of YouTube content it refers to.
// It handles:
//   - Video URLs and IDs
//   - Playlist URLs and IDs
//   - Channel URLs (all formats)
//   - Search queries (prefixed with ?)
//
// Priority order: Search (?) > Video > Playlist > Channel
func ResolveQuery(input string) (QueryResult, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return QueryResult{}, ErrUnresolvableQuery
	}

	// Check for explicit search query (starts with ?)
	if strings.HasPrefix(input, "?") {
		searchQuery := strings.TrimPrefix(input, "?")
		if searchQuery == "" {
			return QueryResult{}, ErrUnresolvableQuery
		}
		return QueryResult{
			Type:        QueryTypeSearch,
			SearchQuery: searchQuery,
		}, nil
	}

	// Try to parse as URL to check for combined video+playlist
	if parsedURL, err := url.Parse(input); err == nil && isYouTubeHost(parsedURL.Host) {
		// Check for watch URL with both video and playlist
		if strings.HasPrefix(parsedURL.Path, "/watch") {
			videoID := parsedURL.Query().Get("v")
			playlistID := parsedURL.Query().Get("list")

			if IsValidVideoID(videoID) {
				result := QueryResult{
					Type:    QueryTypeVideo,
					VideoID: videoID,
				}
				// Include playlist context if present
				if IsValidPlaylistID(playlistID) {
					result.PlaylistID = playlistID
				}
				return result, nil
			}
		}
	}

	// Try to resolve as video
	if videoID, err := ParseVideoID(input); err == nil {
		return QueryResult{
			Type:    QueryTypeVideo,
			VideoID: videoID,
		}, nil
	}

	// Try to resolve as playlist
	if playlistID, err := ParsePlaylistID(input); err == nil {
		return QueryResult{
			Type:       QueryTypePlaylist,
			PlaylistID: playlistID,
		}, nil
	}

	// Try to resolve as channel
	if channel, err := ParseChannelIdentifier(input); err == nil {
		return QueryResult{
			Type:    QueryTypeChannel,
			Channel: channel,
		}, nil
	}

	return QueryResult{}, ErrUnresolvableQuery
}
