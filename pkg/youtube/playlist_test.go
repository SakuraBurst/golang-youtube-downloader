package youtube

import (
	"testing"
)

func TestPlaylist_Fields(t *testing.T) {
	playlist := Playlist{
		ID:         "PLtest123",
		Title:      "Test Playlist",
		Author:     Author{Name: "Test Channel", ChannelID: "UCtest123"},
		VideoCount: 10,
	}

	if playlist.ID != "PLtest123" {
		t.Errorf("ID = %q, want %q", playlist.ID, "PLtest123")
	}
	if playlist.Title != "Test Playlist" {
		t.Errorf("Title = %q, want %q", playlist.Title, "Test Playlist")
	}
	if playlist.Author.Name != "Test Channel" {
		t.Errorf("Author.Name = %q, want %q", playlist.Author.Name, "Test Channel")
	}
	if playlist.VideoCount != 10 {
		t.Errorf("VideoCount = %d, want %d", playlist.VideoCount, 10)
	}
}

func TestParsePlaylistMetadata_ExtractsTitle(t *testing.T) {
	// This is a simplified test with mock JSON data
	jsonData := `{"header":{"playlistHeaderRenderer":{"title":{"simpleText":"Test Playlist Title"}}}}`

	title, err := parsePlaylistTitle(jsonData)
	if err != nil {
		t.Fatalf("parsePlaylistTitle failed: %v", err)
	}
	if title != "Test Playlist Title" {
		t.Errorf("title = %q, want %q", title, "Test Playlist Title")
	}
}

func TestParsePlaylistMetadata_ExtractsVideoCount(t *testing.T) {
	tests := []struct {
		name  string
		json  string
		want  int
		error bool
	}{
		{
			name:  "simple count",
			json:  `{"header":{"playlistHeaderRenderer":{"numVideosText":{"runs":[{"text":"42"}]}}}}`,
			want:  42,
			error: false,
		},
		{
			name:  "count with text",
			json:  `{"header":{"playlistHeaderRenderer":{"numVideosText":{"runs":[{"text":"100 videos"}]}}}}`,
			want:  100,
			error: false,
		},
		{
			name:  "missing count",
			json:  `{"header":{"playlistHeaderRenderer":{}}}`,
			want:  0,
			error: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := parsePlaylistVideoCount(tt.json)
			if tt.error && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.error && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if count != tt.want {
				t.Errorf("count = %d, want %d", count, tt.want)
			}
		})
	}
}

func TestParsePlaylistMetadata_ExtractsAuthor(t *testing.T) {
	jsonData := `{"header":{"playlistHeaderRenderer":{"ownerText":{"runs":[{"text":"Channel Name","navigationEndpoint":{"browseEndpoint":{"browseId":"UCtest123"}}}]}}}}`

	author, err := parsePlaylistAuthor(jsonData)
	if err != nil {
		t.Fatalf("parsePlaylistAuthor failed: %v", err)
	}
	if author.Name != "Channel Name" {
		t.Errorf("author.Name = %q, want %q", author.Name, "Channel Name")
	}
	if author.ChannelID != "UCtest123" {
		t.Errorf("author.ChannelID = %q, want %q", author.ChannelID, "UCtest123")
	}
}

func TestPlaylistVideo_Fields(t *testing.T) {
	pv := PlaylistVideo{
		ID:    "dQw4w9WgXcQ",
		Title: "Test Video",
		Author: Author{
			Name:      "Test Channel",
			ChannelID: "UCtest123",
		},
		DurationSeconds: 212,
		Index:           5,
	}

	if pv.ID != "dQw4w9WgXcQ" {
		t.Errorf("ID = %q, want %q", pv.ID, "dQw4w9WgXcQ")
	}
	if pv.Title != "Test Video" {
		t.Errorf("Title = %q, want %q", pv.Title, "Test Video")
	}
	if pv.Index != 5 {
		t.Errorf("Index = %d, want %d", pv.Index, 5)
	}
}

