package youtube

import (
	"encoding/json"
	"regexp"
	"strconv"
)

// Playlist represents a YouTube playlist with its metadata.
type Playlist struct {
	// ID is the playlist identifier.
	ID string

	// Title is the playlist's title.
	Title string

	// Author contains information about the playlist's creator.
	Author Author

	// VideoCount is the number of videos in the playlist.
	VideoCount int

	// Description is the playlist's description (may be empty).
	Description string

	// Thumbnails are the available thumbnail images for the playlist.
	Thumbnails []Thumbnail
}

// parsePlaylistTitle extracts the title from playlist JSON data.
func parsePlaylistTitle(jsonData string) (string, error) {
	var data struct {
		Header struct {
			PlaylistHeaderRenderer struct {
				Title struct {
					SimpleText string `json:"simpleText"`
				} `json:"title"`
			} `json:"playlistHeaderRenderer"`
		} `json:"header"`
	}

	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return "", err
	}

	return data.Header.PlaylistHeaderRenderer.Title.SimpleText, nil
}

// parsePlaylistVideoCount extracts the video count from playlist JSON data.
func parsePlaylistVideoCount(jsonData string) (int, error) {
	var data struct {
		Header struct {
			PlaylistHeaderRenderer struct {
				NumVideosText struct {
					Runs []struct {
						Text string `json:"text"`
					} `json:"runs"`
				} `json:"numVideosText"`
			} `json:"playlistHeaderRenderer"`
		} `json:"header"`
	}

	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return 0, err
	}

	runs := data.Header.PlaylistHeaderRenderer.NumVideosText.Runs
	if len(runs) == 0 {
		return 0, nil
	}

	// Extract number from text like "42" or "100 videos"
	text := runs[0].Text
	numRegex := regexp.MustCompile(`\d+`)
	match := numRegex.FindString(text)
	if match == "" {
		return 0, nil
	}

	count, err := strconv.Atoi(match)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// parsePlaylistAuthor extracts the author from playlist JSON data.
func parsePlaylistAuthor(jsonData string) (Author, error) {
	var data struct {
		Header struct {
			PlaylistHeaderRenderer struct {
				OwnerText struct {
					Runs []struct {
						Text               string `json:"text"`
						NavigationEndpoint struct {
							BrowseEndpoint struct {
								BrowseID string `json:"browseId"`
							} `json:"browseEndpoint"`
						} `json:"navigationEndpoint"`
					} `json:"runs"`
				} `json:"ownerText"`
			} `json:"playlistHeaderRenderer"`
		} `json:"header"`
	}

	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return Author{}, err
	}

	runs := data.Header.PlaylistHeaderRenderer.OwnerText.Runs
	if len(runs) == 0 {
		return Author{}, nil
	}

	return Author{
		Name:      runs[0].Text,
		ChannelID: runs[0].NavigationEndpoint.BrowseEndpoint.BrowseID,
	}, nil
}
