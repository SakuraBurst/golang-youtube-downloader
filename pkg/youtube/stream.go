package youtube

import "fmt"

// Container represents a media container format (e.g., mp4, webm).
type Container string

// Common container types.
const (
	ContainerMP4  Container = "mp4"
	ContainerWebM Container = "webm"
	ContainerMP3  Container = "mp3"
	ContainerOGG  Container = "ogg"
	ContainerMKV  Container = "mkv"
	Container3GP  Container = "3gp"
)

// StreamInfo contains common information about a media stream.
type StreamInfo struct {
	// URL is the direct URL to download the stream.
	URL string

	// Quality is a human-readable quality label (e.g., "1080p", "128kbps").
	Quality string

	// Bitrate is the stream's bitrate in bits per second.
	Bitrate int64

	// Codec is the codec identifier (e.g., "avc1.640028", "mp4a.40.2").
	Codec string

	// Container is the media container format.
	Container Container

	// Size is the content length in bytes (may be 0 if unknown).
	Size int64

	// MimeType is the MIME type of the stream.
	MimeType string

	// ContentLength is the content length in bytes.
	ContentLength int64
}

// VideoStreamInfo contains information about a video-only stream.
type VideoStreamInfo struct {
	StreamInfo

	// Width is the video width in pixels.
	Width int

	// Height is the video height in pixels.
	Height int

	// Framerate is the video framerate (frames per second).
	Framerate int

	// VideoCodec is the video codec (e.g., "avc1.640028", "vp9").
	VideoCodec string
}

// IsVideoOnly returns true (video streams are video-only by definition).
func (v *VideoStreamInfo) IsVideoOnly() bool {
	return true
}

// AudioStreamInfo contains information about an audio-only stream.
type AudioStreamInfo struct {
	StreamInfo

	// AudioCodec is the audio codec (e.g., "mp4a.40.2", "opus").
	AudioCodec string

	// SampleRate is the audio sample rate in Hz.
	SampleRate int

	// ChannelCount is the number of audio channels.
	ChannelCount int

	// AudioLanguage is the language of the audio track (may be empty).
	AudioLanguage string

	// IsDefault indicates if this is the default audio track.
	IsDefault bool
}

// IsAudioOnly returns true (audio streams are audio-only by definition).
func (a *AudioStreamInfo) IsAudioOnly() bool {
	return true
}

// MuxedStreamInfo contains information about a muxed stream (video + audio).
type MuxedStreamInfo struct {
	VideoStreamInfo
	AudioStreamInfo
}

// QualityLabel returns a human-readable quality label for a given video height.
func QualityLabel(height int) string {
	switch {
	case height >= 2160:
		return "4K"
	case height >= 1440:
		return "1440p"
	case height >= 1080:
		return "1080p"
	case height >= 720:
		return "720p"
	case height >= 480:
		return "480p"
	case height >= 360:
		return "360p"
	case height >= 240:
		return "240p"
	default:
		return fmt.Sprintf("%dp", height)
	}
}

// StreamManifest contains all available streams for a video.
type StreamManifest struct {
	// VideoStreams contains all video-only streams.
	VideoStreams []VideoStreamInfo

	// AudioStreams contains all audio-only streams.
	AudioStreams []AudioStreamInfo

	// MuxedStreams contains all muxed (video+audio) streams.
	MuxedStreams []MuxedStreamInfo
}

// GetBestVideoStream returns the highest quality video stream.
func (m *StreamManifest) GetBestVideoStream() *VideoStreamInfo {
	if len(m.VideoStreams) == 0 {
		return nil
	}

	best := &m.VideoStreams[0]
	for i := range m.VideoStreams {
		if m.VideoStreams[i].Height > best.Height {
			best = &m.VideoStreams[i]
		}
	}
	return best
}

// GetBestAudioStream returns the highest quality audio stream.
func (m *StreamManifest) GetBestAudioStream() *AudioStreamInfo {
	if len(m.AudioStreams) == 0 {
		return nil
	}

	best := &m.AudioStreams[0]
	for i := range m.AudioStreams {
		if m.AudioStreams[i].Bitrate > best.Bitrate {
			best = &m.AudioStreams[i]
		}
	}
	return best
}

// DownloadOption represents a single download option combining video and/or audio streams.
type DownloadOption struct {
	// Container is the output container format.
	Container Container

	// IsAudioOnly indicates if this option extracts audio only.
	IsAudioOnly bool

	// VideoStream is the video stream for this option (nil for audio-only).
	VideoStream *VideoStreamInfo

	// AudioStream is the audio stream for this option.
	AudioStream *AudioStreamInfo
}

// QualityLabel returns a human-readable label for this download option.
func (o *DownloadOption) QualityLabel() string {
	if o.IsAudioOnly {
		return "Audio"
	}
	if o.VideoStream != nil {
		if o.VideoStream.Quality != "" {
			return o.VideoStream.Quality
		}
		return QualityLabel(o.VideoStream.Height)
	}
	return ""
}

// GetDownloadOptions generates all available download options from the stream manifest.
// It creates video+audio combinations and audio-only options.
func (m *StreamManifest) GetDownloadOptions() []DownloadOption {
	var options []DownloadOption

	// Find the best audio stream for each container type
	bestAudioMP4 := m.findBestAudioByContainer(ContainerMP4)
	bestAudioWebM := m.findBestAudioByContainer(ContainerWebM)

	// Generate video+audio options from adaptive formats
	for i := range m.VideoStreams {
		vs := &m.VideoStreams[i]

		// Find compatible audio stream (prefer same container)
		var audioStream *AudioStreamInfo
		switch {
		case vs.Container == ContainerMP4 && bestAudioMP4 != nil:
			audioStream = bestAudioMP4
		case vs.Container == ContainerWebM && bestAudioWebM != nil:
			audioStream = bestAudioWebM
		default:
			// Fallback to any available audio
			audioStream = m.GetBestAudioStream()
		}

		if audioStream != nil {
			options = append(options, DownloadOption{
				Container:   vs.Container,
				IsAudioOnly: false,
				VideoStream: vs,
				AudioStream: audioStream,
			})
		} else {
			// Video only if no audio available
			options = append(options, DownloadOption{
				Container:   vs.Container,
				IsAudioOnly: false,
				VideoStream: vs,
				AudioStream: nil,
			})
		}
	}

	// Generate video+audio options from muxed streams
	for i := range m.MuxedStreams {
		ms := &m.MuxedStreams[i]
		options = append(options, DownloadOption{
			Container:   ms.VideoStreamInfo.Container,
			IsAudioOnly: false,
			VideoStream: &ms.VideoStreamInfo,
			AudioStream: &ms.AudioStreamInfo,
		})
	}

	// Generate audio-only options
	for i := range m.AudioStreams {
		as := &m.AudioStreams[i]
		options = append(options, DownloadOption{
			Container:   as.Container,
			IsAudioOnly: true,
			VideoStream: nil,
			AudioStream: as,
		})
	}

	return options
}

// findBestAudioByContainer finds the highest bitrate audio stream with the specified container.
func (m *StreamManifest) findBestAudioByContainer(container Container) *AudioStreamInfo {
	var best *AudioStreamInfo
	for i := range m.AudioStreams {
		as := &m.AudioStreams[i]
		if as.Container == container {
			if best == nil || as.Bitrate > best.Bitrate {
				best = as
			}
		}
	}
	return best
}
