package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/SakuraBurst/golang-youtube-downloader/pkg/download"
	"github.com/SakuraBurst/golang-youtube-downloader/pkg/ffmpeg"
	"github.com/SakuraBurst/golang-youtube-downloader/pkg/filename"
	"github.com/SakuraBurst/golang-youtube-downloader/pkg/youtube"
)

type downloadOptions struct {
	output  string
	quality string
	format  string
}

func newDownloadCmd() *cobra.Command {
	opts := &downloadOptions{}

	cmd := &cobra.Command{
		Use:   "download <url>",
		Short: "Download a YouTube video, playlist, or channel",
		Long: `Download YouTube content from the given URL.

Supports various YouTube URL formats including:
  - Video: https://www.youtube.com/watch?v=VIDEO_ID
  - Video: https://youtu.be/VIDEO_ID
  - Playlist: https://www.youtube.com/playlist?list=PLAYLIST_ID
  - Channel: https://www.youtube.com/channel/CHANNEL_ID
  - Channel: https://www.youtube.com/@handle`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			return runDownload(cmd, url, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.output, "output", "o", ".", "Output directory for downloaded files")
	cmd.Flags().StringVarP(&opts.quality, "quality", "q", "best", "Video quality (best, 1080p, 720p, 480p, 360p, audio)")
	cmd.Flags().StringVarP(&opts.format, "format", "f", "mp4", "Output format (mp4, webm, mp3)")

	return cmd
}

func runDownload(cmd *cobra.Command, url string, opts *downloadOptions) error {
	if url == "" {
		return errors.New("URL is required")
	}

	// Create default dependencies
	fetcher := &youtube.WatchPageFetcher{
		Client: http.DefaultClient,
	}
	downloader := download.NewDownloader(http.DefaultClient)

	err := runDownloadWithDeps(cmd.Context(), cmd.OutOrStdout(), url, opts, fetcher, downloader, ffmpeg.MuxStreamsWithContext)
	if err != nil {
		// Wrap the error with user-friendly message
		return WrapError(err)
	}
	return nil
}

// MuxerFunc is a function type for muxing video and audio streams.
type MuxerFunc func(ctx context.Context, videoPath, audioPath, outputPath string) error

// runDownloadWithDeps implements the download command logic with injectable dependencies.
func runDownloadWithDeps(
	ctx context.Context,
	w io.Writer,
	urlStr string,
	opts *downloadOptions,
	fetcher *youtube.WatchPageFetcher,
	downloader *download.Downloader,
	muxer MuxerFunc,
) error {
	// Resolve the query to determine content type
	query, err := youtube.ResolveQuery(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL or ID: %w", err)
	}

	switch query.Type {
	case youtube.QueryTypeVideo:
		return downloadSingleVideo(ctx, w, query.VideoID, opts, fetcher, downloader, muxer, "")

	case youtube.QueryTypePlaylist:
		return downloadPlaylist(ctx, w, query.PlaylistID, opts, fetcher, downloader, muxer)

	case youtube.QueryTypeChannel:
		return downloadChannel(ctx, w, query.Channel, opts, fetcher, downloader, muxer)

	case youtube.QueryTypeSearch:
		return errors.New("search queries are not supported for download")

	default:
		return errors.New("unsupported content type")
	}
}

// downloadSingleVideo downloads a single video by its ID.
func downloadSingleVideo(
	ctx context.Context,
	w io.Writer,
	videoID string,
	opts *downloadOptions,
	fetcher *youtube.WatchPageFetcher,
	downloader *download.Downloader,
	muxer MuxerFunc,
	numberPrefix string,
) error {
	_, _ = fmt.Fprintf(w, "Fetching video info: %s\n", videoID)

	// Fetch the watch page
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

	_, _ = fmt.Fprintf(w, "Title: %s\n", video.Title)
	_, _ = fmt.Fprintf(w, "Author: %s\n", video.Author.Name)
	_, _ = fmt.Fprintf(w, "Duration: %s\n", video.DurationString())

	// Check if we have streaming data
	if playerResponse.StreamingData == nil {
		return errors.New("no streaming data available")
	}

	// Get stream manifest
	manifest := playerResponse.StreamingData.GetStreamManifest()

	// Determine if audio-only mode
	audioOnly := strings.EqualFold(opts.format, "mp3") || strings.EqualFold(opts.quality, "audio")

	// Get preferred container
	container := parseContainer(opts.format)

	// Determine output path
	containerStr := string(container)
	if audioOnly {
		containerStr = "mp3"
	}
	outputFilename := filename.ApplyTemplate(filename.DefaultTemplate, video, containerStr, numberPrefix)
	outputPath := filepath.Join(opts.output, outputFilename)

	if audioOnly {
		return downloadAudioOnly(ctx, w, manifest, outputPath, downloader)
	}

	// Get quality preference and select best option
	quality := parseQualityPreference(opts.quality)
	options := manifest.GetDownloadOptions()
	selectedOption := youtube.SelectBestOption(options, quality, container)

	if selectedOption == nil {
		// Try to use muxed stream if no adaptive option is available
		if len(manifest.MuxedStreams) > 0 {
			return downloadMuxedStream(ctx, w, &manifest.MuxedStreams[0], outputPath, downloader)
		}
		return errors.New("no suitable stream found for the requested quality")
	}

	_, _ = fmt.Fprintf(w, "Selected quality: %s\n", selectedOption.QualityLabel())

	// Check if we need to mux separate streams
	if selectedOption.VideoStream != nil && selectedOption.AudioStream != nil && selectedOption.VideoStream.URL != "" {
		// Check if streams have separate URLs (need muxing)
		if selectedOption.AudioStream.URL != "" && selectedOption.VideoStream.URL != selectedOption.AudioStream.URL {
			return downloadAndMux(ctx, w, video, selectedOption, outputPath, downloader, muxer)
		}
	}

	// Download single stream (muxed or video-only)
	if selectedOption.VideoStream != nil && selectedOption.VideoStream.URL != "" {
		return downloadSingleStream(ctx, w, selectedOption.VideoStream.URL, outputPath, downloader)
	}

	// Fallback to first muxed stream
	if len(manifest.MuxedStreams) > 0 && manifest.MuxedStreams[0].VideoStreamInfo.URL != "" {
		return downloadMuxedStream(ctx, w, &manifest.MuxedStreams[0], outputPath, downloader)
	}

	return errors.New("no downloadable stream found")
}

// downloadSingleStream downloads a single stream to the output path.
func downloadSingleStream(ctx context.Context, w io.Writer, url, outputPath string, downloader *download.Downloader) error {
	_, _ = fmt.Fprintf(w, "Downloading to: %s\n", outputPath)

	// Create a progress bar
	bar := progressbar.NewOptions64(
		-1, // Unknown size initially
		progressbar.OptionSetWriter(w),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprintln(w)
		}),
	)

	progressCallback := func(p download.Progress) {
		if p.Total > 0 && bar.GetMax64() != p.Total {
			bar.ChangeMax64(p.Total)
		}
		_ = bar.Set64(p.Downloaded)
	}

	err := downloader.DownloadStream(ctx, url, outputPath, progressCallback)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	_ = bar.Finish()
	_, _ = fmt.Fprintf(w, "Download complete: %s\n", outputPath)
	return nil
}

