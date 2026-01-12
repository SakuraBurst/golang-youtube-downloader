package tagging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SakuraBurst/golang-youtube-downloader/pkg/youtube"
)

func TestTagInjector_InjectTags_SetsBasicMetadata(t *testing.T) {
	// Create a temporary MP3 file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.mp3")

	// Create a minimal valid MP3 file (ID3v2 header + minimal frame data)
	// This is the minimum required for tag libraries to open the file
	mp3Data := createMinimalMP3()
	if err := os.WriteFile(testFile, mp3Data, 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	video := &youtube.Video{
		ID:          "dQw4w9WgXcQ",
		Title:       "Test Video Title",
		Description: "Test description",
		Author: youtube.Author{
			Name:      "Test Channel",
			ChannelID: "UCtest123",
			URL:       "https://www.youtube.com/channel/UCtest123",
		},
	}

	injector := NewTagInjector()
	err := injector.InjectTags(testFile, video)
	if err != nil {
		t.Fatalf("InjectTags failed: %v", err)
	}

	// Verify tags were written by reading them back
	tags, err := ReadTags(testFile)
	if err != nil {
		t.Fatalf("ReadTags failed: %v", err)
	}

	if tags.Title != video.Title {
		t.Errorf("Title mismatch: got %q, want %q", tags.Title, video.Title)
	}

	if tags.Artist != video.Author.Name {
		t.Errorf("Artist mismatch: got %q, want %q", tags.Artist, video.Author.Name)
	}
}

func TestTagInjector_InjectTags_SetsAlbumFromChannelName(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.mp3")

	mp3Data := createMinimalMP3()
	if err := os.WriteFile(testFile, mp3Data, 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	video := &youtube.Video{
		ID:    "abc123",
		Title: "Another Test",
		Author: youtube.Author{
			Name: "Music Channel",
		},
	}

	injector := NewTagInjector()
	err := injector.InjectTags(testFile, video)
	if err != nil {
		t.Fatalf("InjectTags failed: %v", err)
	}

	tags, err := ReadTags(testFile)
	if err != nil {
		t.Fatalf("ReadTags failed: %v", err)
	}

	// Album should be set to the channel name by default
	if tags.Album != video.Author.Name {
		t.Errorf("Album mismatch: got %q, want %q", tags.Album, video.Author.Name)
	}
}

func TestReadTags_ReturnsEmptyForUntaggedFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "untagged.mp3")

	mp3Data := createMinimalMP3()
	if err := os.WriteFile(testFile, mp3Data, 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tags, err := ReadTags(testFile)
	if err != nil {
		t.Fatalf("ReadTags failed: %v", err)
	}

	// For a fresh file, tags should be empty
	if tags.Title != "" {
		t.Errorf("Expected empty title for untagged file, got %q", tags.Title)
	}
}

func TestTagInjector_InjectTags_M4AFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.m4a")

	// Create a minimal valid M4A file
	m4aData := createMinimalM4A()
	if err := os.WriteFile(testFile, m4aData, 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	video := &youtube.Video{
		ID:    "xyz789",
		Title: "M4A Test",
		Author: youtube.Author{
			Name: "M4A Channel",
		},
	}

	injector := NewTagInjector()
	err := injector.InjectTags(testFile, video)
	if err != nil {
		t.Fatalf("InjectTags failed: %v", err)
	}

	tags, err := ReadTags(testFile)
	if err != nil {
		t.Fatalf("ReadTags failed: %v", err)
	}

	if tags.Title != video.Title {
		t.Errorf("Title mismatch: got %q, want %q", tags.Title, video.Title)
	}
}

// createMinimalMP3 creates a minimal valid MP3 file with ID3v2 header.
func createMinimalMP3() []byte {
	// ID3v2.3 header (10 bytes) + padding
	// ID3 marker + version 2.3 + flags + size (syncsafe integer)
	header := []byte{
		'I', 'D', '3', // ID3 marker
		0x03, 0x00, // Version 2.3.0
		0x00,                   // Flags
		0x00, 0x00, 0x00, 0x00, // Size (syncsafe, 0 bytes)
	}

	// Minimal MP3 frame header (valid sync word)
	// 0xFF 0xFB = MPEG Audio Layer 3
	mp3Frame := []byte{
		0xFF, 0xFB, 0x90, 0x00, // MP3 frame header (Layer 3, 128kbps, 44100Hz, stereo)
	}
	// Add some padding to make it look like audio data
	padding := make([]byte, 417) // Typical frame size for 128kbps

	result := make([]byte, 0, len(header)+len(mp3Frame)+len(padding))
	result = append(result, header...)
	result = append(result, mp3Frame...)
	result = append(result, padding...)
	return result
}

// createMinimalM4A creates a minimal valid M4A/MP4 container.
func createMinimalM4A() []byte {
	// Minimal ftyp box + moov box structure
	// This is a very simplified M4A container
	ftyp := []byte{
		0x00, 0x00, 0x00, 0x14, // Box size (20 bytes)
		'f', 't', 'y', 'p', // Box type
		'M', '4', 'A', ' ', // Major brand
		0x00, 0x00, 0x00, 0x00, // Minor version
		'M', '4', 'A', ' ', // Compatible brand
	}

	// Minimal moov box (movie box)
	moov := []byte{
		0x00, 0x00, 0x00, 0x08, // Box size (8 bytes, empty moov)
		'm', 'o', 'o', 'v', // Box type
	}

	return append(ftyp, moov...)
}
