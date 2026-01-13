package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/SakuraBurst/golang-youtube-downloader/pkg/youtube"
)

func newInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <url>",
		Short: "Show video metadata",
		Long: `Display metadata information for a YouTube video.

Shows details including:
  - Title
  - Author/Channel
  - Duration
  - Available formats and qualities`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			return runInfo(cmd, url)
		},
	}

	return cmd
}

func runInfo(cmd *cobra.Command, url string) error {
	if url == "" {
		return errors.New("URL is required")
	}

	// Create a default fetcher with standard HTTP client
	fetcher := &youtube.WatchPageFetcher{
		Client: http.DefaultClient,
	}

	return runInfoWithFetcher(cmd.Context(), cmd.OutOrStdout(), url, fetcher)
}

// runInfoWithFetcher implements the info command logic with a configurable fetcher.
// This allows for dependency injection in tests.
func runInfoWithFetcher(ctx context.Context, w io.Writer, urlStr string, fetcher *youtube.WatchPageFetcher) error {
	// Parse the video ID from the URL
	videoID, err := youtube.ParseVideoID(urlStr)
	if err != nil {
		return fmt.Errorf("invalid video URL or ID: %w", err)
	}

	// Fetch the watch page
	_, _ = fmt.Fprintf(w, "Fetching info for video: %s\n\n", videoID)

	watchPage, err := fetcher.Fetch(ctx, videoID)
	if err != nil {
		return fmt.Errorf("failed to fetch video page: %w", err)
	}

	// Extract player response
	playerResponse, err := watchPage.ExtractPlayerResponse()
	if err != nil {
		return fmt.Errorf("failed to extract video data: %w", err)
	}

	// Check playability status
	if playerResponse.PlayabilityStatus.Status != "OK" {
		reason := playerResponse.PlayabilityStatus.Reason
		if reason == "" {
			reason = "unknown reason"
		}
		return fmt.Errorf("video unavailable: %s", reason)
	}

	// Convert to Video struct
	video, err := playerResponse.ToVideo()
	if err != nil {
		return fmt.Errorf("failed to parse video metadata: %w", err)
	}

	// Display video information
	_, _ = fmt.Fprintf(w, "Title:    %s\n", video.Title)
	_, _ = fmt.Fprintf(w, "Author:   %s\n", video.Author.Name)
	_, _ = fmt.Fprintf(w, "Duration: %s\n", video.DurationString())
	_, _ = fmt.Fprintf(w, "Views:    %d\n", video.ViewCount)

	if video.IsLive {
		_, _ = fmt.Fprintf(w, "Status:   Live Stream\n")
	}

	// Display available formats
	if playerResponse.StreamingData != nil {
		manifest := playerResponse.StreamingData.GetStreamManifest()
		displayStreamInfo(w, manifest)
	}

	return nil
}

// displayStreamInfo outputs information about available streams.
func displayStreamInfo(w io.Writer, manifest *youtube.StreamManifest) {
	_, _ = fmt.Fprintf(w, "\nAvailable Formats:\n")

	// Video streams
	if len(manifest.VideoStreams) > 0 {
		_, _ = fmt.Fprintf(w, "\n  Video:\n")
		for i := range manifest.VideoStreams {
			vs := &manifest.VideoStreams[i]
			quality := vs.Quality
			if quality == "" {
				quality = youtube.QualityLabel(vs.Height)
			}
			_, _ = fmt.Fprintf(w, "    - %s (%s, %s)\n", quality, vs.Container, vs.VideoCodec)
		}
	}

	// Audio streams
	if len(manifest.AudioStreams) > 0 {
		_, _ = fmt.Fprintf(w, "\n  Audio:\n")
		for i := range manifest.AudioStreams {
			as := &manifest.AudioStreams[i]
			_, _ = fmt.Fprintf(w, "    - %s (%s, %dkbps)\n", as.Container, as.AudioCodec, as.Bitrate/1000)
		}
	}

	// Muxed streams
	if len(manifest.MuxedStreams) > 0 {
		_, _ = fmt.Fprintf(w, "\n  Muxed (Video+Audio):\n")
		for i := range manifest.MuxedStreams {
			ms := &manifest.MuxedStreams[i]
			quality := ms.VideoStreamInfo.Quality
			if quality == "" {
				quality = youtube.QualityLabel(ms.Height)
			}
			_, _ = fmt.Fprintf(w, "    - %s (%s)\n", quality, ms.VideoStreamInfo.Container)
		}
	}
}
