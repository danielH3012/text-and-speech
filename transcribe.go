package textandspeech

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NormalizeSTTLanguage standardizes language codes:
// "id" for Indonesian, "en" for English, or "" for auto-detect (bilingual / dwibahasa).
func NormalizeSTTLanguage(lang string) string {
	l := strings.ToLower(strings.TrimSpace(lang))
	switch l {
	case "id", "indonesia", "indonesian", "bahasa":
		return "id"
	case "en", "english", "inggris":
		return "en"
	case "auto", "detect", "both", "all", "":
		return ""
	default:
		return l
	}
}

// NewTranscriber creates a new STT transcriber with user-provided STTConfig.
func NewTranscriber(cfg STTConfig) Transcriber {
	return &GenericTranscriber{config: cfg}
}

// preprocessAudio cleans background noise, removes DC rumble,
// normalizes vocal loudness, and trims dead silence using ffmpeg.
func preprocessAudio(inputPath string) (string, func()) {
	_ = os.MkdirAll("./tmp", 0755)
	cleanedPath := filepath.Join("./tmp", fmt.Sprintf("clean_%s.wav", uuid.NewString()))
	cleanup := func() { _ = os.Remove(cleanedPath) }

	filterChain := "highpass=f=80,lowpass=f=8000,afftdn=nf=-25,loudnorm=I=-16:TP=-1.5:LRA=11,silenceremove=start_periods=1:start_duration=0.05:start_threshold=-50dB:stop_periods=-1:stop_duration=0.8:stop_threshold=-50dB"

	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath, "-af", filterChain, "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", cleanedPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[preprocessAudio] ffmpeg filter notice: %v (%s). Using original audio.", err, strings.TrimSpace(string(out)))
		return inputPath, func() {}
	}

	if fi, err := os.Stat(cleanedPath); err == nil && fi.Size() > 1024 {
		log.Printf("[preprocessAudio] Cleaned audio generated (%d bytes)", fi.Size())
		return cleanedPath, cleanup
	}

	return inputPath, func() {}
}

// parseTranscriptResponse extracts text from a transcription response,
// supporting OpenAI-compatible JSON shapes and plain-text responses.
func parseTranscriptResponse(raw []byte, contentType string) string {
	if strings.Contains(strings.ToLower(contentType), "json") {
		var payload struct {
			Text       string `json:"text"`
			OutputText string `json:"output_text"`
			Transcript string `json:"transcript"`
			Result     struct {
				Text string `json:"text"`
			} `json:"result"`
			Data struct {
				Text string `json:"text"`
			} `json:"data"`
			Results struct {
				Transcripts []struct {
					Transcript string `json:"transcript"`
				} `json:"transcripts"`
				Channels []struct {
					Alternatives []struct {
						Transcript string `json:"transcript"`
					} `json:"alternatives"`
				} `json:"channels"`
			} `json:"results"`
		}
		if err := json.Unmarshal(raw, &payload); err == nil {
			candidates := []string{
				payload.Text,
				payload.OutputText,
				payload.Transcript,
				payload.Result.Text,
				payload.Data.Text,
			}
			if len(payload.Results.Transcripts) > 0 {
				candidates = append(candidates, payload.Results.Transcripts[0].Transcript)
			}
			if len(payload.Results.Channels) > 0 && len(payload.Results.Channels[0].Alternatives) > 0 {
				candidates = append(candidates, payload.Results.Channels[0].Alternatives[0].Transcript)
			}
			for _, candidate := range candidates {
				if strings.TrimSpace(candidate) != "" {
					return strings.TrimSpace(candidate)
				}
			}
			// When JSON was parsed successfully but candidate was empty (e.g. silent audio),
			// return empty string instead of raw JSON
			return ""
		}
	}
	return strings.TrimSpace(string(raw))
}

// GenericTranscriber is an STT adapter driven directly by user-provided STTConfig.
type GenericTranscriber struct {
	config STTConfig
}

func (t *GenericTranscriber) Name() string {
	if m := t.config.GetModel(); m != "" {
		return m
	}
	return "generic"
}

func (t *GenericTranscriber) Transcribe(audioPath string) (string, error) {
	url := t.config.GetURL()
	if strings.TrimSpace(url) == "" {
		return "", fmt.Errorf("STT URL (STT_URL) is required in config")
	}

	apiKey := t.config.GetAPIKey()
	if strings.TrimSpace(apiKey) == "" {
		return "", fmt.Errorf("STT APIKey (API_KEY) is required in config")
	}

	model := t.config.GetModel()
	language := NormalizeSTTLanguage(t.config.GetLanguage())
	fileField := t.config.GetFileField()
	modelField := t.config.GetModelField()
	langField := t.config.GetLanguageField()
	authHeader := t.config.GetAuthHeader()
	authPrefix := t.config.GetAuthPrefix()

	// Print active settings
	maskedKey := apiKey
	if len(maskedKey) > 8 {
		maskedKey = maskedKey[:4] + "..." + maskedKey[len(maskedKey)-4:]
	} else if len(maskedKey) > 0 {
		maskedKey = "***"
	}
	log.Printf("[STT Active Settings] URL: %s", url)
	log.Printf("[STT Active Settings] Model: %s | Language: %s | FileField: %s | ModelField: %s", model, language, fileField, modelField)
	log.Printf("[STT Active Settings] AuthHeader: %s | AuthPrefix: %q | API_KEY: %s", authHeader, authPrefix, maskedKey)

	cleanPath, cleanup := preprocessAudio(audioPath)
	defer cleanup()
	audioPath = cleanPath

	file, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	filename := filepath.Base(audioPath)
	if filepath.Ext(filename) == "" {
		filename = "audio.wav"
	}

	part, err := writer.CreateFormFile(fileField, filename)
	if err != nil {
		return "", fmt.Errorf("failed to create multipart form: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}

	_ = writer.WriteField(modelField, model)
	// Write language field if explicitly specified (e.g. "id" or "en").
	// If empty/auto, omit the field to allow automatic language detection.
	if language != "" && langField != "" {
		_ = writer.WriteField(langField, language)
	}
	for k, v := range t.config.ExtraFields {
		_ = writer.WriteField(k, v)
	}
	_ = writer.Close()

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	fullAuth := apiKey
	if strings.TrimSpace(authPrefix) != "" {
		fullAuth = strings.TrimSpace(authPrefix) + " " + apiKey
	}
	req.Header.Set(authHeader, fullAuth)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for k, v := range t.config.ExtraHeaders {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request to %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("transcription API error (%d): %s", resp.StatusCode, string(respBytes))
	}

	result := parseTranscriptResponse(respBytes, resp.Header.Get("Content-Type"))
	log.Printf("[%s STT] transcribed text: %q", t.Name(), result)
	return result, nil
}

// Transcribe converts an audio file into a text string using the user-provided STTConfig.
func Transcribe(cfg STTConfig, audioPath string) (string, error) {
	transcriber := NewTranscriber(cfg)
	return transcriber.Transcribe(audioPath)
}
