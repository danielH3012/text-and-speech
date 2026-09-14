package textandspeech

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanTextForTTS(t *testing.T) {
	raw := "Hello **world**, check this `code` and [link](https://example.com)!"
	expected := "Hello world, check this code and link!"
	got := CleanTextForTTS(raw)
	if got != expected {
		t.Errorf("CleanTextForTTS() = %q, want %q", got, expected)
	}
}

func TestDetectLanguageTTS(t *testing.T) {
	if lang := DetectLanguageTTS("Ini adalah pengujian bahasa Indonesia"); lang != "id" {
		t.Errorf("expected id, got %s", lang)
	}
	if lang := DetectLanguageTTS("This is an English test sentence"); lang != "en" {
		t.Errorf("expected en, got %s", lang)
	}
}

func TestSaveBytesAsWAV(t *testing.T) {
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "test.wav")

	dummyWav := []byte("RIFF\x24\x00\x00\x00WAVEfmt \x10\x00\x00\x00")
	if err := saveBytesAsWAV(dummyWav, outPath); err != nil {
		t.Fatalf("saveBytesAsWAV failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read created wav file: %v", err)
	}
	if len(data) != len(dummyWav) {
		t.Errorf("file size = %d, want %d", len(data), len(dummyWav))
	}
}

func TestTTSConfig(t *testing.T) {
	// 1. Using exact uppercase field names
	cfgExact := TTSConfig{
		API_KEY:           "test-key",
		TTS_URL:           "https://api.inworld.ai/tts/v1/voice",
		TTS_MODEL:         "inworld-tts-2",
		TTS_VOICE:         "Sarah",
		TTS_AUTH_PREFIX:   "Basic ",
		TTS_DECODE:        "base64",
		TTS_BODY_TEMPLATE: `{"text":"{{.Text}}","voiceId":"{{.Voice}}","modelId":"{{.Model}}"}`,
	}

	synthExact := NewSynthesizer(cfgExact)
	if synthExact.Name() != "inworld-tts-2" {
		t.Errorf("expected name inworld-tts-2, got %s", synthExact.Name())
	}

	// 2. Using camelCase aliases
	cfgAlias := TTSConfig{
		APIKey:       "test-key",
		URL:          "https://api.example.com/tts",
		Model:        "tts-1",
		Voice:        "alloy",
		AuthPrefix:   "Bearer",
		Decode:       "base64",
		BodyTemplate: `{"model":"{{.Model}}","text":"{{.Text}}"}`,
	}

	synth := NewSynthesizer(cfgAlias)
	if synth.Name() != "tts-1" {
		t.Errorf("expected name tts-1, got %s", synth.Name())
	}

	// Verify error when config fields are missing
	emptySynth := NewSynthesizer(TTSConfig{})
	_, err := emptySynth.Synthesize("halo", "", "")
	if err == nil {
		t.Error("expected error when URL is missing, got nil")
	}
}

func TestSTTConfig(t *testing.T) {
	cfg := STTConfig{
		APIKey:   "test-key",
		URL:      "https://api.example.com/stt",
		Model:    "whisper-large-v3",
		Language: "id",
	}

	transcriber := NewTranscriber(cfg)
	if transcriber.Name() != "whisper-large-v3" {
		t.Errorf("expected name whisper-large-v3, got %s", transcriber.Name())
	}

	// Verify error when config fields are missing
	emptyTranscriber := NewTranscriber(STTConfig{})
	_, err := emptyTranscriber.Transcribe("dummy.wav")
	if err == nil {
		t.Error("expected error when URL is missing, got nil")
	}
}

