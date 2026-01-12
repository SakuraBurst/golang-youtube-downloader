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

func TestParseSignatureCipher_ValidCipher(t *testing.T) {
	// Example signatureCipher format: s=encrypted_sig&sp=sig&url=actual_url
	cipher := "s=ABC123XYZ&sp=sig&url=https%3A%2F%2Fexample.com%2Fvideo"

	parsed, err := ParseSignatureCipher(cipher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.Signature != "ABC123XYZ" {
		t.Errorf("expected signature %q, got %q", "ABC123XYZ", parsed.Signature)
	}
	if parsed.SignatureParam != "sig" {
		t.Errorf("expected signature param %q, got %q", "sig", parsed.SignatureParam)
	}
	if parsed.URL != "https://example.com/video" {
		t.Errorf("expected URL %q, got %q", "https://example.com/video", parsed.URL)
	}
}

func TestParseSignatureCipher_DefaultSignatureParam(t *testing.T) {
	// Without sp parameter, defaults to "signature"
	cipher := "s=ABC123XYZ&url=https%3A%2F%2Fexample.com%2Fvideo"

	parsed, err := ParseSignatureCipher(cipher)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.SignatureParam != "signature" {
		t.Errorf("expected default signature param %q, got %q", "signature", parsed.SignatureParam)
	}
}

func TestParseSignatureCipher_MissingSignature(t *testing.T) {
	cipher := "url=https%3A%2F%2Fexample.com%2Fvideo"

	_, err := ParseSignatureCipher(cipher)
	if err == nil {
		t.Error("expected error for missing signature")
	}
}

func TestParseSignatureCipher_MissingURL(t *testing.T) {
	cipher := "s=ABC123XYZ&sp=sig"

	_, err := ParseSignatureCipher(cipher)
	if err == nil {
		t.Error("expected error for missing URL")
	}
}

func TestParseSignatureCipher_EmptyString(t *testing.T) {
	_, err := ParseSignatureCipher("")
	if err == nil {
		t.Error("expected error for empty cipher string")
	}
}

func TestFormatResponse_NeedsCipherDecryption(t *testing.T) {
	// Stream with direct URL - no decryption needed
	directStream := FormatResponse{
		Itag:     137,
		URL:      "https://example.com/video",
		MimeType: "video/mp4; codecs=\"avc1.640028\"",
	}
	if directStream.NeedsCipherDecryption() {
		t.Error("stream with direct URL should not need cipher decryption")
	}

	// Stream with signatureCipher - needs decryption
	cipherStream := FormatResponse{
		Itag:            137,
		MimeType:        "video/mp4; codecs=\"avc1.640028\"",
		SignatureCipher: "s=ABC123XYZ&sp=sig&url=https%3A%2F%2Fexample.com%2Fvideo",
	}
	if !cipherStream.NeedsCipherDecryption() {
		t.Error("stream with signatureCipher should need cipher decryption")
	}
}

func TestSignatureCipher_BuildURL(t *testing.T) {
	cipher := &SignatureCipher{
		URL:            "https://example.com/video",
		SignatureParam: "sig",
		Signature:      "decrypted_sig",
	}

	url := cipher.BuildURL()
	expected := "https://example.com/video&sig=decrypted_sig"
	if url != expected {
		t.Errorf("expected URL %q, got %q", expected, url)
	}
}

func TestDownloadOption_Basic(t *testing.T) {
	opt := DownloadOption{
		Container:   ContainerMP4,
		IsAudioOnly: false,
		VideoStream: &VideoStreamInfo{
			StreamInfo: StreamInfo{Quality: "1080p", Bitrate: 5000000},
			Width:      1920,
			Height:     1080,
		},
		AudioStream: &AudioStreamInfo{
			StreamInfo: StreamInfo{Quality: "AUDIO_QUALITY_MEDIUM", Bitrate: 128000},
		},
	}

	if opt.Container != ContainerMP4 {
		t.Errorf("expected container mp4, got %v", opt.Container)
	}
	if opt.IsAudioOnly {
		t.Error("expected IsAudioOnly to be false")
	}
	if opt.VideoStream == nil {
		t.Error("expected video stream to be set")
	}
	if opt.AudioStream == nil {
		t.Error("expected audio stream to be set")
	}
}

func TestDownloadOption_QualityLabel(t *testing.T) {
	opt := DownloadOption{
		Container:   ContainerMP4,
		IsAudioOnly: false,
		VideoStream: &VideoStreamInfo{
			StreamInfo: StreamInfo{Quality: "1080p", Bitrate: 5000000},
			Width:      1920,
			Height:     1080,
		},
	}

	label := opt.QualityLabel()
	if label != "1080p" {
		t.Errorf("expected quality label '1080p', got %q", label)
	}
}

func TestDownloadOption_QualityLabel_AudioOnly(t *testing.T) {
	opt := DownloadOption{
		Container:   ContainerMP4,
		IsAudioOnly: true,
		AudioStream: &AudioStreamInfo{
			StreamInfo: StreamInfo{Quality: "AUDIO_QUALITY_HIGH", Bitrate: 256000},
		},
	}

	label := opt.QualityLabel()
	if label != "Audio" {
		t.Errorf("expected quality label 'Audio', got %q", label)
	}
}

func TestStreamManifest_GetDownloadOptions_Basic(t *testing.T) {
	manifest := &StreamManifest{
		VideoStreams: []VideoStreamInfo{
			{
				StreamInfo: StreamInfo{URL: "https://example.com/v1080", Quality: "1080p", Bitrate: 5000000, Container: ContainerMP4},
				Width:      1920,
				Height:     1080,
			},
		},
		AudioStreams: []AudioStreamInfo{
			{
				StreamInfo: StreamInfo{URL: "https://example.com/a128", Quality: "AUDIO_QUALITY_MEDIUM", Bitrate: 128000, Container: ContainerMP4},
			},
		},
	}

	options := manifest.GetDownloadOptions()
	if len(options) == 0 {
		t.Fatal("expected at least one download option")
	}

	// Should have video+audio options and audio-only options
	hasVideoOption := false
	hasAudioOnlyOption := false
	for _, opt := range options {
		if !opt.IsAudioOnly && opt.VideoStream != nil {
			hasVideoOption = true
		}
		if opt.IsAudioOnly {
			hasAudioOnlyOption = true
		}
	}

	if !hasVideoOption {
		t.Error("expected at least one video+audio download option")
	}
	if !hasAudioOnlyOption {
		t.Error("expected at least one audio-only download option")
	}
}

func TestStreamManifest_GetDownloadOptions_MuxedStreams(t *testing.T) {
	manifest := &StreamManifest{
		MuxedStreams: []MuxedStreamInfo{
			{
				VideoStreamInfo: VideoStreamInfo{
					StreamInfo: StreamInfo{URL: "https://example.com/muxed", Quality: "360p", Bitrate: 500000, Container: ContainerMP4},
					Width:      640,
					Height:     360,
				},
				AudioStreamInfo: AudioStreamInfo{
					AudioCodec: "mp4a.40.2",
				},
			},
		},
	}

	options := manifest.GetDownloadOptions()
	if len(options) == 0 {
		t.Fatal("expected at least one download option from muxed streams")
	}

	// Check that we have an option from the muxed stream
	foundMuxed := false
	for _, opt := range options {
		if opt.VideoStream != nil && opt.VideoStream.Height == 360 {
			foundMuxed = true
			break
		}
	}
	if !foundMuxed {
		t.Error("expected to find muxed stream download option")
	}
}

func TestStreamManifest_GetDownloadOptions_MultipleQualities(t *testing.T) {
	manifest := &StreamManifest{
		VideoStreams: []VideoStreamInfo{
			{StreamInfo: StreamInfo{URL: "https://example.com/v1080", Quality: "1080p", Bitrate: 5000000, Container: ContainerMP4}, Width: 1920, Height: 1080},
			{StreamInfo: StreamInfo{URL: "https://example.com/v720", Quality: "720p", Bitrate: 2500000, Container: ContainerMP4}, Width: 1280, Height: 720},
			{StreamInfo: StreamInfo{URL: "https://example.com/v480", Quality: "480p", Bitrate: 1000000, Container: ContainerMP4}, Width: 854, Height: 480},
		},
		AudioStreams: []AudioStreamInfo{
			{StreamInfo: StreamInfo{URL: "https://example.com/a128", Quality: "AUDIO_QUALITY_MEDIUM", Bitrate: 128000, Container: ContainerMP4}},
		},
	}

	options := manifest.GetDownloadOptions()

	// Count video options (excluding audio-only)
	videoOptionCount := 0
	for _, opt := range options {
		if !opt.IsAudioOnly && opt.VideoStream != nil {
			videoOptionCount++
		}
	}

	// Should have at least one option per video quality
	if videoOptionCount < 3 {
		t.Errorf("expected at least 3 video quality options, got %d", videoOptionCount)
	}
}

func TestStreamManifest_GetDownloadOptions_Empty(t *testing.T) {
	manifest := &StreamManifest{
		VideoStreams: []VideoStreamInfo{},
		AudioStreams: []AudioStreamInfo{},
		MuxedStreams: []MuxedStreamInfo{},
	}

	options := manifest.GetDownloadOptions()
	if len(options) != 0 {
		t.Errorf("expected 0 options for empty manifest, got %d", len(options))
	}
}

func TestVideoQualityPreference_String(t *testing.T) {
	tests := []struct {
		pref     VideoQualityPreference
		expected string
	}{
		{QualityLowest, "Lowest quality"},
		{QualityUpTo360p, "≤ 360p"},
		{QualityUpTo480p, "≤ 480p"},
		{QualityUpTo720p, "≤ 720p"},
		{QualityUpTo1080p, "≤ 1080p"},
		{QualityHighest, "Highest quality"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.pref.String()
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestVideoQualityPreference_MaxHeight(t *testing.T) {
	tests := []struct {
		pref      VideoQualityPreference
		maxHeight int
	}{
		{QualityLowest, 0},
		{QualityUpTo360p, 360},
		{QualityUpTo480p, 480},
		{QualityUpTo720p, 720},
		{QualityUpTo1080p, 1080},
		{QualityHighest, 0}, // 0 means no limit
	}

	for _, tt := range tests {
		t.Run(tt.pref.String(), func(t *testing.T) {
			got := tt.pref.MaxHeight()
			if got != tt.maxHeight {
				t.Errorf("expected max height %d, got %d", tt.maxHeight, got)
			}
		})
	}
}

func TestSelectBestOption_Highest(t *testing.T) {
	options := []DownloadOption{
		{Container: ContainerMP4, VideoStream: &VideoStreamInfo{StreamInfo: StreamInfo{Quality: "360p"}, Height: 360}},
		{Container: ContainerMP4, VideoStream: &VideoStreamInfo{StreamInfo: StreamInfo{Quality: "720p"}, Height: 720}},
		{Container: ContainerMP4, VideoStream: &VideoStreamInfo{StreamInfo: StreamInfo{Quality: "1080p"}, Height: 1080}},
	}

	best := SelectBestOption(options, QualityHighest, ContainerMP4)
	if best == nil {
		t.Fatal("expected to find a best option")
	}
	if best.VideoStream.Height != 1080 {
		t.Errorf("expected highest quality (1080p), got %dp", best.VideoStream.Height)
	}
}

func TestSelectBestOption_Lowest(t *testing.T) {
	options := []DownloadOption{
		{Container: ContainerMP4, VideoStream: &VideoStreamInfo{StreamInfo: StreamInfo{Quality: "360p"}, Height: 360}},
		{Container: ContainerMP4, VideoStream: &VideoStreamInfo{StreamInfo: StreamInfo{Quality: "720p"}, Height: 720}},
		{Container: ContainerMP4, VideoStream: &VideoStreamInfo{StreamInfo: StreamInfo{Quality: "1080p"}, Height: 1080}},
	}

	best := SelectBestOption(options, QualityLowest, ContainerMP4)
	if best == nil {
		t.Fatal("expected to find a best option")
	}
	if best.VideoStream.Height != 360 {
		t.Errorf("expected lowest quality (360p), got %dp", best.VideoStream.Height)
	}
}

func TestSelectBestOption_UpTo720p(t *testing.T) {
	options := []DownloadOption{
		{Container: ContainerMP4, VideoStream: &VideoStreamInfo{StreamInfo: StreamInfo{Quality: "360p"}, Height: 360}},
		{Container: ContainerMP4, VideoStream: &VideoStreamInfo{StreamInfo: StreamInfo{Quality: "720p"}, Height: 720}},
		{Container: ContainerMP4, VideoStream: &VideoStreamInfo{StreamInfo: StreamInfo{Quality: "1080p"}, Height: 1080}},
		{Container: ContainerMP4, VideoStream: &VideoStreamInfo{StreamInfo: StreamInfo{Quality: "4K"}, Height: 2160}},
	}

	best := SelectBestOption(options, QualityUpTo720p, ContainerMP4)
	if best == nil {
		t.Fatal("expected to find a best option")
	}
	// Should select 720p (highest within limit)
	if best.VideoStream.Height != 720 {
		t.Errorf("expected 720p (highest within limit), got %dp", best.VideoStream.Height)
	}
}

func TestSelectBestOption_ContainerPreference(t *testing.T) {
	options := []DownloadOption{
		{Container: ContainerMP4, VideoStream: &VideoStreamInfo{StreamInfo: StreamInfo{Quality: "1080p"}, Height: 1080}},
		{Container: ContainerWebM, VideoStream: &VideoStreamInfo{StreamInfo: StreamInfo{Quality: "1080p"}, Height: 1080}},
	}

	// Prefer MP4
	best := SelectBestOption(options, QualityHighest, ContainerMP4)
	if best == nil {
		t.Fatal("expected to find a best option")
	}
	if best.Container != ContainerMP4 {
		t.Errorf("expected MP4 container, got %s", best.Container)
	}

	// Prefer WebM
	best = SelectBestOption(options, QualityHighest, ContainerWebM)
	if best == nil {
		t.Fatal("expected to find a best option")
	}
	if best.Container != ContainerWebM {
		t.Errorf("expected WebM container, got %s", best.Container)
	}
}

func TestSelectBestOption_NoOptions(t *testing.T) {
	var options []DownloadOption

	best := SelectBestOption(options, QualityHighest, ContainerMP4)
	if best != nil {
		t.Error("expected nil for empty options")
	}
}

func TestSelectBestOption_AudioOnlyExcluded(t *testing.T) {
	options := []DownloadOption{
		{Container: ContainerMP4, IsAudioOnly: true, AudioStream: &AudioStreamInfo{StreamInfo: StreamInfo{Bitrate: 128000}}},
		{Container: ContainerMP4, VideoStream: &VideoStreamInfo{StreamInfo: StreamInfo{Quality: "720p"}, Height: 720}},
	}

	best := SelectBestOption(options, QualityHighest, ContainerMP4)
	if best == nil {
		t.Fatal("expected to find a best option")
	}
	if best.IsAudioOnly {
		t.Error("should not select audio-only option when selecting video quality")
	}
}
