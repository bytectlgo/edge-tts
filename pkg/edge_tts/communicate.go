package edge_tts

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Communicate is the main structure for communicating with Edge TTS service
type Communicate struct {
	config *TTSConfig
	client *http.Client
	proxy  string
	wsURL  string
	state  *CommunicateState
}

// NewCommunicate creates a new Communicate instance
func NewCommunicate(text, voice string, opts ...Option) *Communicate {
	// Get system proxy
	proxy := os.Getenv("HTTP_PROXY")
	if proxy == "" {
		proxy = os.Getenv("HTTPS_PROXY")
	}

	config := NewTTSConfig(text, voice)

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		panic(err)
	}

	return &Communicate{
		config: config,
		client: &http.Client{},
		proxy:  proxy,
		wsURL:  WSSURL,
		state: &CommunicateState{
			PartialText: []byte(text),
		},
	}
}

// Option defines configuration options
type Option func(*TTSConfig)

// WithRate sets the speech rate
func WithRate(rate string) Option {
	return func(c *TTSConfig) {
		c.Rate = rate
	}
}

// WithVolume sets the volume
func WithVolume(volume string) Option {
	return func(c *TTSConfig) {
		c.Volume = volume
	}
}

// WithPitch sets the pitch
func WithPitch(pitch string) Option {
	return func(c *TTSConfig) {
		c.Pitch = pitch
	}
}

// Stream method implementation
func (c *Communicate) Stream(ctx context.Context) (<-chan TTSChunk, error) {
	ch := make(chan TTSChunk, 100)

	go func() {
		defer close(ch)

		// Create WebSocket dialer
		dialer := websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS12,
				MaxVersion:         tls.VersionTLS13,
				CipherSuites: []uint16{
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
					tls.TLS_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				},
				NextProtos: []string{"http/1.1"},
			},
			EnableCompression: true,
		}

		// Generate connection ID and security token
		connID := uuid.New().String()
		secMsGec := generateSecMsGec()

		// Build complete WebSocket URL (参数顺序与 Python 一致)
		wsURL := fmt.Sprintf("%s&ConnectionId=%s&Sec-MS-GEC=%s&Sec-MS-GEC-Version=%s",
			c.wsURL, connID, secMsGec, SEC_MS_GEC_VERSION)

		// Prepare request headers
		headers := http.Header{}
		for k, v := range WSSHeaders {
			headers.Set(k, v)
		}

		// 添加 MUID Cookie (关键修复!)
		headers = headersWithMUID(headers)

		// Establish WebSocket connection
		conn, _, err := dialer.Dial(wsURL, headers)
		if err != nil {
			ch <- TTSChunk{Type: "error", Data: []byte(err.Error())}
			return
		}

		// Send command request (使用 JavaScript 风格的时间戳)
		cmdReq := fmt.Sprintf("X-Timestamp:%s\r\nContent-Type:application/json; charset=utf-8\r\nPath:speech.config\r\n\r\n{\"context\":{\"synthesis\":{\"audio\":{\"metadataoptions\":{\"sentenceBoundaryEnabled\":\"false\",\"wordBoundaryEnabled\":\"true\"},\"outputFormat\":\"audio-24khz-48kbitrate-mono-mp3\"}}}}\r\n",
			dateToString())

		if err := conn.WriteMessage(websocket.TextMessage, []byte(cmdReq)); err != nil {
			ch <- TTSChunk{Type: "error", Data: []byte(err.Error())}
			return
		}

		// Send SSML request (时间戳格式需要加 Z 后缀)
		ssmlReq := fmt.Sprintf("X-RequestId:%s\r\nContent-Type:application/ssml+xml\r\nX-Timestamp:%sZ\r\nPath:ssml\r\n\r\n%s",
			uuid.New().String(),
			dateToString(),
			c.createSSML())

		if err := conn.WriteMessage(websocket.TextMessage, []byte(ssmlReq)); err != nil {
			ch <- TTSChunk{Type: "error", Data: []byte(err.Error())}
			return
		}

		// Process response data
		for {
			select {
			case <-ctx.Done():
				// Gracefully close connection when context is canceled
				conn.WriteControl(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
					time.Now().Add(time.Second))
				return
			default:
				// Set read timeout
				conn.SetReadDeadline(time.Now().Add(30 * time.Second))
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
						return
					}
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						// Timeout error, continue waiting
						continue
					}
					if strings.Contains(err.Error(), "broken pipe") {
						// Ignore broken pipe error
						return
					}
					ch <- TTSChunk{Type: "error", Data: []byte(err.Error())}
					return
				}

				// Process binary message (audio data)
				if messageType == websocket.BinaryMessage {
					// Message too short to contain header length
					if len(message) < 2 {
						ch <- TTSChunk{Type: "error", Data: []byte("binary message is too short")}
						return
					}

					// First two bytes are header length
					headerLength := int(binary.BigEndian.Uint16(message[:2]))
					if headerLength > len(message) {
						ch <- TTSChunk{Type: "error", Data: []byte("header length is greater than message length")}
						return
					}

					// Parse headers and data
					headers, data := getHeadersAndData(message, headerLength)

					// Check path
					if path, ok := headers["Path"]; !ok || path != "audio" {
						ch <- TTSChunk{Type: "error", Data: []byte("received binary message, but the path is not audio")}
						return
					}

					// Check Content-Type
					contentType, ok := headers["Content-Type"]
					if !ok {
						// If no Content-Type, data must be empty
						if len(data) > 0 {
							// ch <- TTSChunk{Type: "error", Data: []byte("received binary message with no Content-Type, but with data")}
							continue
						}
						continue
					}

					// Check if Content-Type is audio/mpeg
					if contentType != "audio/mpeg" {
						ch <- TTSChunk{Type: "error", Data: []byte("received binary message with unexpected Content-Type: " + contentType)}
						continue
					}

					// Skip if data is empty
					if len(data) == 0 {
						ch <- TTSChunk{Type: "error", Data: []byte("received binary message, but it is missing the audio data")}
						continue
					}

					// Send audio data
					ch <- TTSChunk{
						Type: "audio",
						Data: data,
					}
					continue
				}

				// Process text message (metadata)
				if messageType == websocket.TextMessage {
					// Parse message headers and data
					parts := bytes.Split(message, []byte("\r\n\r\n"))
					if len(parts) != 2 {
						continue
					}

					headers := string(parts[0])
					data := parts[1]

					// Check if it's an end message
					if strings.Contains(headers, "Path:turn.end") {
						ch <- TTSChunk{
							Type: "end",
							Data: nil,
						}
						// Close connection after receiving end message
						conn.Close()
						return
					}

					// Check if it's a metadata message
					if strings.Contains(headers, "Path:audio.metadata") {
						// Parse metadata
						var metadata struct {
							Metadata []struct {
								Type string `json:"Type"`
								Data struct {
									Offset   int64 `json:"Offset"`
									Duration int64 `json:"Duration"`
									Text     struct {
										Text string `json:"Text"`
									} `json:"Text"`
								} `json:"Data"`
							} `json:"Metadata"`
						}

						if err := json.Unmarshal(data, &metadata); err != nil {
							ch <- TTSChunk{Type: "error", Data: []byte(err.Error())}
							return
						}

						// Process each metadata item
						for _, meta := range metadata.Metadata {
							if meta.Type == "WordBoundary" {
								// 确保文本内容不为空
								if meta.Data.Text.Text == "" {
									continue
								}
								ch <- TTSChunk{
									Type:     "WordBoundary",
									Offset:   float64(meta.Data.Offset),
									Duration: float64(meta.Data.Duration),
									Text:     meta.Data.Text.Text,
								}
							}
						}
					}
				}
			}
		}
	}()

	return ch, nil
}

