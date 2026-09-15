package textandspeech

import "strings"

// TTSConfig defines the configuration options for Text-To-Speech (TTS).
// All settings are provided directly via user input.
// Supports both exact field names (API_KEY, TTS_URL, etc.) and camelCase aliases (APIKey, URL, etc.).
type TTSConfig struct {
	// Exact input fields requested:
	API_KEY           string `json:"api_key"`
	TTS_URL           string `json:"tts_url"`
	TTS_MODEL         string `json:"tts_model"`
	TTS_VOICE         string `json:"tts_voice"`
	TTS_AUTH_PREFIX   string `json:"tts_auth_prefix"`
	TTS_DECODE        string `json:"tts_decode"`
	TTS_BODY_TEMPLATE string `json:"tts_body_template"`

	// Standard Go aliases for flexible usage:
	APIKey       string `json:"-"`
	URL          string `json:"-"`
	Model        string `json:"-"`
	Voice        string `json:"-"`
	AuthPrefix   string `json:"-"`
	Decode       string `json:"-"`
	BodyTemplate string `json:"-"`

	// Optional additional parameters
	AuthHeader   string            `json:"tts_auth_header,omitempty"`
	ExtraHeaders map[string]string `json:"extra_headers,omitempty"`
}

func (c TTSConfig) GetAPIKey() string {
	if c.API_KEY != "" {
		return c.API_KEY
	}
	return c.APIKey
}

func (c TTSConfig) GetURL() string {
	if c.TTS_URL != "" {
		return c.TTS_URL
	}
	return c.URL
}

func (c TTSConfig) GetModel() string {
	if c.TTS_MODEL != "" {
		return c.TTS_MODEL
	}
	return c.Model
}

func (c TTSConfig) GetVoice() string {
	if c.TTS_VOICE != "" {
		return c.TTS_VOICE
	}
	return c.Voice
}

func (c TTSConfig) GetAuthPrefix() string {
	if c.TTS_AUTH_PREFIX != "" {
		return c.TTS_AUTH_PREFIX
	}
	if c.AuthPrefix != "" {
		return c.AuthPrefix
	}
	url := strings.ToLower(c.GetURL())
	if strings.Contains(url, "inworld") {
		return "Basic "
	}
	if strings.Contains(url, "elevenlabs") || strings.Contains(url, "cartesia") {
		return ""
	}
	return "Bearer "
}

func (c TTSConfig) GetAuthHeader() string {
	if c.AuthHeader != "" {
		return c.AuthHeader
	}
	url := strings.ToLower(c.GetURL())
	if strings.Contains(url, "elevenlabs") {
		return "xi-api-key"
	}
	if strings.Contains(url, "cartesia") {
		return "X-API-Key"
	}
	return "Authorization"
}

func (c TTSConfig) GetDecode() string {
	if c.TTS_DECODE != "" {
		return c.TTS_DECODE
	}
	if c.Decode != "" {
		return c.Decode
	}
	url := strings.ToLower(c.GetURL())
	if strings.Contains(url, "inworld") {
		return "base64"
	}
	return ""
}

func (c TTSConfig) GetBodyTemplate() string {
	if c.TTS_BODY_TEMPLATE != "" {
		return c.TTS_BODY_TEMPLATE
	}
	if c.BodyTemplate != "" {
		return c.BodyTemplate
	}

	url := strings.ToLower(c.GetURL())

	// 1. Inworld AI TTS
	if strings.Contains(url, "inworld") {
		return `{"text":"{{.Text}}","voiceId":"{{.Voice}}","modelId":"{{.Model}}"}`
	}

	// 2. OpenAI / Groq TTS
	if strings.Contains(url, "openai") {
		return `{"model":"{{.Model}}","input":"{{.Text}}","voice":"{{.Voice}}"}`
	}

	// 3. Cartesia TTS (Sonic API)
	if strings.Contains(url, "cartesia") {
		return `{"model_id":"{{.Model}}","transcript":"{{.Text}}","voice":{"mode":"id","id":"{{.Voice}}"},"language":"{{.Lang}}"}`
	}

	// 4. ElevenLabs TTS
	if strings.Contains(url, "elevenlabs") {
		return `{"text":"{{.Text}}","model_id":"{{.Model}}"}`
	}

	// 5. Default generic template
	return `{"text":"{{.Text}}","voice":"{{.Voice}}","model":"{{.Model}}"}`
}

// STTConfig defines configuration options for Speech-To-Text (Transcribe).
// All settings are provided directly via user input.
// Supports both uppercase input fields and standard camelCase aliases.
type STTConfig struct {
	API_KEY            string `json:"api_key"`
	STT_URL            string `json:"stt_url"`
	STT_MODEL          string `json:"stt_model"`
	STT_LANGUAGE       string `json:"stt_language,omitempty"`
	STT_FILE_FIELD     string `json:"stt_file_field,omitempty"`
	STT_MODEL_FIELD    string `json:"stt_model_field,omitempty"`
	STT_LANGUAGE_FIELD string `json:"stt_language_field,omitempty"`
	STT_HEADER         string `json:"stt_auth_header,omitempty"`
	STT_AUTH_PREFIX    string `json:"stt_auth_prefix,omitempty"`

	APIKey        string `json:"-"`
	URL           string `json:"-"`
	Model         string `json:"-"`
	Language      string `json:"-"`
	FileField     string `json:"-"`
	ModelField    string `json:"-"`
	LanguageField string `json:"-"`
	AuthHeader    string `json:"-"`
	AuthPrefix    string `json:"-"`

	ExtraFields  map[string]string `json:"extra_fields,omitempty"`
	ExtraHeaders map[string]string `json:"extra_headers,omitempty"`
}

