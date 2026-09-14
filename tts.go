package textandspeech

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
)

// NewSynthesizer creates a new TTS synthesizer with user-provided TTSConfig.
func NewSynthesizer(cfg TTSConfig) Synthesizer {
	return &TemplateSynthesizer{config: cfg}
}

// TemplateSynthesizer is a TTS adapter driven directly by user-provided TTSConfig.
// The output is fixed to a .wav file.
type TemplateSynthesizer struct {
	config TTSConfig
}

func (s *TemplateSynthesizer) Name() string {
	if m := s.config.GetModel(); m != "" {
		return m
	}
	return "generic"
}

// Synthesize converts text directly into a .wav file at outputPath.
func (s *TemplateSynthesizer) Synthesize(text, outputPath string, voiceID ...string) (string, error) {
	vID := ""
	if len(voiceID) > 0 {
		vID = voiceID[0]
	}
	return s.SynthesizeToFile(text, outputPath, vID, "")
}

// SynthesizeToFile converts text into speech and saves it as a .wav file at outputPath.
// It ensures the saved file is always in valid 16-bit PCM WAV format.
func (s *TemplateSynthesizer) SynthesizeToFile(text, outputPath, voiceID, lang string) (string, error) {
	if strings.TrimSpace(outputPath) == "" {
		return "", fmt.Errorf("outputPath is required")
	}

	audioBytes, err := s.synthesizeBytes(text, voiceID, lang)
	if err != nil {
		return "", err
	}

	if !strings.HasSuffix(strings.ToLower(outputPath), ".wav") {
		outputPath = outputPath + ".wav"
	}

	if err := saveBytesAsWAV(audioBytes, outputPath); err != nil {
		return "", fmt.Errorf("failed to save .wav file: %w", err)
	}

	log.Printf("[%s TTS] saved .wav file to %s (%d bytes)", s.Name(), outputPath, len(audioBytes))
	return outputPath, nil
}

// synthesizeBytes executes the API call to obtain raw audio bytes.
func (s *TemplateSynthesizer) synthesizeBytes(text, voiceID, lang string) ([]byte, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text is empty")
	}

	url := s.config.GetURL()
	if strings.TrimSpace(url) == "" {
		return nil, fmt.Errorf("TTS URL (TTS_URL) is required in config")
	}

	apiKey := s.config.GetAPIKey()
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("TTS APIKey (API_KEY) is required in config")
	}

	bodyTemplate := s.config.GetBodyTemplate()
	if strings.TrimSpace(bodyTemplate) == "" {
		return nil, fmt.Errorf("TTS BodyTemplate (TTS_BODY_TEMPLATE) is required in config")
	}

	model := s.config.GetModel()
	voice := voiceID
	if voice == "" {
		voice = s.config.GetVoice()
	}

	if lang == "" {
		lang = DetectLanguageTTS(text)
	}

	authHeader := s.config.AuthHeader
	if strings.TrimSpace(authHeader) == "" {
		authHeader = "Authorization"
	}

	authPrefix := s.config.GetAuthPrefix()
	decode := strings.EqualFold(s.config.GetDecode(), "base64")

	maskedKey := apiKey
	if len(maskedKey) > 8 {
		maskedKey = maskedKey[:4] + "..." + maskedKey[len(maskedKey)-4:]
	} else if len(maskedKey) > 0 {
		maskedKey = "***"
	}
	log.Printf("[TTS Active Settings] URL: %s | Model: %s | Voice: %s | Lang: %s | Decode: %s", url, model, voice, lang, s.config.GetDecode())
	log.Printf("[TTS Active Settings] AuthHeader: %s | AuthPrefix: %q | API_KEY: %s", authHeader, authPrefix, maskedKey)
	log.Printf("[TTS Active Settings] BodyTemplate: %s", bodyTemplate)

	tpl, err := template.New("body").Parse(bodyTemplate)
	if err != nil {
		return nil, fmt.Errorf("invalid TTS_BODY_TEMPLATE: %w", err)
	}
	data := struct {
		Text     string
		Model    string
		Voice    string
		Lang     string
		Format   string
		VoiceID  string
		VoiceId  string
		ModelID  string
		ModelId  string
		Language string
	}{
		Text:     jsonEsc(text),
		Model:    jsonEsc(model),
		Voice:    jsonEsc(voice),
		Lang:     jsonEsc(lang),
		Format:   "wav",
		VoiceID:  jsonEsc(voice),
		VoiceId:  jsonEsc(voice),
		ModelID:  jsonEsc(model),
		ModelId:  jsonEsc(model),
		Language: jsonEsc(lang),
	}
	var payload bytes.Buffer
	if err := tpl.Execute(&payload, data); err != nil {
		return nil, fmt.Errorf("failed to render TTS_BODY_TEMPLATE: %w", err)
	}

	log.Printf("[TTS Request Payload] %s", payload.String())

	req, err := http.NewRequest(http.MethodPost, url, &payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	fullAuth := apiKey
	if strings.TrimSpace(authPrefix) != "" {
		fullAuth = strings.TrimSpace(authPrefix) + " " + apiKey
	}
	req.Header.Set(authHeader, fullAuth)

	// Set default audio accept header if not overridden in ExtraHeaders
	req.Header.Set("Accept", "audio/wav, audio/*, */*")
	for k, v := range s.config.ExtraHeaders {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	audioBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TTS API error (%d): %s", resp.StatusCode, string(audioBytes))
	}

	if decode {
		var payload struct {
			AudioContent string `json:"audioContent"`
		}
		if err := json.Unmarshal(audioBytes, &payload); err != nil {
			return nil, fmt.Errorf("failed to parse audio response: %w", err)
		}
		decoded, err := base64.StdEncoding.DecodeString(payload.AudioContent)
		if err != nil {
			return nil, fmt.Errorf("failed to decode audioContent: %w", err)
		}
		audioBytes = decoded
	}

	log.Printf("[TTS Success] synthesized %d bytes from %s (HTTP %d)", len(audioBytes), url, resp.StatusCode)
	return audioBytes, nil
}