func TestParsePlaylistVideos_ExtractsVideos(t *testing.T) {
	// Mock JSON with playlist video renderers
	jsonData := `{
		"contents": {
			"twoColumnBrowseResultsRenderer": {
				"tabs": [{
					"tabRenderer": {
						"content": {
							"sectionListRenderer": {
								"contents": [{
									"itemSectionRenderer": {
										"contents": [{
											"playlistVideoListRenderer": {
												"contents": [
													{
														"playlistVideoRenderer": {
															"videoId": "video1",
															"title": {"runs": [{"text": "First Video"}]},
															"lengthSeconds": "120",
															"index": {"simpleText": "1"},
															"shortBylineText": {"runs": [{"text": "Channel One", "navigationEndpoint": {"browseEndpoint": {"browseId": "UC111"}}}]}
														}
													},
													{
														"playlistVideoRenderer": {
															"videoId": "video2",
															"title": {"runs": [{"text": "Second Video"}]},
															"lengthSeconds": "300",
															"index": {"simpleText": "2"},
															"shortBylineText": {"runs": [{"text": "Channel Two", "navigationEndpoint": {"browseEndpoint": {"browseId": "UC222"}}}]}
														}
													}
												]
											}
										}]
									}
								}]
							}
						}
					}
				}]
			}
		}
	}`

	videos, continuation, err := parsePlaylistVideos(jsonData)
	if err != nil {
		t.Fatalf("parsePlaylistVideos failed: %v", err)
	}

	if len(videos) != 2 {
		t.Fatalf("got %d videos, want 2", len(videos))
	}

	// Check first video
	if videos[0].ID != "video1" {
		t.Errorf("videos[0].ID = %q, want %q", videos[0].ID, "video1")
	}
	if videos[0].Title != "First Video" {
		t.Errorf("videos[0].Title = %q, want %q", videos[0].Title, "First Video")
	}
	if videos[0].DurationSeconds != 120 {
		t.Errorf("videos[0].DurationSeconds = %d, want %d", videos[0].DurationSeconds, 120)
	}
	if videos[0].Index != 1 {
		t.Errorf("videos[0].Index = %d, want %d", videos[0].Index, 1)
	}

	// Check second video
	if videos[1].ID != "video2" {
		t.Errorf("videos[1].ID = %q, want %q", videos[1].ID, "video2")
	}
	if videos[1].Title != "Second Video" {
		t.Errorf("videos[1].Title = %q, want %q", videos[1].Title, "Second Video")
	}

	// No continuation in this test
	if continuation != "" {
		t.Errorf("continuation = %q, want empty", continuation)
	}
}

func TestParsePlaylistVideos_ExtractsContinuation(t *testing.T) {
	// Mock JSON with continuation token
	jsonData := `{
		"contents": {
			"twoColumnBrowseResultsRenderer": {
				"tabs": [{
					"tabRenderer": {
						"content": {
							"sectionListRenderer": {
								"contents": [{
									"itemSectionRenderer": {
										"contents": [{
											"playlistVideoListRenderer": {
												"contents": [
													{
														"playlistVideoRenderer": {
															"videoId": "video1",
															"title": {"runs": [{"text": "Video"}]},
															"lengthSeconds": "60",
															"index": {"simpleText": "1"}
														}
													},
													{
														"continuationItemRenderer": {
															"continuationEndpoint": {
																"continuationCommand": {
																	"token": "continuation_token_123"
																}
															}
														}
													}
												]
											}
										}]
									}
								}]
							}
						}
					}
				}]
			}
		}
	}`

	videos, continuation, err := parsePlaylistVideos(jsonData)
	if err != nil {
		t.Fatalf("parsePlaylistVideos failed: %v", err)
	}

	if len(videos) != 1 {
		t.Fatalf("got %d videos, want 1", len(videos))
	}

	if continuation != "continuation_token_123" {
		t.Errorf("continuation = %q, want %q", continuation, "continuation_token_123")
	}
}

func TestParsePlaylistContinuation_ExtractsVideos(t *testing.T) {
	// Mock continuation response JSON
	jsonData := `{
		"onResponseReceivedActions": [{
			"appendContinuationItemsAction": {
				"continuationItems": [
					{
						"playlistVideoRenderer": {
							"videoId": "video3",
							"title": {"runs": [{"text": "Third Video"}]},
							"lengthSeconds": "180",
							"index": {"simpleText": "3"}
						}
					},
					{
						"playlistVideoRenderer": {
							"videoId": "video4",
							"title": {"runs": [{"text": "Fourth Video"}]},
							"lengthSeconds": "240",
							"index": {"simpleText": "4"}
						}
					}
				]
			}
		}]
	}`

	videos, continuation, err := parsePlaylistContinuation(jsonData)
	if err != nil {
		t.Fatalf("parsePlaylistContinuation failed: %v", err)
	}

	if len(videos) != 2 {
		t.Fatalf("got %d videos, want 2", len(videos))
	}

	if videos[0].ID != "video3" {
		t.Errorf("videos[0].ID = %q, want %q", videos[0].ID, "video3")
	}
	if videos[1].ID != "video4" {
		t.Errorf("videos[1].ID = %q, want %q", videos[1].ID, "video4")
	}

	if continuation != "" {
		t.Errorf("continuation = %q, want empty", continuation)
	}
}
