// Package ffmpeg provides utilities for detecting and working with FFmpeg.
package ffmpeg

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ErrNotFound is returned when FFmpeg is not found on the system.
var ErrNotFound = errors.New("ffmpeg not found")

// cliFileName returns the FFmpeg executable name for the current OS.
func cliFileName() string {
	if runtime.GOOS == "windows" {
		return "ffmpeg.exe"
	}
	return "ffmpeg"
}

// probeDirectoryPaths returns all directories to search for FFmpeg.
func probeDirectoryPaths() []string {
	var paths []string
	seen := make(map[string]bool)

	addPath := func(p string) {
		if p != "" && !seen[p] {
			seen[p] = true
			paths = append(paths, p)
		}
	}

	// Current working directory
	if wd, err := os.Getwd(); err == nil {
		addPath(wd)
	}

	// Executable directory (bundled location)
	if exe, err := os.Executable(); err == nil {
		addPath(filepath.Dir(exe))
	}

	// PATH environment variable
	pathEnv := os.Getenv("PATH")
	if pathEnv != "" {
		var sep string
		if runtime.GOOS == "windows" {
			sep = ";"
		} else {
			sep = ":"
		}
		for _, p := range strings.Split(pathEnv, sep) {
			addPath(p)
		}
	}

	return paths
}

// TryGetCliFilePath searches for the FFmpeg executable and returns its path.
// Returns nil if FFmpeg is not found.
func TryGetCliFilePath() *string {
	name := cliFileName()
	for _, dir := range probeDirectoryPaths() {
		fullPath := filepath.Join(dir, name)
		if _, err := os.Stat(fullPath); err == nil {
			return &fullPath
		}
	}
	return nil
}

// GetCliFilePath searches for the FFmpeg executable and returns its path.
// Returns ErrNotFound if FFmpeg is not found.
func GetCliFilePath() (string, error) {
	path := TryGetCliFilePath()
	if path == nil {
		return "", ErrNotFound
	}
	return *path, nil
}

// IsAvailable returns true if FFmpeg is available on the system.
func IsAvailable() bool {
	return TryGetCliFilePath() != nil
}

// IsBundled returns true if FFmpeg is bundled with the application
// (i.e., located in the same directory as the executable).
func IsBundled() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	bundledPath := filepath.Join(filepath.Dir(exe), cliFileName())
	_, err = os.Stat(bundledPath)
	return err == nil
}