// Save method implementation
func (c *Communicate) Save(ctx context.Context, audioPath string, subtitlePath string) error {
	ch, err := c.Stream(ctx)
	if err != nil {
		return err
	}

	// Create audio file
	audioFile, err := os.Create(audioPath)
	if err != nil {
		return err
	}
	defer audioFile.Close()

	// Create subtitle file (if specified)
	var subtitleFile *os.File
	if subtitlePath != "" {
		subtitleFile, err = os.Create(subtitlePath)
		if err != nil {
			return err
		}
		defer subtitleFile.Close()
	}

	// Create subtitle generator
	submaker := NewSubMaker()

	audioReceived := false

	for chunk := range ch {
		if chunk.Type == "error" {
			return fmt.Errorf("error during streaming: %s", string(chunk.Data))
		}
		if chunk.Type == "audio" {
			audioReceived = true
			// Write audio data to buffer
			if _, err := audioFile.Write(chunk.Data); err != nil {
				return err
			}
		} else if chunk.Type == "WordBoundary" && subtitleFile != nil {
			if err := submaker.Feed(chunk); err != nil {
				return fmt.Errorf("error feeding chunk: %v", err)
			}
		}
	}

	// Check if audio data was received
	if !audioReceived {
		return ErrNoAudioReceived
	}

	// Generate subtitle file
	if subtitleFile != nil {
		if _, err := subtitleFile.WriteString(submaker.GetSRT()); err != nil {
			return err
		}
	}

	return nil
}

// createSSML creates SSML string
func (c *Communicate) createSSML() string {
	return fmt.Sprintf(
		"<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='en-US'>"+
			"<voice name='%s'>"+
			"<prosody pitch='%s' rate='%s' volume='%s'>"+
			"%s"+
			"</prosody>"+
			"</voice>"+
			"</speak>",
		c.config.Voice,
		c.config.Pitch,
		c.config.Rate,
		c.config.Volume,
		c.config.Text,
	)
}

// getHeadersAndData extracts headers and data from binary message
func getHeadersAndData(data []byte, headerLength int) (map[string]string, []byte) {
	headers := make(map[string]string)

	// If data length is less than 2 bytes, return empty result
	if len(data) < 2 {
		return headers, nil
	}

	// If header length is less than or equal to 2, no header data
	if headerLength <= 2 {
		return headers, data[2:]
	}

	// Parse header data (from 3rd byte to headerLength)
	headerData := data[2:headerLength]

	// Parse headers
	lines := bytes.Split(headerData, []byte("\r\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := bytes.SplitN(line, []byte(":"), 2)
		if len(parts) == 2 {
			key := string(bytes.TrimSpace(parts[0]))
			value := string(bytes.TrimSpace(parts[1]))
			headers[key] = value
		}
	}

	// Extract content data (from headerLength to end)
	var content []byte
	if len(data) > headerLength {
		// Skip possible extra newlines
		start := headerLength
		for start < len(data) && (data[start] == '\r' || data[start] == '\n') {
			start++
		}
		if start < len(data) {
			content = make([]byte, len(data)-start)
			copy(content, data[start:])
		}
	}

	return headers, content
}
