package edge_tts

import "errors"

// 错误类型定义
var (
	ErrUnknownResponse    = errors.New("unknown response from server")
	ErrUnexpectedResponse = errors.New("unexpected response from server")
	ErrNoAudioReceived    = errors.New("no audio received from server")
	ErrWebSocketError     = errors.New("websocket error")
)

// TTSConfig 定义了文本转语音的配置
type TTSConfig struct {
	Voice  string
	Rate   string
	Volume string
	Pitch  string
	Text   string
}

// TTSChunk 表示一个音频数据块或元数据
type TTSChunk struct {
	Type     string                 // "audio" 或 "WordBoundary"
	Data     []byte                 // 音频数据
	Offset   float64                // 仅用于 WordBoundary
	Duration float64                // 仅用于 WordBoundary
	Text     string                 // 仅用于 WordBoundary
	Metadata map[string]interface{} // 其他元数据
}

// CommunicateState 表示通信状态
type CommunicateState struct {
	PartialText        []byte
	OffsetCompensation int64
	LastDurationOffset int64
	StreamWasCalled    bool
	RequestID          string
	Timestamp          string
	SSML               string
}

// VoiceTag 定义了语音标签
type VoiceTag struct {
	ContentCategories  []string `json:"ContentCategories"`
	VoicePersonalities []string `json:"VoicePersonalities"`
}

// Voice 定义了一个语音
type Voice struct {
	ShortName  string   `json:"ShortName"`
	Gender     string   `json:"Gender"`
	VoiceTag   VoiceTag `json:"VoiceTag"`
	Locale     string   `json:"Locale"`
	LocalName  string   `json:"LocalName"`
	StyleList  []string `json:"StyleList"`
	SampleRate int      `json:"SampleRate"`
}

// NewTTSConfig 创建一个新的 TTSConfig
func NewTTSConfig(text, voice string) *TTSConfig {
	return &TTSConfig{
		Voice:  voice,
		Rate:   "+0%",
		Volume: "+0%",
		Pitch:  "+0Hz",
		Text:   text,
	}
}

// Validate 验证 TTSConfig 的参数
func (c *TTSConfig) Validate() error {
	// TODO: 实现参数验证
	return nil
}
