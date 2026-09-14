package textandspeech

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
	return c.AuthPrefix
}

func (c TTSConfig) GetDecode() string {
	if c.TTS_DECODE != "" {
		return c.TTS_DECODE
	}
	return c.Decode
}

func (c TTSConfig) GetBodyTemplate() string {
	if c.TTS_BODY_TEMPLATE != "" {
		return c.TTS_BODY_TEMPLATE
	}
	return c.BodyTemplate
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
	return "model"
}

func (c STTConfig) GetLanguageField() string {
	if c.STT_LANGUAGE_FIELD != "" {
		return c.STT_LANGUAGE_FIELD
	}
	if c.LanguageField != "" {
		return c.LanguageField
	}
	return "language"
}

func (c STTConfig) GetAuthHeader() string {
	if c.STT_HEADER != "" {
		return c.STT_HEADER
	}
	if c.AuthHeader != "" {
		return c.AuthHeader
	}
	return "Authorization"
}

func (c STTConfig) GetAuthPrefix() string {
	if c.STT_AUTH_PREFIX != "" {
		return c.STT_AUTH_PREFIX
	}
	return c.AuthPrefix
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
