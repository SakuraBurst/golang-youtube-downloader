package youtube

import (
	"testing"
)

func TestStreamInfo_HasRequiredFields(t *testing.T) {
	stream := StreamInfo{
		URL:       "https://example.com/stream",
		Quality:   "1080p",
		Bitrate:   5000000,
		Codec:     "avc1.640028",
		Container: "mp4",
		Size:      50000000,
	}

	if stream.URL == "" {
		t.Error("URL should be set")
	}
	if stream.Quality == "" {
		t.Error("Quality should be set")
	}
	if stream.Bitrate == 0 {
		t.Error("Bitrate should be set")
	}
	if stream.Codec == "" {
		t.Error("Codec should be set")
	}
	if stream.Container == "" {
		t.Error("Container should be set")
	}
}

func TestVideoStreamInfo_HasVideoFields(t *testing.T) {
	stream := VideoStreamInfo{
		StreamInfo: StreamInfo{
			URL:       "https://example.com/video",
			Quality:   "1080p",
			Bitrate:   5000000,
			Codec:     "avc1.640028",
			Container: "mp4",
		},
		Width:      1920,
		Height:     1080,
		Framerate:  30,
		VideoCodec: "avc1.640028",
	}

	if stream.Width == 0 {
		t.Error("Width should be set")
	}
	if stream.Height == 0 {
		t.Error("Height should be set")
	}
	if stream.Framerate == 0 {
		t.Error("Framerate should be set")
	}
}

func TestAudioStreamInfo_HasAudioFields(t *testing.T) {
	stream := AudioStreamInfo{
		StreamInfo: StreamInfo{
			URL:       "https://example.com/audio",
			Quality:   "128kbps",
			Bitrate:   128000,
			Codec:     "mp4a.40.2",
			Container: "mp4",
		},
		AudioCodec:   "mp4a.40.2",
		SampleRate:   44100,
		ChannelCount: 2,
	}

	if stream.AudioCodec == "" {
		t.Error("AudioCodec should be set")
	}
	if stream.SampleRate == 0 {
		t.Error("SampleRate should be set")
	}
	if stream.ChannelCount == 0 {
		t.Error("ChannelCount should be set")
	}
}

func TestStreamInfo_IsVideoOnly(t *testing.T) {
	videoOnly := &VideoStreamInfo{
		StreamInfo: StreamInfo{
			Quality: "1080p",
		},
		Width:  1920,
		Height: 1080,
	}

	if !videoOnly.IsVideoOnly() {
		t.Error("should be video only")
	}
}

func TestStreamInfo_IsAudioOnly(t *testing.T) {
	audioOnly := &AudioStreamInfo{
		StreamInfo: StreamInfo{
			Quality: "128kbps",
		},
	}

	if !audioOnly.IsAudioOnly() {
		t.Error("should be audio only")
	}
}

func TestMuxedStreamInfo_HasBoth(t *testing.T) {
	muxed := MuxedStreamInfo{
		VideoStreamInfo: VideoStreamInfo{
			StreamInfo: StreamInfo{
				URL:       "https://example.com/muxed",
				Quality:   "720p",
				Container: "mp4",
			},
			Width:  1280,
			Height: 720,
		},
		AudioStreamInfo: AudioStreamInfo{
			StreamInfo: StreamInfo{
				Codec: "mp4a.40.2",
			},
			AudioCodec: "mp4a.40.2",
		},
	}

	if muxed.Width == 0 {
		t.Error("should have video width")
	}
	if muxed.AudioCodec == "" {
		t.Error("should have audio codec")
	}
}

func TestQualityLabel_Standard(t *testing.T) {
	tests := []struct {
		height   int
		expected string
	}{
		{2160, "4K"},
		{1440, "1440p"},
		{1080, "1080p"},
		{720, "720p"},
		{480, "480p"},
		{360, "360p"},
		{240, "240p"},
		{144, "144p"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			label := QualityLabel(tt.height)
			if label != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, label)
			}
		})
	}
}

func TestContainer_CommonTypes(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"mp4", true},
		{"webm", true},
		{"mp3", true},
		{"ogg", true},
		{"mkv", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Container(tt.name)
			if tt.valid && c == "" {
				t.Error("container should be set")
			}
		})
	}
}