func (c STTConfig) GetAPIKey() string {
	if c.API_KEY != "" {
		return c.API_KEY
	}
	return c.APIKey
}

func (c STTConfig) GetURL() string {
	if c.STT_URL != "" {
		return c.STT_URL
	}
	return c.URL
}

func (c STTConfig) GetModel() string {
	if c.STT_MODEL != "" {
		return c.STT_MODEL
	}
	return c.Model
}

func (c STTConfig) GetLanguage() string {
	if c.STT_LANGUAGE != "" {
		return c.STT_LANGUAGE
	}
	return c.Language
}

func (c STTConfig) GetFileField() string {
	if c.STT_FILE_FIELD != "" {
		return c.STT_FILE_FIELD
	}
	if c.FileField != "" {
		return c.FileField
	}
	return "file"
}

func (c STTConfig) GetModelField() string {
	if c.STT_MODEL_FIELD != "" {
		return c.STT_MODEL_FIELD
	}
	if c.ModelField != "" {
		return c.ModelField
	}
	url := strings.ToLower(c.GetURL())
	if strings.Contains(url, "elevenlabs") || strings.Contains(url, "assemblyai") {
		return "model_id"
	}
	return "model"
}

func (c STTConfig) GetLanguageField() string {
	if c.STT_LANGUAGE_FIELD != "" {
		return c.STT_LANGUAGE_FIELD
	}
	if c.LanguageField != "" {
		return c.LanguageField
	}

	url := strings.ToLower(c.GetURL())

	// Providers using "language_code" (ElevenLabs, AssemblyAI, Sarvam)
	if strings.Contains(url, "elevenlabs") || strings.Contains(url, "assemblyai") || strings.Contains(url, "sarvam") {
		return "language_code"
	}

	// Google Cloud Speech API ("languageCode")
	if strings.Contains(url, "googleapis.com") || strings.Contains(url, "google") {
		return "languageCode"
	}

	// Azure Cognitive Services ("locale")
	if strings.Contains(url, "azure.com") || strings.Contains(url, "cognitiveservices") {
		return "locale"
	}

	// Yandex SpeechKit ("lang")
	if strings.Contains(url, "yandex") {
		return "lang"
	}

	// Default standard: OpenAI Whisper, Groq, Deepgram, Cloudflare, Rev AI, Faster-Whisper
	return "language"
}

func (c STTConfig) GetAuthHeader() string {
	if c.STT_HEADER != "" {
		return c.STT_HEADER
	}
	if c.AuthHeader != "" {
		return c.AuthHeader
	}
	url := strings.ToLower(c.GetURL())
	if strings.Contains(url, "elevenlabs") {
		return "xi-api-key"
	}
	return "Authorization"
}

func (c STTConfig) GetAuthPrefix() string {
	if c.STT_AUTH_PREFIX != "" {
		return c.STT_AUTH_PREFIX
	}
	if c.AuthPrefix != "" {
		return c.AuthPrefix
	}
	url := strings.ToLower(c.GetURL())
	if strings.Contains(url, "elevenlabs") || strings.Contains(url, "assemblyai") {
		return ""
	}
	return "Bearer "
}

// SynthesizeInput represents parameters for synthesizing speech from text.
type SynthesizeInput struct {
	Text       string `json:"text"`
	OutputPath string `json:"output_path,omitempty"`
	VoiceID    string `json:"voice_id,omitempty"`
	Language   string `json:"language,omitempty"`
}

// TranscribeInput represents parameters for transcribing speech from audio.
type TranscribeInput struct {
	AudioPath string `json:"audio_path"`
	Language  string `json:"language,omitempty"`
}

// Adapter is the minimal contract every provider adapter exposes.
type Adapter interface {
	// Name returns the provider identifier (e.g. "inworld", "openai", "elevenlabs").
	Name() string
}

// Synthesizer is the pluggable contract for any text-to-speech provider.
type Synthesizer interface {
	Adapter
	// Synthesize converts text directly into a .wav file at outputPath.
	Synthesize(text, outputPath string, voiceID ...string) (string, error)
	// SynthesizeToFile converts text into speech and saves it as a .wav file at outputPath.
	SynthesizeToFile(text, outputPath, voiceID, lang string) (string, error)
}

// Transcriber is the pluggable contract for any speech-to-text provider.
type Transcriber interface {
	Adapter
	// Transcribe converts an audio file path into clean text.
	Transcribe(audioPath string) (string, error)
}

// STTTranscriber is kept as an alias for backwards compatibility.
type STTTranscriber = Transcriber

// TTSSynthesizer is kept as an alias for backwards compatibility.
type TTSSynthesizer = Synthesizer
