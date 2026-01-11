package main

import (
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ytdl",
		Short: "YouTube downloader CLI",
		Long:  "A CLI tool for downloading YouTube videos, playlists, and channel content.",
	}
	return cmd
}
