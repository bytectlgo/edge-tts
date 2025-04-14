package edge_tts

import "errors"

// Error type definitions
var (
	ErrUnknownResponse    = errors.New("unknown response from server")
	ErrUnexpectedResponse = errors.New("unexpected response from server")
	ErrNoAudioReceived    = errors.New("no audio received from server")
	ErrWebSocketError     = errors.New("websocket error")
)

// TTSConfig defines the text-to-speech configuration
type TTSConfig struct {
	Voice  string
	Rate   string
	Volume string
	Pitch  string
	Text   string
}

// TTSChunk represents an audio data chunk or metadata
type TTSChunk struct {
	Type     string                 // "audio" or "WordBoundary"
	Data     []byte                 // Audio data
	Offset   float64                // Only used for WordBoundary
	Duration float64                // Only used for WordBoundary
	Text     string                 // Only used for WordBoundary
	Metadata map[string]interface{} // Other metadata
}

// CommunicateState represents the communication state
type CommunicateState struct {
	PartialText        []byte
	OffsetCompensation int64
	LastDurationOffset int64
	StreamWasCalled    bool
	RequestID          string
	Timestamp          string
	SSML               string
}

// VoiceTag defines the voice tag
type VoiceTag struct {
	ContentCategories  []string `json:"ContentCategories"`
	VoicePersonalities []string `json:"VoicePersonalities"`
}

// Voice represents a voice
type Voice struct {
	Name       string   `json:"Name"`
	ShortName  string   `json:"ShortName"`
	Gender     string   `json:"Gender"`
	Locale     string   `json:"Locale"`
	LocalName  string   `json:"LocalName"`
	SampleRate int      `json:"SampleRate"`
	StyleList  []string `json:"StyleList"`
	VoiceTag   VoiceTag `json:"VoiceTag"`
}

// NewTTSConfig creates a new TTSConfig
func NewTTSConfig(text, voice string) *TTSConfig {
	return &TTSConfig{
		Voice:  voice,
		Rate:   "+0%",
		Volume: "+0%",
		Pitch:  "+0Hz",
		Text:   text,
	}
}

// Validate validates the TTSConfig parameters
func (c *TTSConfig) Validate() error {
	// TODO: Implement parameter validation
	return nil
}
