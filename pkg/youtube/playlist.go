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

// PlaylistVideo represents a video entry within a playlist.
type PlaylistVideo struct {
	// ID is the video's unique identifier.
	ID string

	// Title is the video's title.
	Title string

	// Author is the video's uploader/channel.
	Author Author

	// DurationSeconds is the video duration in seconds.
	DurationSeconds int

	// Index is the position of this video in the playlist (1-based).
	Index int

	// Thumbnails are the available thumbnail images.
	Thumbnails []Thumbnail
}

// playlistVideoRenderer represents the JSON structure for a playlist video item.
type playlistVideoRenderer struct {
	VideoID         string              `json:"videoId"`
	Title           runText             `json:"title"`
	LengthSeconds   string              `json:"lengthSeconds"`
	Index           simpleText          `json:"index"`
	ShortBylineText runTextWithEndpoint `json:"shortBylineText"`
	Thumbnail       thumbnailList       `json:"thumbnail"`
}

// runText represents a text field with "runs" array.
type runText struct {
	Runs []struct {
		Text string `json:"text"`
	} `json:"runs"`
}

// simpleText represents a text field with "simpleText".
type simpleText struct {
	SimpleText string `json:"simpleText"`
}

// runTextWithEndpoint represents a text field with navigation endpoint.
type runTextWithEndpoint struct {
	Runs []struct {
		Text               string `json:"text"`
		NavigationEndpoint struct {
			BrowseEndpoint struct {
				BrowseID string `json:"browseId"`
			} `json:"browseEndpoint"`
		} `json:"navigationEndpoint"`
	} `json:"runs"`
}

// thumbnailList represents a thumbnail container.
type thumbnailList struct {
	Thumbnails []ThumbnailResponse `json:"thumbnails"`
}

// getText extracts text from runText.
func (r runText) getText() string {
	if len(r.Runs) > 0 {
		return r.Runs[0].Text
	}
	return ""
}

// toPlaylistVideo converts a playlistVideoRenderer to PlaylistVideo.
func (pvr *playlistVideoRenderer) toPlaylistVideo() PlaylistVideo {
	duration, _ := strconv.Atoi(pvr.LengthSeconds)
	index, _ := strconv.Atoi(pvr.Index.SimpleText)

	var author Author
	if len(pvr.ShortBylineText.Runs) > 0 {
		author = Author{
			Name:      pvr.ShortBylineText.Runs[0].Text,
			ChannelID: pvr.ShortBylineText.Runs[0].NavigationEndpoint.BrowseEndpoint.BrowseID,
		}
	}

	thumbnails := make([]Thumbnail, len(pvr.Thumbnail.Thumbnails))
	for i, t := range pvr.Thumbnail.Thumbnails {
		thumbnails[i] = Thumbnail(t)
	}

	return PlaylistVideo{
		ID:              pvr.VideoID,
		Title:           pvr.Title.getText(),
		Author:          author,
		DurationSeconds: duration,
		Index:           index,
		Thumbnails:      thumbnails,
	}
}

// parsePlaylistVideos extracts video entries from playlist initial data JSON.
// Returns the list of videos and a continuation token if more videos are available.
func parsePlaylistVideos(jsonData string) ([]PlaylistVideo, string, error) {
	var data struct {
		Contents struct {
			TwoColumnBrowseResultsRenderer struct {
				Tabs []struct {
					TabRenderer struct {
						Content struct {
							SectionListRenderer struct {
								Contents []struct {
									ItemSectionRenderer struct {
										Contents []struct {
											PlaylistVideoListRenderer struct {
												Contents []json.RawMessage `json:"contents"`
											} `json:"playlistVideoListRenderer"`
										} `json:"contents"`
									} `json:"itemSectionRenderer"`
								} `json:"contents"`
							} `json:"sectionListRenderer"`
						} `json:"content"`
					} `json:"tabRenderer"`
				} `json:"tabs"`
			} `json:"twoColumnBrowseResultsRenderer"`
		} `json:"contents"`
	}

	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, "", err
	}

	var videos []PlaylistVideo
	var continuation string

	// Navigate to the playlist video list
	for _, tab := range data.Contents.TwoColumnBrowseResultsRenderer.Tabs {
		for _, section := range tab.TabRenderer.Content.SectionListRenderer.Contents {
			for _, item := range section.ItemSectionRenderer.Contents {
				for _, content := range item.PlaylistVideoListRenderer.Contents {
					video, cont := parsePlaylistContent(content)
					if video != nil {
						videos = append(videos, *video)
					}
					if cont != "" {
						continuation = cont
					}
				}
			}
		}
	}

	return videos, continuation, nil
}

// parsePlaylistContent parses a single content item from playlist video list.
// Returns either a PlaylistVideo or a continuation token.
func parsePlaylistContent(content json.RawMessage) (video *PlaylistVideo, continuationToken string) {
	// Try to parse as video renderer
	var videoWrapper struct {
		PlaylistVideoRenderer *playlistVideoRenderer `json:"playlistVideoRenderer"`
	}
	if err := json.Unmarshal(content, &videoWrapper); err == nil && videoWrapper.PlaylistVideoRenderer != nil {
		pv := videoWrapper.PlaylistVideoRenderer.toPlaylistVideo()
		return &pv, ""
	}

	// Try to parse as continuation item
	var contWrapper struct {
		ContinuationItemRenderer struct {
			ContinuationEndpoint struct {
				ContinuationCommand struct {
					Token string `json:"token"`
				} `json:"continuationCommand"`
			} `json:"continuationEndpoint"`
		} `json:"continuationItemRenderer"`
	}
	if err := json.Unmarshal(content, &contWrapper); err == nil {
		token := contWrapper.ContinuationItemRenderer.ContinuationEndpoint.ContinuationCommand.Token
		if token != "" {
			return nil, token
		}
	}

	return nil, ""
}

// parsePlaylistContinuation extracts videos from a continuation response.
// Returns the list of videos and a continuation token if more videos are available.
func parsePlaylistContinuation(jsonData string) ([]PlaylistVideo, string, error) {
	var data struct {
		OnResponseReceivedActions []struct {
			AppendContinuationItemsAction struct {
				ContinuationItems []json.RawMessage `json:"continuationItems"`
			} `json:"appendContinuationItemsAction"`
		} `json:"onResponseReceivedActions"`
	}

	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, "", err
	}

	var videos []PlaylistVideo
	var continuation string

	for _, action := range data.OnResponseReceivedActions {
		for _, content := range action.AppendContinuationItemsAction.ContinuationItems {
			video, cont := parsePlaylistContent(content)
			if video != nil {
				videos = append(videos, *video)
			}
			if cont != "" {
				continuation = cont
			}
		}
	}

	return videos, continuation, nil
}