// downloadMuxedStream downloads a muxed stream.
func downloadMuxedStream(ctx context.Context, w io.Writer, stream *youtube.MuxedStreamInfo, outputPath string, downloader *download.Downloader) error {
	if stream.VideoStreamInfo.URL == "" {
		return errors.New("muxed stream has no URL")
	}
	return downloadSingleStream(ctx, w, stream.VideoStreamInfo.URL, outputPath, downloader)
}

// downloadAudioOnly downloads audio-only stream.
func downloadAudioOnly(ctx context.Context, w io.Writer, manifest *youtube.StreamManifest, outputPath string, downloader *download.Downloader) error {
	bestAudio := manifest.GetBestAudioStream()
	if bestAudio == nil {
		return errors.New("no audio stream available")
	}

	if bestAudio.URL == "" {
		return errors.New("audio stream has no URL")
	}

	_, _ = fmt.Fprintf(w, "Downloading audio: %s\n", bestAudio.AudioCodec)
	return downloadSingleStream(ctx, w, bestAudio.URL, outputPath, downloader)
}

// downloadAndMux downloads video and audio streams separately and muxes them.
func downloadAndMux(
	ctx context.Context,
	w io.Writer,
	video *youtube.Video,
	option *youtube.DownloadOption,
	outputPath string,
	downloader *download.Downloader,
	muxer MuxerFunc,
) error {
	// Create temp directory for intermediate files
	tempDir, err := os.MkdirTemp("", "ytdl-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Download video stream with progress bar
	videoPath := filepath.Join(tempDir, "video."+string(option.VideoStream.Container))
	_, _ = fmt.Fprintf(w, "Downloading video stream...\n")
	if err := downloadStreamWithProgress(ctx, w, downloader, option.VideoStream.URL, videoPath, "Video"); err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}

	// Download audio stream with progress bar
	audioPath := filepath.Join(tempDir, "audio."+string(option.AudioStream.Container))
	_, _ = fmt.Fprintf(w, "Downloading audio stream...\n")
	if err := downloadStreamWithProgress(ctx, w, downloader, option.AudioStream.URL, audioPath, "Audio"); err != nil {
		return fmt.Errorf("failed to download audio: %w", err)
	}

	// Mux streams together
	if muxer == nil {
		return errors.New("muxer not available (FFmpeg required)")
	}

	_, _ = fmt.Fprintf(w, "Muxing streams...\n")
	if err := muxer(ctx, videoPath, audioPath, outputPath); err != nil {
		return fmt.Errorf("failed to mux streams: %w", err)
	}

	_, _ = fmt.Fprintf(w, "Download complete: %s\n", outputPath)
	return nil
}

