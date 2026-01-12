package ffmpeg

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCliFileName(t *testing.T) {
	name := cliFileName()
	if runtime.GOOS == "windows" {
		if name != "ffmpeg.exe" {
			t.Errorf("Expected ffmpeg.exe on Windows, got %s", name)
		}
	} else {
		if name != "ffmpeg" {
			t.Errorf("Expected ffmpeg on non-Windows, got %s", name)
		}
	}
}

func TestTryGetCliFilePath_FindsInProvidedPath(t *testing.T) {
	// Create a temp directory with a fake ffmpeg
	tmpDir := t.TempDir()
	ffmpegPath := filepath.Join(tmpDir, cliFileName())
	if err := os.WriteFile(ffmpegPath, []byte("fake ffmpeg"), 0o755); err != nil {
		t.Fatalf("Failed to create fake ffmpeg: %v", err)
	}

	// Save current PATH and restore after test
	oldPath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", oldPath) }()

	// Set PATH to our temp directory
	var sep string
	if runtime.GOOS == "windows" {
		sep = ";"
	} else {
		sep = ":"
	}
	_ = os.Setenv("PATH", tmpDir+sep+oldPath)

	// Should find the ffmpeg
	result := TryGetCliFilePath()
	if result == nil {
		t.Fatal("Expected to find ffmpeg in PATH")
	}

	if *result != ffmpegPath {
		t.Errorf("Expected %s, got %s", ffmpegPath, *result)
	}
}

func TestTryGetCliFilePath_FindsInCurrentDirectory(t *testing.T) {
	// Create a temp directory with a fake ffmpeg
	tmpDir := t.TempDir()
	ffmpegPath := filepath.Join(tmpDir, cliFileName())
	if err := os.WriteFile(ffmpegPath, []byte("fake ffmpeg"), 0o755); err != nil {
		t.Fatalf("Failed to create fake ffmpeg: %v", err)
	}

	// Save current directory and restore after test
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Should find ffmpeg in current directory
	result := TryGetCliFilePath()
	if result == nil {
		t.Fatal("Expected to find ffmpeg in current directory")
	}
}

func TestTryGetCliFilePath_ReturnsNilWhenNotFound(t *testing.T) {
	// Save current PATH and restore after test
	oldPath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", oldPath) }()

	// Set PATH to an empty directory
	tmpDir := t.TempDir()
	_ = os.Setenv("PATH", tmpDir)

	// Save current directory and change to temp dir
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Should not find ffmpeg
	result := TryGetCliFilePath()
	if result != nil {
		t.Errorf("Expected nil when ffmpeg not found, got %s", *result)
	}
}

func TestIsAvailable(t *testing.T) {
	// Create a temp directory with a fake ffmpeg
	tmpDir := t.TempDir()
	ffmpegPath := filepath.Join(tmpDir, cliFileName())
	if err := os.WriteFile(ffmpegPath, []byte("fake ffmpeg"), 0o755); err != nil {
		t.Fatalf("Failed to create fake ffmpeg: %v", err)
	}

	// Save current PATH and restore after test
	oldPath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", oldPath) }()

	// Set PATH to our temp directory
	var sep string
	if runtime.GOOS == "windows" {
		sep = ";"
	} else {
		sep = ":"
	}
	_ = os.Setenv("PATH", tmpDir+sep+oldPath)

	// Should be available
	if !IsAvailable() {
		t.Error("Expected IsAvailable() to return true")
	}
}

func TestIsAvailable_ReturnsFalseWhenNotFound(t *testing.T) {
	// Save current PATH and restore after test
	oldPath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", oldPath) }()

	// Set PATH to an empty directory
	tmpDir := t.TempDir()
	_ = os.Setenv("PATH", tmpDir)

	// Save current directory and change to temp dir
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Should not be available
	if IsAvailable() {
		t.Error("Expected IsAvailable() to return false when ffmpeg not found")
	}
}

func TestGetCliFilePath_ReturnsErrorWhenNotFound(t *testing.T) {
	// Save current PATH and restore after test
	oldPath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", oldPath) }()

	// Set PATH to an empty directory
	tmpDir := t.TempDir()
	_ = os.Setenv("PATH", tmpDir)

	// Save current directory and change to temp dir
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Should return error
	_, err = GetCliFilePath()
	if err == nil {
		t.Error("Expected error when ffmpeg not found")
	}
}

func TestGetCliFilePath_ReturnsPathWhenFound(t *testing.T) {
	// Create a temp directory with a fake ffmpeg
	tmpDir := t.TempDir()
	ffmpegPath := filepath.Join(tmpDir, cliFileName())
	if err := os.WriteFile(ffmpegPath, []byte("fake ffmpeg"), 0o755); err != nil {
		t.Fatalf("Failed to create fake ffmpeg: %v", err)
	}

	// Save current PATH and restore after test
	oldPath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", oldPath) }()

	// Set PATH to our temp directory
	var sep string
	if runtime.GOOS == "windows" {
		sep = ";"
	} else {
		sep = ":"
	}
	_ = os.Setenv("PATH", tmpDir+sep+oldPath)

	// Should return path without error
	path, err := GetCliFilePath()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if path != ffmpegPath {
		t.Errorf("Expected %s, got %s", ffmpegPath, path)
	}
}

func TestBuildMuxArgs(t *testing.T) {
	tests := []struct {
		name       string
		videoPath  string
		audioPath  string
		outputPath string
		wantArgs   []string
	}{
		{
			name:       "basic mux",
			videoPath:  "video.mp4",
			audioPath:  "audio.m4a",
			outputPath: "output.mp4",
			wantArgs:   []string{"-i", "video.mp4", "-i", "audio.m4a", "-c", "copy", "-y", "output.mp4"},
		},
		{
			name:       "paths with spaces",
			videoPath:  "my video.mp4",
			audioPath:  "my audio.m4a",
			outputPath: "my output.mp4",
			wantArgs:   []string{"-i", "my video.mp4", "-i", "my audio.m4a", "-c", "copy", "-y", "my output.mp4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := buildMuxArgs(tt.videoPath, tt.audioPath, tt.outputPath)
			if len(args) != len(tt.wantArgs) {
				t.Errorf("buildMuxArgs() = %v, want %v", args, tt.wantArgs)
				return
			}
			for i, arg := range args {
				if arg != tt.wantArgs[i] {
					t.Errorf("buildMuxArgs()[%d] = %v, want %v", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestMuxStreams_ReturnsErrorWhenFFmpegNotFound(t *testing.T) {
	// Save current PATH and restore after test
	oldPath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", oldPath) }()

	// Set PATH to an empty directory
	tmpDir := t.TempDir()
	_ = os.Setenv("PATH", tmpDir)

	// Save current directory and change to temp dir
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Should return ErrNotFound
	err = MuxStreams("video.mp4", "audio.m4a", "output.mp4")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestMuxStreams_ReturnsErrorForMissingInputFiles(t *testing.T) {
	// Skip if ffmpeg not available
	if !IsAvailable() {
		t.Skip("FFmpeg not available")
	}

	tmpDir := t.TempDir()
	videoPath := filepath.Join(tmpDir, "nonexistent_video.mp4")
	audioPath := filepath.Join(tmpDir, "nonexistent_audio.m4a")
	outputPath := filepath.Join(tmpDir, "output.mp4")

	// Should return error for missing input files
	err := MuxStreams(videoPath, audioPath, outputPath)
	if err == nil {
		t.Error("Expected error for missing input files")
	}
}
