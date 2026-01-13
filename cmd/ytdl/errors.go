package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/SakuraBurst/golang-youtube-downloader/pkg/ffmpeg"
	"github.com/SakuraBurst/golang-youtube-downloader/pkg/youtube"
)

// UserFriendlyError wraps an error with a user-friendly message and suggestion.
type UserFriendlyError struct {
	Message    string
	Suggestion string
	Cause      error
}

func (e *UserFriendlyError) Error() string {
	return e.Message
}

func (e *UserFriendlyError) Unwrap() error {
	return e.Cause
}

// FormatUserError returns a formatted string for display to the user.
func (e *UserFriendlyError) FormatUserError() string {
	var sb strings.Builder
	sb.WriteString("Error: ")
	sb.WriteString(e.Message)
	if e.Suggestion != "" {
		sb.WriteString("\n\nSuggestion: ")
		sb.WriteString(e.Suggestion)
	}
	return sb.String()
}

// WrapError converts common errors into user-friendly messages.
func WrapError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific YouTube errors
	if errors.Is(err, youtube.ErrInvalidVideoID) {
		return &UserFriendlyError{
			Message:    "Invalid video URL or ID",
			Suggestion: "Make sure you're using a valid YouTube URL like:\n  - https://www.youtube.com/watch?v=VIDEO_ID\n  - https://youtu.be/VIDEO_ID\n  - Or just the 11-character video ID",
			Cause:      err,
		}
	}

	if errors.Is(err, youtube.ErrInvalidPlaylistID) {
		return &UserFriendlyError{
			Message:    "Invalid playlist URL or ID",
			Suggestion: "Make sure you're using a valid YouTube playlist URL like:\n  - https://www.youtube.com/playlist?list=PLAYLIST_ID",
			Cause:      err,
		}
	}

	if errors.Is(err, youtube.ErrInvalidChannelID) {
		return &UserFriendlyError{
			Message:    "Invalid channel URL or ID",
			Suggestion: "Make sure you're using a valid YouTube channel URL like:\n  - https://www.youtube.com/channel/CHANNEL_ID\n  - https://www.youtube.com/@handle",
			Cause:      err,
		}
	}

	if errors.Is(err, youtube.ErrUnresolvableQuery) {
		return &UserFriendlyError{
			Message:    "Unable to recognize the URL or ID",
			Suggestion: "Check that the URL is a valid YouTube video, playlist, or channel URL",
			Cause:      err,
		}
	}

	// Check for FFmpeg errors
	if errors.Is(err, ffmpeg.ErrNotFound) {
		return &UserFriendlyError{
			Message:    "FFmpeg not found",
			Suggestion: "FFmpeg is required for muxing video and audio streams.\nPlease install FFmpeg and make sure it's in your PATH.\nDownload from: https://ffmpeg.org/download.html",
			Cause:      err,
		}
	}

	// Check for network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return &UserFriendlyError{
				Message:    "Connection timed out",
				Suggestion: "Check your internet connection and try again",
				Cause:      err,
			}
		}
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return &UserFriendlyError{
				Message:    "Request timed out",
				Suggestion: "The server took too long to respond. Try again later",
				Cause:      err,
			}
		}
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return &UserFriendlyError{
			Message:    "Could not resolve host",
			Suggestion: "Check your internet connection and DNS settings",
			Cause:      err,
		}
	}

	// Check for I/O and filesystem errors
	if errors.Is(err, os.ErrPermission) {
		return &UserFriendlyError{
			Message:    "Permission denied",
			Suggestion: "Check that you have write permissions to the output directory",
			Cause:      err,
		}
	}

	if errors.Is(err, os.ErrNotExist) {
		return &UserFriendlyError{
			Message:    "File or directory not found",
			Suggestion: "Make sure the output directory exists",
			Cause:      err,
		}
	}

	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		if errors.Is(pathErr.Err, syscall.ENOSPC) {
			return &UserFriendlyError{
				Message:    "No space left on device",
				Suggestion: "Free up some disk space and try again",
				Cause:      err,
			}
		}
	}

	// Check for video unavailable errors
	errStr := err.Error()
	if strings.Contains(errStr, "unavailable") {
		return &UserFriendlyError{
			Message:    "Video is unavailable",
			Suggestion: "The video may be:\n  - Private or deleted\n  - Age-restricted\n  - Blocked in your region\n  - Requires sign-in",
			Cause:      err,
		}
	}

	// Check for rate limiting
	if strings.Contains(errStr, "429") || strings.Contains(strings.ToLower(errStr), "rate limit") {
		return &UserFriendlyError{
			Message:    "Too many requests - rate limited by YouTube",
			Suggestion: "Wait a few minutes before trying again",
			Cause:      err,
		}
	}

	// Check for HTTP errors
	if strings.Contains(errStr, "403") {
		return &UserFriendlyError{
			Message:    "Access forbidden (HTTP 403)",
			Suggestion: "The content may be restricted or your IP may be blocked",
			Cause:      err,
		}
	}

	if strings.Contains(errStr, "404") {
		return &UserFriendlyError{
			Message:    "Content not found (HTTP 404)",
			Suggestion: "The video, playlist, or channel may have been deleted",
			Cause:      err,
		}
	}

	// Return original error if no specific handling
	return err
}

// PrintError prints an error in a user-friendly format.
func PrintError(w io.Writer, err error) {
	if err == nil {
		return
	}

	var userErr *UserFriendlyError
	if errors.As(err, &userErr) {
		_, _ = fmt.Fprintln(w, userErr.FormatUserError())
	} else {
		_, _ = fmt.Fprintf(w, "Error: %v\n", err)
	}
}