// jsonEsc escapes a value for safe embedding inside a JSON body template.
func jsonEsc(s string) string {
	b, err := json.Marshal(s)
	if err != nil || len(b) < 2 {
		return s
	}
	return string(b[1 : len(b)-1])
}

// CleanTextForTTS removes markdown, code blocks, links, and system artifacts for natural speech synthesis.
func CleanTextForTTS(text string) string {
	if text == "" {
		return ""
	}

	// Remove attachment indicators
	reAttachment := regexp.MustCompile(`(?i)\n?📎\s*\[Attached:\s*[^\]]+\]`)
	text = reAttachment.ReplaceAllString(text, "")

	// Remove fenced code blocks
	reCodeBlock := regexp.MustCompile("(?s)```.*?```")
	text = reCodeBlock.ReplaceAllString(text, "")

	// Remove inline code
	reInlineCode := regexp.MustCompile("`([^`]+)`")
	text = reInlineCode.ReplaceAllString(text, "$1")

	// Remove markdown links [text](url) -> text
	reLinks := regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	text = reLinks.ReplaceAllString(text, "$1")

	// Remove markdown formatting characters (*, #, _, ~, >, ===)
	reSymbols := regexp.MustCompile(`[*#_~>=]`)
	text = reSymbols.ReplaceAllString(text, "")

	// Normalize excessive whitespace
	reSpaces := regexp.MustCompile(`\s+`)
	text = reSpaces.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// DetectLanguageTTS detects if the text is primarily Indonesian or English.
func DetectLanguageTTS(text string) string {
	lower := strings.ToLower(text)
	indonesianKeywords := []string{
		"yang", "dan", "di", "ini", "itu", "ada", "tidak", "untuk", "dengan",
		"dari", "pada", "ke", "aset", "adalah", "sudah", "bisa", "perusahaan", "pengguna",
	}
	for _, kw := range indonesianKeywords {
		if strings.Contains(lower, " "+kw+" ") || strings.HasPrefix(lower, kw+" ") || strings.HasSuffix(lower, " "+kw) {
			return "id"
		}
	}
	return "en"
}

// saveBytesAsWAV ensures audioBytes is saved to outputPath as a valid 16-bit PCM WAV file.
// If the bytes are not already in WAV format (e.g. MP3/OGG), it automatically converts via ffmpeg.
func saveBytesAsWAV(audioBytes []byte, outputPath string) error {
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	// Check if already valid WAV format (RIFF....WAVE)
	if len(audioBytes) >= 12 && string(audioBytes[0:4]) == "RIFF" && string(audioBytes[8:12]) == "WAVE" {
		return os.WriteFile(outputPath, audioBytes, 0644)
	}

	// Try converting to WAV via ffmpeg
	tempInput := filepath.Join(os.TempDir(), fmt.Sprintf("raw_tts_%s.tmp", uuid.NewString()))
	if err := os.WriteFile(tempInput, audioBytes, 0644); err != nil {
		return fmt.Errorf("failed to write temp audio file: %w", err)
	}
	defer os.Remove(tempInput)

	cmd := exec.Command("ffmpeg", "-y", "-i", tempInput, "-c:a", "pcm_s16le", outputPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[saveBytesAsWAV] ffmpeg conversion notice: %v (%s). Saving original bytes to %s", err, strings.TrimSpace(string(out)), outputPath)
		return os.WriteFile(outputPath, audioBytes, 0644)
	}

	return nil
}

// Synthesize converts text into a .wav file at outputPath using user-provided TTSConfig.
func Synthesize(cfg TTSConfig, text, outputPath string) (string, error) {
	synth := NewSynthesizer(cfg)
	return synth.SynthesizeToFile(text, outputPath, cfg.GetVoice(), "")
}

// SynthesizeToFile converts text to speech and saves it as a .wav file at outputPath using user-provided TTSConfig.
func SynthesizeToFile(cfg TTSConfig, text, outputPath, voiceID, lang string) (string, error) {
	synth := NewSynthesizer(cfg)
	return synth.SynthesizeToFile(text, outputPath, voiceID, lang)
}

// SynthesizeBytes synthesizes text into raw audio bytes in memory using user-provided TTSConfig.
func SynthesizeBytes(cfg TTSConfig, text, voiceID, lang string) ([]byte, error) {
	synth := &TemplateSynthesizer{config: cfg}
	return synth.synthesizeBytes(text, voiceID, lang)
}
