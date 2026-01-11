package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDownloadCommandExists(t *testing.T) {
	rootCmd := newRootCmd()
	downloadCmd, _, err := rootCmd.Find([]string{"download"})
	if err != nil {
		t.Fatalf("download command not found: %v", err)
	}
	if downloadCmd.Use != "download <url>" {
		t.Errorf("expected Use to be 'download <url>', got %q", downloadCmd.Use)
	}
}

func TestDownloadCommandRequiresURL(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"download"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("download command should fail without URL argument")
	}
}

func TestDownloadCommandAcceptsURL(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"download", "https://www.youtube.com/watch?v=dQw4w9WgXcQ"})

	// Command should not error on valid URL (even if download not implemented yet)
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("download command failed with valid URL: %v", err)
	}
}

func TestDownloadCommandHasOutputFlag(t *testing.T) {
	rootCmd := newRootCmd()
	downloadCmd, _, _ := rootCmd.Find([]string{"download"})

	flag := downloadCmd.Flags().Lookup("output")
	if flag == nil {
		t.Error("download command should have --output flag")
	}
}

func TestDownloadCommandHasQualityFlag(t *testing.T) {
	rootCmd := newRootCmd()
	downloadCmd, _, _ := rootCmd.Find([]string{"download"})

	flag := downloadCmd.Flags().Lookup("quality")
	if flag == nil {
		t.Error("download command should have --quality flag")
	}
}

func TestDownloadCommandHasFormatFlag(t *testing.T) {
	rootCmd := newRootCmd()
	downloadCmd, _, _ := rootCmd.Find([]string{"download"})

	flag := downloadCmd.Flags().Lookup("format")
	if flag == nil {
		t.Error("download command should have --format flag")
	}
}

func TestDownloadCommandHelp(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"download", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("download help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "download") {
		t.Error("help should mention download")
	}
	if !strings.Contains(output, "output") {
		t.Error("help should mention output flag")
	}
}
