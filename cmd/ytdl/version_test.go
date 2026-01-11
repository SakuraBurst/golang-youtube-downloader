package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommandExists(t *testing.T) {
	rootCmd := newRootCmd()
	versionCmd, _, err := rootCmd.Find([]string{"version"})
	if err != nil {
		t.Fatalf("version command not found: %v", err)
	}
	if versionCmd.Use != "version" {
		t.Errorf("expected Use to be 'version', got %q", versionCmd.Use)
	}
}

func TestVersionCommandOutput(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ytdl") {
		t.Error("version output should contain 'ytdl'")
	}
	if !strings.Contains(strings.ToLower(output), "version") {
		t.Error("version output should contain version info")
	}
}

func TestVersionCommandShowsCommit(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()
	// Should show commit info (can be "unknown" or actual commit hash)
	if !strings.Contains(strings.ToLower(output), "commit") {
		t.Error("version output should contain commit info")
	}
}

func TestVersionCommandShowsBuildDate(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()
	// Should show build date info (can be "unknown" or actual date)
	if !strings.Contains(strings.ToLower(output), "build") || !strings.Contains(strings.ToLower(output), "date") {
		t.Error("version output should contain build date info")
	}
}
