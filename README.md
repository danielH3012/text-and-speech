# Text and Speech (Go Package)

Package Go untuk integrasi **Text-to-Speech (TTS)** dan **Speech-to-Text (STT / Transcribe)** universal berbasis HTTP API template. Seluruh konfigurasi disediakan langsung melalui input pengguna (**User Input**) tanpa ketergantungan pada file `.env`.

- **TTS Output**: Menghasilkan file audio berformat **`.wav`** (16-bit PCM WAV).
- **STT Output**: Menghasilkan teks transkripsi berformat **`string`**.

---

## Daftar Isi
1. [Instalasi & Prasyarat](#prasyarat)
2. [Tipe Data & Konfigurasi (Types)](#tipe-data--konfigurasi-types)
   - [TTSConfig](#1-ttsconfig)
   - [STTConfig](#2-sttconfig)
   - [Interfaces](#3-interfaces)
3. [Dokumentasi Fungsi Text-to-Speech (TTS)](#dokumentasi-fungsi-text-to-speech-tts)
   - [Fungsi Top-Level](#fungsi-top-level-tts)
   - [Instance Method (Synthesizer)](#instance-method-synthesizer)
   - [Helper Text Utility](#helper-text-utility)
4. [Dokumentasi Fungsi Speech-to-Text (Transcribe)](#dokumentasi-fungsi-speech-to-text-transcribe)
   - [Fungsi Top-Level](#fungsi-top-level-transcribe)
   - [Instance Method (Transcriber)](#instance-method-transcriber)
5. [Contoh Penggunaan Lengkap](#contoh-penggunaan-lengkap)

---

## Prasyarat
- **Go**: Versi 1.20 ke atas.
- **FFmpeg** (Opsional tapi direkomendasikan): Digunakan untuk konversi otomatis format audio ke `.wav` (PCM 16-bit) dan filter noise audio sebelum transkripsi.

---

## Tipe Data & Konfigurasi (Types)

### 1. `TTSConfig`
Digunakan untuk menentukan parameter endpoint penyedia layanan TTS (Inworld, OpenAI, ElevenLabs, Cartesia, dll.).

| Field | Tipe | Keterangan | Contoh |
| :--- | :--- | :--- | :--- |
| `API_KEY` | `string` | Kunci otentikasi API penyedia TTS. | `"sk-..."` atau `"ZHFfcUZ..."` |
| `TTS_URL` | `string` | Endpoint URL POST untuk sintesis suara. | `"https://api.inworld.ai/tts/v1/voice"` |
| `TTS_MODEL` | `string` | Nama / ID model TTS. | `"inworld-tts-2"` atau `"tts-1"` |
| `TTS_VOICE` | `string` | Voice ID atau nama suara default. | `"Sarah"` atau `"alloy"` |
| `TTS_AUTH_PREFIX` | `string` | Awalan token otentikasi pada header. | `"Basic "` atau `"Bearer "` |
| `TTS_DECODE` | `string` | Isi `"base64"` jika API mengembalikan payload JSON terenkode Base64. | `"base64"` |
| `TTS_BODY_TEMPLATE` | `string` | Template JSON untuk request body (Go template). | `{"text":"{{.Text}}","voiceId":"{{.Voice}}"}` |
| `AuthHeader` | `string` | *(Opsional)* Nama header auth (default: `"Authorization"`). | `"xi-api-key"` |
| `Format` | `string` | *(Opsional)* Format audio yang diminta ke API jika didukung. | `"wav"` |
| `Accept` | `string` | *(Opsional)* Nilai header `Accept`. | `"audio/wav"` |
| `ExtraHeaders` | `map[string]string` | *(Opsional)* Header HTTP tambahan. | `{"X-Custom": "val"}` |

> **Catatan:** `TTSConfig` juga mendukung penamaan field camelCase (`APIKey`, `URL`, `Model`, `Voice`, `AuthPrefix`, `Decode`, `BodyTemplate`) sebagai alias.

---

### 2. `STTConfig`
Digunakan untuk menentukan parameter endpoint penyedia layanan Speech-to-Text / Whisper / Scribe.

| Field | Tipe | Keterangan | Default |
| :--- | :--- | :--- | :--- |
| `API_KEY` | `string` | Kunci otentikasi API STT. | Wajib diisi |
| `STT_URL` | `string` | Endpoint URL multipart POST untuk transkripsi audio. | Wajib diisi |
| `STT_MODEL` | `string` | Nama model transkripsi. | `"whisper-large-v3"`, `"scribe_v2"` |
| `STT_LANGUAGE` | `string` | *(Opsional)* Bahasa audio yang diharapkan. | `"id"`, `"en"` |
| `STT_FILE_FIELD` | `string` | *(Opsional)* Nama multipart field untuk file audio. | `"file"` |
| `STT_MODEL_FIELD` | `string` | *(Opsional)* Nama multipart field untuk model. | `"model"` |
| `STT_LANGUAGE_FIELD` | `string` | *(Opsional)* Nama multipart field untuk bahasa. | `"language"` |
| `STT_HEADER` | `string` | *(Opsional)* Nama header otentikasi. | `"Authorization"` |
| `STT_AUTH_PREFIX` | `string` | *(Opsional)* Awalan token otentikasi. | `"Bearer "` |
| `ExtraFields` | `map[string]string` | *(Opsional)* Field form multipart tambahan. | `nil` |
| `ExtraHeaders` | `map[string]string` | *(Opsional)* Header HTTP tambahan. | `nil` |

---

### 3. Interfaces

```go
type Synthesizer interface {
    Name() string
    Synthesize(text, voiceID, lang string) ([]byte, error)
    SynthesizeToFile(text, outputPath, voiceID, lang string) (string, error)
}

type Transcriber interface {
    Name() string
    Transcribe(audioPath string) (string, error)
}
```

---

## Dokumentasi Fungsi Text-to-Speech (TTS)

### Fungsi Top-Level (TTS)

#### 1. `Synthesize`
Mengonversi teks ke file audio `.wav`.
```go
func Synthesize(cfg TTSConfig, text string, outputPath ...string) (string, error)
```
- **Input**:
  - `cfg`: `TTSConfig`
  - `text`: `string` (Teks yang ingin diucapkan)
  - `outputPath` *(opsional)*: `string` (Path file tujuan. Default: `"output.wav"`)
- **Output**:
  - `string`: Path file `.wav` yang dihasilkan.
  - `error`: Error jika sintesis gagal atau validasi konfigurasi tidak lengkap.

---

#### 2. `SynthesizeToFile`
Menyimpan hasil sintesis ke file `.wav` dengan kontrol voice ID dan bahasa spesifik.
```go
func SynthesizeToFile(cfg TTSConfig, text, outputPath, voiceID, lang string) (string, error)
```
- **Input**:
  - `cfg`: `TTSConfig`
  - `text`: `string` (Teks sumber)
  - `outputPath`: `string` (Path file tujuan, jika kosong default `"output.wav"`)
  - `voiceID`: `string` (Voice ID override, atau kosongkan untuk memakai default dari config)
  - `lang`: `string` (Bahasa override, atau kosongkan untuk deteksi otomatis)
- **Output**:
  - `string`: Path file `.wav` yang dihasilkan.
  - `error`: Error jika operasi gagal.

---

#### 3. `SynthesizeToWav`
Alias praktis untuk membuat file `.wav` dengan voice dan bahasa default.
```go
func SynthesizeToWav(cfg TTSConfig, text, outputPath string) (string, error)
```
- **Input**:
  - `cfg`: `TTSConfig`
  - `text`: `string`
  - `outputPath`: `string`
- **Output**:
  - `string`: Path file `.wav`.
  - `error`: Error jika gagal.

---

#### 4. `TextToSpeech`
Alias untuk `Synthesize`.
```go
func TextToSpeech(cfg TTSConfig, text string, outputPath ...string) (string, error)
```
- **Input**: `cfg TTSConfig`, `text string`, `outputPath ...string`
- **Output**: `string` (Path file `.wav`), `error`

---

#### 5. `SynthesizeBytes`
Melakukan sintesis suara dan mengembalikan data mentah audio dalam bentuk byte buffer di memori tanpa menyimpan file ke disk.
```go
func SynthesizeBytes(cfg TTSConfig, text, voiceID, lang string) ([]byte, error)
```
- **Input**:
  - `cfg`: `TTSConfig`
  - `text`: `string`
  - `voiceID`: `string`
  - `lang`: `string`
- **Output**:
  - `[]byte`: Byte mentah audio hasil sintesis.
  - `error`: Error jika gagal.

---

### Instance Method (Synthesizer)

#### Inisialisasi Instance
```go
func NewSynthesizer(cfg TTSConfig) Synthesizer
func NewTemplateSynthesizer(cfg TTSConfig) *TemplateSynthesizer
```

#### Method pada `*TemplateSynthesizer`
- `(s *TemplateSynthesizer) SynthesizeToFile(text, outputPath, voiceID, lang string) (string, error)`:
  Menyimpan audio hasil sintesis sebagai file `.wav`.
- `(s *TemplateSynthesizer) Synthesize(text, voiceID, lang string) ([]byte, error)`:
  Mengembalikan byte audio mentah.
- `(s *TemplateSynthesizer) Name() string`:
  Mengembalikan nama model atau identitas synthesizer.

---

### Helper Text Utility

#### `CleanTextForTTS`
Membersihkan format markdown, link URL, indikator attachment, code block, dan whitespace berlebih agar teks terdengar natural saat dibaca oleh TTS.
```go
func CleanTextForTTS(text string) string
```
- **Input**: `text string`
- **Output**: `string` (Teks bersih)

#### `DetectLanguageTTS`
Mendeteksi apakah teks mayoritas berbahasa Indonesia (`"id"`) atau Inggris (`"en"`).
```go
func DetectLanguageTTS(text string) string
```
- **Input**: `text string`
- **Output**: `string` (`"id"` atau `"en"`)

---

## Dokumentasi Fungsi Speech-to-Text (Transcribe)

### Fungsi Top-Level (Transcribe)

#### 1. `Transcribe`
Mengirimkan file audio ke API STT dan mengembalikan hasil transkripsi dalam bentuk `string`.
```go
func Transcribe(cfg STTConfig, audioPath string) (string, error)
```
- **Input**:
  - `cfg`: `STTConfig` (Konfigurasi endpoint STT)
  - `audioPath`: `string` (Path file audio lokal, misalnya `.wav`, `.mp3`, `.m4a`, dll.)
- **Output**:
  - `string`: Teks hasil transkripsi.
  - `error`: Error jika upload atau transkripsi gagal.

---

#### 2. `TranscribeAudio`
Alias untuk fungsi `Transcribe`.
```go
func TranscribeAudio(cfg STTConfig, audioPath string) (string, error)
```
- **Input**: `cfg STTConfig`, `audioPath string`
- **Output**: `string` (Hasil teks), `error`

---

### Instance Method (Transcriber)

#### Inisialisasi Instance
```go
func NewTranscriber(cfg STTConfig) Transcriber
func NewGenericTranscriber(cfg STTConfig) *GenericTranscriber
```

#### Method pada `*GenericTranscriber`
- `(t *GenericTranscriber) Transcribe(audioPath string) (string, error)`:
  Melakukan preprocessing audio (membersihkan background noise dan normalisasi volume) lalu mengunggah ke provider STT dan mengembalikan teks `string`.
- `(t *GenericTranscriber) Name() string`:
  Mengembalikan nama model provider.

---

## Contoh Penggunaan Lengkap

### 1. Contoh Text-to-Speech (Inworld AI)
```go
package main

import (
	"log"
	"text-and-speech"
)

func main() {
	cfg := textandspeech.TTSConfig{
		API_KEY:           "ZHFfcUZQd05PR0o4MDhrbFlfN0Q1cTJYUUNJX0lIUUg6d21IejNlU2JRd2dvREVlLWFZdXlpSA==",
		TTS_URL:           "https://api.inworld.ai/tts/v1/voice",
		TTS_MODEL:         "inworld-tts-2",
		TTS_VOICE:         "Sarah",
		TTS_AUTH_PREFIX:   "Basic ",
		TTS_DECODE:        "base64",
		TTS_BODY_TEMPLATE: `{"text":"{{.Text}}","voiceId":"{{.Voice}}","modelId":"{{.Model}}"}`,
	}

	// Cara 1: Menggunakan fungsi langsung
	wavFile, err := textandspeech.Synthesize(cfg, "Selamat pagi, sistem siap digunakan.", "output.wav")
	if err != nil {
		log.Fatalf("Error sintesis: %v", err)
	}
	log.Println("File .wav berhasil dibuat di:", wavFile)

	// Cara 2: Menggunakan instance Synthesizer
	synth := textandspeech.NewSynthesizer(cfg)
	wavFile2, err := synth.SynthesizeToFile("Pesan kedua dengan instansiasi objek.", "pesan2.wav", "", "")
	if err != nil {
		log.Fatalf("Error sintesis: %v", err)
	}
	log.Println("File .wav kedua dibuat di:", wavFile2)
}
```

---

### 2. Contoh Text-to-Speech (OpenAI TTS)
```go
package main

import (
	"log"
	"text-and-speech"
)

func main() {
	cfg := textandspeech.TTSConfig{
		API_KEY:           "sk-proj-...",
		TTS_URL:           "https://api.openai.com/v1/audio/speech",
		TTS_MODEL:         "tts-1",
		TTS_VOICE:         "alloy",
		TTS_AUTH_PREFIX:   "Bearer ",
		TTS_BODY_TEMPLATE: `{"model":"{{.Model}}","input":"{{.Text}}","voice":"{{.Voice}}"}`,
	}

	wavFile, err := textandspeech.Synthesize(cfg, "Hello from OpenAI TTS", "speech_openai.wav")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Println("File WAV tersimpan di:", wavFile)
}
```

---

### 3. Contoh Speech-to-Text (ElevenLabs Scribe / Groq Whisper)
```go
package main

import (
	"log"
	"text-and-speech"
)

func main() {
	// Contoh ElevenLabs Scribe
	sttCfg := textandspeech.STTConfig{
		API_KEY:         "sk_0cb17ea21407c2ea4b68ce7e03cc67059df7ea882d8c1665",
		STT_URL:         "https://api.elevenlabs.io/v1/speech-to-text",
		STT_MODEL:       "scribe_v2",
		STT_MODEL_FIELD: "model_id",
		STT_HEADER:      "xi-api-key",
	}

	// Transkripsikan file rekaman suara ke teks string
	text, err := textandspeech.Transcribe(sttCfg, "audio_rekaman.wav")
	if err != nil {
		log.Fatalf("Gagal transkripsi: %v", err)
	}

	log.Printf("Hasil Transkripsi: %q\n", text)
}
```
