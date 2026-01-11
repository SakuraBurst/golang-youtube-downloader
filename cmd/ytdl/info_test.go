package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestInfoCommandExists(t *testing.T) {
	rootCmd := newRootCmd()
	infoCmd, _, err := rootCmd.Find([]string{"info"})
	if err != nil {
		t.Fatalf("info command not found: %v", err)
	}
	if infoCmd.Use != "info <url>" {
		t.Errorf("expected Use to be 'info <url>', got %q", infoCmd.Use)
	}
}

func TestInfoCommandRequiresURL(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"info"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("info command should fail without URL argument")
	}
}

func TestInfoCommandAcceptsURL(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"info", "https://www.youtube.com/watch?v=dQw4w9WgXcQ"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("info command failed with valid URL: %v", err)
	}
}

func TestInfoCommandHelp(t *testing.T) {
	rootCmd := newRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"info", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("info help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "info") {
		t.Error("help should mention info")
	}
	if !strings.Contains(output, "metadata") || !strings.Contains(output, "video") {
		t.Error("help should mention video metadata")
	}
}

func TestInfoCommandShortDescription(t *testing.T) {
	rootCmd := newRootCmd()
	infoCmd, _, _ := rootCmd.Find([]string{"info"})

	if infoCmd.Short == "" {
		t.Error("info command should have a short description")
	}
}
