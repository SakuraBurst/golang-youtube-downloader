package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
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
		Short: "Download a YouTube video",
		Long: `Download a YouTube video from the given URL.

Supports various YouTube URL formats including:
  - https://www.youtube.com/watch?v=VIDEO_ID
  - https://youtu.be/VIDEO_ID
  - https://www.youtube.com/embed/VIDEO_ID`,
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

	// TODO: Implement actual download logic
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Downloading: %s\n", url)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Output: %s\n", opts.output)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Quality: %s\n", opts.quality)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Format: %s\n", opts.format)

	return nil
}
