package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Build information set via ldflags
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display the version, commit hash, and build date of ytdl.",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ytdl Version: %s\nCommit: %s\nBuild Date: %s\n", version, commit, buildDate)
		},
	}
}