// downloadStreamWithProgress downloads a stream with a progress bar.
func downloadStreamWithProgress(ctx context.Context, w io.Writer, downloader *download.Downloader, url, filePath, description string) error {
	bar := progressbar.NewOptions64(
		-1,
		progressbar.OptionSetWriter(w),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprintln(w)
		}),
	)

	progressCallback := func(p download.Progress) {
		if p.Total > 0 && bar.GetMax64() != p.Total {
			bar.ChangeMax64(p.Total)
		}
		_ = bar.Set64(p.Downloaded)
	}

	err := downloader.DownloadStream(ctx, url, filePath, progressCallback)
	if err != nil {
		return err
	}

	_ = bar.Finish()
	return nil
}

// parseQualityPreference converts a quality string to VideoQualityPreference.
func parseQualityPreference(quality string) youtube.VideoQualityPreference {
	switch strings.ToLower(quality) {
	case "best", "highest":
		return youtube.QualityHighest
	case "1080p", "1080":
		return youtube.QualityUpTo1080p
	case "720p", "720":
		return youtube.QualityUpTo720p
	case "480p", "480":
		return youtube.QualityUpTo480p
	case "360p", "360":
		return youtube.QualityUpTo360p
	case "worst", "lowest", "audio":
		return youtube.QualityLowest
	default:
		return youtube.QualityHighest
	}
}

// parseContainer converts a format string to Container.
func parseContainer(format string) youtube.Container {
	switch strings.ToLower(format) {
	case "webm":
		return youtube.ContainerWebM
	case "mp3":
		return youtube.ContainerMP3
	case "mp4":
		return youtube.ContainerMP4
	default:
		return youtube.ContainerMP4
	}
}

// downloadPlaylist downloads all videos from a playlist.
func downloadPlaylist(
	ctx context.Context,
	w io.Writer,
	playlistID string,
	opts *downloadOptions,
	fetcher *youtube.WatchPageFetcher,
	downloader *download.Downloader,
	muxer MuxerFunc,
) error {
	_, _ = fmt.Fprintf(w, "Playlist download: %s\n", playlistID)
	_, _ = fmt.Fprintf(w, "Note: Full playlist fetching requires additional API implementation.\n")
	_, _ = fmt.Fprintf(w, "Currently, only individual video downloads are fully supported.\n")

	// For now, we'll indicate this is a placeholder for future implementation
	// A complete implementation would:
	// 1. Fetch the playlist page
	// 2. Parse the initial data to get video list
	// 3. Handle pagination for playlists with many videos
	// 4. Download each video in sequence or parallel

	// The youtube package has the playlist parsing logic, but we need to add
	// a playlist page fetcher similar to WatchPageFetcher

	return errors.New("playlist download requires fetching playlist page - not yet implemented")
}

// downloadChannel downloads all videos from a channel.
func downloadChannel(
	ctx context.Context,
	w io.Writer,
	channel youtube.ChannelIdentifier,
	opts *downloadOptions,
	fetcher *youtube.WatchPageFetcher,
	downloader *download.Downloader,
	muxer MuxerFunc,
) error {
	_, _ = fmt.Fprintf(w, "Channel download: %s (%s)\n", channel.Value, channel.Type)

	// For channel IDs, we can convert to uploads playlist
	if channel.Type == youtube.ChannelTypeID {
		uploadsPlaylistID := channel.UploadsPlaylistID()
		if uploadsPlaylistID != "" {
			_, _ = fmt.Fprintf(w, "Converting to uploads playlist: %s\n", uploadsPlaylistID)
			return downloadPlaylist(ctx, w, uploadsPlaylistID, opts, fetcher, downloader, muxer)
		}
	}

	// For handles, custom URLs, and users, we would need to resolve to channel ID first
	_, _ = fmt.Fprintf(w, "Note: Channel handles and custom URLs require additional resolution.\n")

	return errors.New("channel download requires resolving channel ID - not yet implemented")
}
