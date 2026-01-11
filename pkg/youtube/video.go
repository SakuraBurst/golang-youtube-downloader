package youtube

import (
	"fmt"
	"time"
)

// Video represents a YouTube video with all its metadata.
type Video struct {
	// ID is the unique 11-character video identifier.
	ID string

	// Title is the video's title.
	Title string

	// Author contains information about the video's uploader/channel.
	Author Author

	// Duration is the length of the video.
	Duration time.Duration

	// Description is the video's description text.
	Description string

	// ViewCount is the number of views the video has.
	ViewCount int64

	// LikeCount is the number of likes (may be hidden by uploader).
	LikeCount int64

	// UploadDate is when the video was uploaded.
	UploadDate time.Time

	// Thumbnails are the available thumbnail images for the video.
	Thumbnails []Thumbnail

	// Keywords are the video's tags/keywords.
	Keywords []string

	// Category is the video's category (e.g., "Music", "Gaming").
	Category string

	// IsLive indicates if this is a live stream.
	IsLive bool

	// IsPrivate indicates if the video is private.
	IsPrivate bool
}

// String returns a string representation of the video.
func (v *Video) String() string {
	return fmt.Sprintf("%s - %s (%s)", v.Author.Name, v.Title, v.DurationString())
}

// DurationString returns the duration formatted as HH:MM:SS or MM:SS.
func (v *Video) DurationString() string {
	d := v.Duration
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// Author represents the channel/uploader of a video.
type Author struct {
	// Name is the channel's display name.
	Name string

	// ChannelID is the unique channel identifier.
	ChannelID string

	// URL is the channel's URL.
	URL string
}

// Thumbnail represents a video thumbnail image.
type Thumbnail struct {
	// URL is the thumbnail image URL.
	URL string

	// Width is the image width in pixels.
	Width int

	// Height is the image height in pixels.
	Height int
}

// Resolution returns the total pixel count for comparison.
func (t Thumbnail) Resolution() int {
	return t.Width * t.Height
}

// GetBestThumbnail returns the highest resolution thumbnail from a slice.
// Returns nil if the slice is empty.
func GetBestThumbnail(thumbnails []Thumbnail) *Thumbnail {
	if len(thumbnails) == 0 {
		return nil
	}

	best := &thumbnails[0]
	for i := range thumbnails {
		if thumbnails[i].Resolution() > best.Resolution() {
			best = &thumbnails[i]
		}
	}
	return best
}
