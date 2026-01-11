package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommandExists(t *testing.T) {
	cmd := newRootCmd()
	if cmd == nil {
		t.Fatal("root command should not be nil")
	}
	if cmd.Use != "ytdl" {
		t.Errorf("expected Use to be 'ytdl', got %q", cmd.Use)
	}
}

func TestRootCommandIsValidCobraCommand(t *testing.T) {
	cmd := newRootCmd()
	if _, ok := interface{}(cmd).(*cobra.Command); !ok {
		t.Fatal("root command should be a *cobra.Command")
	}
}

func TestRootCommandHasShortDescription(t *testing.T) {
	cmd := newRootCmd()
	if cmd.Short == "" {
		t.Error("root command should have a short description")
	}
	if !strings.Contains(cmd.Short, "YouTube") {
		t.Error("short description should mention YouTube")
	}
}

func TestRootCommandHasLongDescription(t *testing.T) {
	cmd := newRootCmd()
	if cmd.Long == "" {
		t.Error("root command should have a long description")
	}
	if len(cmd.Long) < len(cmd.Short) {
		t.Error("long description should be longer than short description")
	}
}

func TestRootCommandShowsHelpByDefault(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("root command execution failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ytdl") {
		t.Error("help output should contain 'ytdl'")
	}
	if !strings.Contains(output, "Usage:") {
		t.Error("help output should contain 'Usage:'")
	}
}

func TestRootCommandHelpFlag(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Available Commands:") && !strings.Contains(output, "Flags:") {
		t.Error("help should show available commands or flags")
	}
}
