package main

import (
	"bytes"
	"errors"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/SakuraBurst/golang-youtube-downloader/pkg/ffmpeg"
	"github.com/SakuraBurst/golang-youtube-downloader/pkg/youtube"
)

func TestWrapErrorInvalidVideoID(t *testing.T) {
	err := WrapError(youtube.ErrInvalidVideoID)

	var userErr *UserFriendlyError
	if !errors.As(err, &userErr) {
		t.Fatal("expected UserFriendlyError")
	}

	if !strings.Contains(userErr.Message, "Invalid video") {
		t.Errorf("message should mention invalid video, got: %s", userErr.Message)
	}

	if userErr.Suggestion == "" {
		t.Error("suggestion should not be empty")
	}
}

func TestWrapErrorInvalidPlaylistID(t *testing.T) {
	err := WrapError(youtube.ErrInvalidPlaylistID)

	var userErr *UserFriendlyError
	if !errors.As(err, &userErr) {
		t.Fatal("expected UserFriendlyError")
	}

	if !strings.Contains(userErr.Message, "Invalid playlist") {
		t.Errorf("message should mention invalid playlist, got: %s", userErr.Message)
	}
}

func TestWrapErrorFFmpegNotFound(t *testing.T) {
	err := WrapError(ffmpeg.ErrNotFound)

	var userErr *UserFriendlyError
	if !errors.As(err, &userErr) {
		t.Fatal("expected UserFriendlyError")
	}

	if !strings.Contains(userErr.Message, "FFmpeg") {
		t.Errorf("message should mention FFmpeg, got: %s", userErr.Message)
	}

	if !strings.Contains(userErr.Suggestion, "install") {
		t.Errorf("suggestion should mention installing FFmpeg, got: %s", userErr.Suggestion)
	}
}

func TestWrapErrorPermissionDenied(t *testing.T) {
	err := WrapError(os.ErrPermission)

	var userErr *UserFriendlyError
	if !errors.As(err, &userErr) {
		t.Fatal("expected UserFriendlyError")
	}

	if !strings.Contains(userErr.Message, "Permission") {
		t.Errorf("message should mention permission, got: %s", userErr.Message)
	}
}

func TestWrapErrorRateLimit(t *testing.T) {
	err := WrapError(errors.New("HTTP 429 Too Many Requests"))

	var userErr *UserFriendlyError
	if !errors.As(err, &userErr) {
		t.Fatal("expected UserFriendlyError")
	}

	if !strings.Contains(userErr.Message, "rate") {
		t.Errorf("message should mention rate limiting, got: %s", userErr.Message)
	}
}

func TestWrapErrorVideoUnavailable(t *testing.T) {
	err := WrapError(errors.New("video unavailable: private"))

	var userErr *UserFriendlyError
	if !errors.As(err, &userErr) {
		t.Fatal("expected UserFriendlyError")
	}

	if !strings.Contains(userErr.Message, "unavailable") {
		t.Errorf("message should mention unavailable, got: %s", userErr.Message)
	}
}

func TestWrapErrorUnknown(t *testing.T) {
	originalErr := errors.New("some random error")
	err := WrapError(originalErr)

	// Unknown errors should pass through unchanged
	if err != originalErr {
		t.Errorf("unknown error should pass through unchanged")
	}
}

func TestWrapErrorNil(t *testing.T) {
	err := WrapError(nil)
	if err != nil {
		t.Error("nil error should return nil")
	}
}

func TestUserFriendlyErrorFormat(t *testing.T) {
	err := &UserFriendlyError{
		Message:    "Something went wrong",
		Suggestion: "Try again later",
		Cause:      errors.New("underlying error"),
	}

	formatted := err.FormatUserError()

	if !strings.Contains(formatted, "Something went wrong") {
		t.Error("formatted error should contain message")
	}

	if !strings.Contains(formatted, "Try again later") {
		t.Error("formatted error should contain suggestion")
	}
}

func TestPrintError(t *testing.T) {
	buf := new(bytes.Buffer)

	userErr := &UserFriendlyError{
		Message:    "Test error",
		Suggestion: "Do something",
	}

	PrintError(buf, userErr)
	output := buf.String()

	if !strings.Contains(output, "Test error") {
		t.Errorf("output should contain error message, got: %s", output)
	}

	if !strings.Contains(output, "Do something") {
		t.Errorf("output should contain suggestion, got: %s", output)
	}
}

func TestPrintErrorNil(t *testing.T) {
	buf := new(bytes.Buffer)
	PrintError(buf, nil)

	if buf.Len() > 0 {
		t.Error("nil error should not produce output")
	}
}

func TestPrintErrorRegular(t *testing.T) {
	buf := new(bytes.Buffer)
	PrintError(buf, errors.New("regular error"))

	if !strings.Contains(buf.String(), "regular error") {
		t.Error("regular error should be printed")
	}
}

// mockNetError implements net.Error for testing
type mockNetError struct {
	timeout bool
}

func (e *mockNetError) Error() string   { return "network error" }
func (e *mockNetError) Timeout() bool   { return e.timeout }
func (e *mockNetError) Temporary() bool { return false }

// Ensure mockNetError implements net.Error
var _ net.Error = (*mockNetError)(nil)

func TestWrapErrorNetworkTimeout(t *testing.T) {
	err := WrapError(&mockNetError{timeout: true})

	var userErr *UserFriendlyError
	if !errors.As(err, &userErr) {
		t.Fatal("expected UserFriendlyError")
	}

	if !strings.Contains(userErr.Message, "timed out") {
		t.Errorf("message should mention timeout, got: %s", userErr.Message)
	}
}
