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

func TestStreamingDataResponse_GetStreamManifest_Basic(t *testing.T) {
	sd := &StreamingDataResponse{
		AdaptiveFormats: []FormatResponse{
			{
				Itag:         137,
				URL:          "https://example.com/video137",
				MimeType:     "video/mp4; codecs=\"avc1.640028\"",
				Bitrate:      5000000,
				Width:        1920,
				Height:       1080,
				Fps:          30,
				QualityLabel: "1080p",
			},
			{
				Itag:            140,
				URL:             "https://example.com/audio140",
				MimeType:        "audio/mp4; codecs=\"mp4a.40.2\"",
				Bitrate:         128000,
				AudioQuality:    "AUDIO_QUALITY_MEDIUM",
				AudioSampleRate: "44100",
				AudioChannels:   2,
			},
		},
	}

	manifest := sd.GetStreamManifest()
	if manifest == nil {
		t.Fatal("expected manifest to be non-nil")
	}

	if len(manifest.VideoStreams) != 1 {
		t.Errorf("expected 1 video stream, got %d", len(manifest.VideoStreams))
	}
	if len(manifest.AudioStreams) != 1 {
		t.Errorf("expected 1 audio stream, got %d", len(manifest.AudioStreams))
	}
}

func TestStreamingDataResponse_GetStreamManifest_VideoStream(t *testing.T) {
	sd := &StreamingDataResponse{
		AdaptiveFormats: []FormatResponse{
			{
				Itag:          137,
				URL:           "https://example.com/video137",
				MimeType:      "video/mp4; codecs=\"avc1.640028\"",
				Bitrate:       5000000,
				Width:         1920,
				Height:        1080,
				Fps:           30,
				QualityLabel:  "1080p",
				ContentLength: "50000000",
			},
		},
	}

	manifest := sd.GetStreamManifest()
	if len(manifest.VideoStreams) != 1 {
		t.Fatalf("expected 1 video stream, got %d", len(manifest.VideoStreams))
	}

	vs := manifest.VideoStreams[0]
	if vs.URL != "https://example.com/video137" {
		t.Errorf("expected URL, got %q", vs.URL)
	}
	if vs.Width != 1920 {
		t.Errorf("expected width 1920, got %d", vs.Width)
	}
	if vs.Height != 1080 {
		t.Errorf("expected height 1080, got %d", vs.Height)
	}
	if vs.Framerate != 30 {
		t.Errorf("expected framerate 30, got %d", vs.Framerate)
	}
	if vs.Bitrate != 5000000 {
		t.Errorf("expected bitrate 5000000, got %d", vs.Bitrate)
	}
	if vs.Quality != "1080p" {
		t.Errorf("expected quality 1080p, got %q", vs.Quality)
	}
	if vs.Container != ContainerMP4 {
		t.Errorf("expected container mp4, got %q", vs.Container)
	}
	if vs.VideoCodec != "avc1.640028" {
		t.Errorf("expected video codec avc1.640028, got %q", vs.VideoCodec)
	}
	if vs.ContentLength != 50000000 {
		t.Errorf("expected content length 50000000, got %d", vs.ContentLength)
	}
}

func TestStreamingDataResponse_GetStreamManifest_AudioStream(t *testing.T) {
	sd := &StreamingDataResponse{
		AdaptiveFormats: []FormatResponse{
			{
				Itag:            140,
				URL:             "https://example.com/audio140",
				MimeType:        "audio/mp4; codecs=\"mp4a.40.2\"",
				Bitrate:         128000,
				AudioQuality:    "AUDIO_QUALITY_MEDIUM",
				AudioSampleRate: "44100",
				AudioChannels:   2,
				ContentLength:   "5000000",
			},
		},
	}

	manifest := sd.GetStreamManifest()
	if len(manifest.AudioStreams) != 1 {
		t.Fatalf("expected 1 audio stream, got %d", len(manifest.AudioStreams))
	}

	as := manifest.AudioStreams[0]
	if as.URL != "https://example.com/audio140" {
		t.Errorf("expected URL, got %q", as.URL)
	}
	if as.Bitrate != 128000 {
		t.Errorf("expected bitrate 128000, got %d", as.Bitrate)
	}
	if as.AudioCodec != "mp4a.40.2" {
		t.Errorf("expected audio codec mp4a.40.2, got %q", as.AudioCodec)
	}
	if as.SampleRate != 44100 {
		t.Errorf("expected sample rate 44100, got %d", as.SampleRate)
	}
	if as.ChannelCount != 2 {
		t.Errorf("expected channel count 2, got %d", as.ChannelCount)
	}
	if as.Container != ContainerMP4 {
		t.Errorf("expected container mp4, got %q", as.Container)
	}
}

func TestStreamingDataResponse_GetStreamManifest_WebMContainer(t *testing.T) {
	sd := &StreamingDataResponse{
		AdaptiveFormats: []FormatResponse{
			{
				Itag:     248,
				URL:      "https://example.com/video248",
				MimeType: "video/webm; codecs=\"vp9\"",
				Bitrate:  4000000,
				Width:    1920,
				Height:   1080,
			},
		},
	}

	manifest := sd.GetStreamManifest()
	if len(manifest.VideoStreams) != 1 {
		t.Fatalf("expected 1 video stream, got %d", len(manifest.VideoStreams))
	}

	vs := manifest.VideoStreams[0]
	if vs.Container != ContainerWebM {
		t.Errorf("expected container webm, got %q", vs.Container)
	}
	if vs.VideoCodec != "vp9" {
		t.Errorf("expected video codec vp9, got %q", vs.VideoCodec)
	}
}

