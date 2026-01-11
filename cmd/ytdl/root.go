package main

import (
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ytdl",
		Short: "YouTube downloader CLI",
		Long: `ytdl - A CLI tool for downloading YouTube videos, playlists, and channel content.

This is a Go port of YoutubeDownloader (https://github.com/Tyrrrz/YoutubeDownloader).
It supports downloading videos in various formats and qualities.`,
		Run: func(cmd *cobra.Command, _ []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(newVersionCmd())

	return cmd
}
