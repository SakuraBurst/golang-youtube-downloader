package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
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

	// TODO: Implement actual video info fetching
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Fetching info for: %s\n", url)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Title: (not implemented)\n")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Author: (not implemented)\n")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Duration: (not implemented)\n")

	return nil
}
