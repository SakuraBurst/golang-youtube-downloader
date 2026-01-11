package main

import (
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