func TestStreamingDataResponse_GetStreamManifest_MuxedFormats(t *testing.T) {
	sd := &StreamingDataResponse{
		Formats: []FormatResponse{
			{
				Itag:         18,
				URL:          "https://example.com/muxed18",
				MimeType:     "video/mp4; codecs=\"avc1.42001E, mp4a.40.2\"",
				Bitrate:      500000,
				Width:        640,
				Height:       360,
				QualityLabel: "360p",
			},
		},
	}

	manifest := sd.GetStreamManifest()
	if len(manifest.MuxedStreams) != 1 {
		t.Fatalf("expected 1 muxed stream, got %d", len(manifest.MuxedStreams))
	}

	ms := manifest.MuxedStreams[0]
	if ms.Width != 640 {
		t.Errorf("expected width 640, got %d", ms.Width)
	}
	if ms.Height != 360 {
		t.Errorf("expected height 360, got %d", ms.Height)
	}
}

func TestStreamingDataResponse_GetStreamManifest_EmptyFormats(t *testing.T) {
	sd := &StreamingDataResponse{
		AdaptiveFormats: []FormatResponse{},
		Formats:         []FormatResponse{},
	}

	manifest := sd.GetStreamManifest()
	if manifest == nil {
		t.Fatal("expected manifest to be non-nil even with empty formats")
	}
	if len(manifest.VideoStreams) != 0 {
		t.Errorf("expected 0 video streams, got %d", len(manifest.VideoStreams))
	}
	if len(manifest.AudioStreams) != 0 {
		t.Errorf("expected 0 audio streams, got %d", len(manifest.AudioStreams))
	}
	if len(manifest.MuxedStreams) != 0 {
		t.Errorf("expected 0 muxed streams, got %d", len(manifest.MuxedStreams))
	}
}

func TestStreamingDataResponse_GetStreamManifest_MultipleStreams(t *testing.T) {
	sd := &StreamingDataResponse{
		AdaptiveFormats: []FormatResponse{
			{Itag: 137, MimeType: "video/mp4; codecs=\"avc1.640028\"", Width: 1920, Height: 1080, Bitrate: 5000000},
			{Itag: 136, MimeType: "video/mp4; codecs=\"avc1.4d401f\"", Width: 1280, Height: 720, Bitrate: 2500000},
			{Itag: 140, MimeType: "audio/mp4; codecs=\"mp4a.40.2\"", Bitrate: 128000, AudioSampleRate: "44100"},
			{Itag: 251, MimeType: "audio/webm; codecs=\"opus\"", Bitrate: 160000, AudioSampleRate: "48000"},
		},
	}

	manifest := sd.GetStreamManifest()
	if len(manifest.VideoStreams) != 2 {
		t.Errorf("expected 2 video streams, got %d", len(manifest.VideoStreams))
	}
	if len(manifest.AudioStreams) != 2 {
		t.Errorf("expected 2 audio streams, got %d", len(manifest.AudioStreams))
	}
}

func TestStreamingDataResponse_GetStreamManifest_VideoOnlyNoAudio(t *testing.T) {
	sd := &StreamingDataResponse{
		AdaptiveFormats: []FormatResponse{
			{Itag: 137, MimeType: "video/mp4; codecs=\"avc1.640028\"", Width: 1920, Height: 1080, Bitrate: 5000000},
			{Itag: 248, MimeType: "video/webm; codecs=\"vp9\"", Width: 1920, Height: 1080, Bitrate: 4000000},
		},
	}

	manifest := sd.GetStreamManifest()
	if len(manifest.VideoStreams) != 2 {
		t.Errorf("expected 2 video streams, got %d", len(manifest.VideoStreams))
	}
	if len(manifest.AudioStreams) != 0 {
		t.Errorf("expected 0 audio streams, got %d", len(manifest.AudioStreams))
	}
}

func TestStreamingDataResponse_GetStreamManifest_AudioOnlyNoVideo(t *testing.T) {
	sd := &StreamingDataResponse{
		AdaptiveFormats: []FormatResponse{
			{Itag: 140, MimeType: "audio/mp4; codecs=\"mp4a.40.2\"", Bitrate: 128000, AudioSampleRate: "44100"},
			{Itag: 251, MimeType: "audio/webm; codecs=\"opus\"", Bitrate: 160000, AudioSampleRate: "48000"},
		},
	}

	manifest := sd.GetStreamManifest()
	if len(manifest.VideoStreams) != 0 {
		t.Errorf("expected 0 video streams, got %d", len(manifest.VideoStreams))
	}
	if len(manifest.AudioStreams) != 2 {
		t.Errorf("expected 2 audio streams, got %d", len(manifest.AudioStreams))
	}
}

func TestStreamingDataResponse_GetStreamManifest_CorrectSeparation(t *testing.T) {
	// Verify that video streams have video properties and audio streams have audio properties
	sd := &StreamingDataResponse{
		AdaptiveFormats: []FormatResponse{
			{Itag: 137, MimeType: "video/mp4; codecs=\"avc1.640028\"", Width: 1920, Height: 1080, Bitrate: 5000000, Fps: 30},
			{Itag: 140, MimeType: "audio/mp4; codecs=\"mp4a.40.2\"", Bitrate: 128000, AudioSampleRate: "44100", AudioChannels: 2},
		},
	}

	manifest := sd.GetStreamManifest()

	// Video streams should have width/height/fps
	for _, vs := range manifest.VideoStreams {
		if vs.Width == 0 || vs.Height == 0 {
			t.Error("video stream should have width and height")
		}
	}

	// Audio streams should have sample rate
	for _, as := range manifest.AudioStreams {
		if as.SampleRate == 0 {
			t.Error("audio stream should have sample rate")
		}
	}
}
