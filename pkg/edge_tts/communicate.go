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

// Communicate 是与 Edge TTS 服务通信的主要结构
type Communicate struct {
	config *TTSConfig
	client *http.Client
	proxy  string
	wsURL  string
	state  *CommunicateState
}

// NewCommunicate 创建一个新的 Communicate 实例
func NewCommunicate(text, voice string, opts ...Option) *Communicate {
	// 获取系统代理
	proxy := os.Getenv("HTTP_PROXY")
	if proxy == "" {
		proxy = os.Getenv("HTTPS_PROXY")
	}

	config := NewTTSConfig(text, voice)

	// 应用选项
	for _, opt := range opts {
		opt(config)
	}

	// 验证配置
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

// Option 定义配置选项
type Option func(*TTSConfig)

// WithRate 设置语速
func WithRate(rate string) Option {
	return func(c *TTSConfig) {
		c.Rate = rate
	}
}

// WithVolume 设置音量
func WithVolume(volume string) Option {
	return func(c *TTSConfig) {
		c.Volume = volume
	}
}

// WithPitch 设置音调
func WithPitch(pitch string) Option {
	return func(c *TTSConfig) {
		c.Pitch = pitch
	}
}

// Stream 方法实现
func (c *Communicate) Stream(ctx context.Context) (<-chan TTSChunk, error) {
	ch := make(chan TTSChunk, 100)

	go func() {
		defer close(ch)

		// 创建 WebSocket dialer
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

		// 生成连接 ID 和安全令牌
		connID := uuid.New().String()
		secMsGec := generateSecMsGec()

		// 构建完整的 WebSocket URL
		wsURL := fmt.Sprintf("%s&Sec-MS-GEC=%s&Sec-MS-GEC-Version=%s&ConnectionId=%s",
			c.wsURL, secMsGec, SEC_MS_GEC_VERSION, connID)

		// 准备请求头
		headers := http.Header{}
		for k, v := range WSSHeaders {
			headers.Set(k, v)
		}

		// 建立 WebSocket 连接
		conn, _, err := dialer.Dial(wsURL, headers)
		if err != nil {
			ch <- TTSChunk{Type: "error", Data: []byte(err.Error())}
			return
		}

		// 发送命令请求
		cmdReq := fmt.Sprintf("X-Timestamp:%s\r\nContent-Type:application/json; charset=utf-8\r\nPath:speech.config\r\n\r\n{\"context\":{\"synthesis\":{\"audio\":{\"metadataoptions\":{\"sentenceBoundaryEnabled\":\"false\",\"wordBoundaryEnabled\":\"true\"},\"outputFormat\":\"audio-24khz-48kbitrate-mono-mp3\"}}}}\r\n",
			time.Now().UTC().Format(time.RFC1123))

		if err := conn.WriteMessage(websocket.TextMessage, []byte(cmdReq)); err != nil {
			ch <- TTSChunk{Type: "error", Data: []byte(err.Error())}
			return
		}

		// 发送 SSML 请求
		ssmlReq := fmt.Sprintf("X-RequestId:%s\r\nContent-Type:application/ssml+xml\r\nX-Timestamp:%s\r\nPath:ssml\r\n\r\n%s",
			uuid.New().String(),
			time.Now().UTC().Format(time.RFC3339),
			c.createSSML())

		if err := conn.WriteMessage(websocket.TextMessage, []byte(ssmlReq)); err != nil {
			ch <- TTSChunk{Type: "error", Data: []byte(err.Error())}
			return
		}

		// 处理响应数据
		for {
			select {
			case <-ctx.Done():
				// 上下文取消时，优雅关闭连接
				conn.WriteControl(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
					time.Now().Add(time.Second))
				return
			default:
				// 设置读取超时
				conn.SetReadDeadline(time.Now().Add(30 * time.Second))
				messageType, message, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
						return
					}
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						// 超时错误，继续等待
						continue
					}
					if strings.Contains(err.Error(), "broken pipe") {
						// 忽略 broken pipe 错误
						return
					}
					ch <- TTSChunk{Type: "error", Data: []byte(err.Error())}
					return
				}

				// 处理二进制消息（音频数据）
				if messageType == websocket.BinaryMessage {
					// 消息太短，无法包含头部长度
					if len(message) < 2 {
						ch <- TTSChunk{Type: "error", Data: []byte("binary message is too short")}
						return
					}

					// 前两个字节是头部长度
					headerLength := int(binary.BigEndian.Uint16(message[:2]))
					if headerLength > len(message) {
						ch <- TTSChunk{Type: "error", Data: []byte("header length is greater than message length")}
						return
					}

					// 解析头部和数据
					headers, data := getHeadersAndData(message, headerLength)

					// 检查路径
					if path, ok := headers["Path"]; !ok || path != "audio" {
						ch <- TTSChunk{Type: "error", Data: []byte("received binary message, but the path is not audio")}
						return
					}

					// 检查Content-Type
					contentType, ok := headers["Content-Type"]
					if !ok {
						// 如果没有Content-Type，数据必须为空
						if len(data) > 0 {
							// ch <- TTSChunk{Type: "error", Data: []byte("received binary message with no Content-Type, but with data")}
							continue
						}
						continue
					}

					// 检查Content-Type是否为audio/mpeg
					if contentType != "audio/mpeg" {
						ch <- TTSChunk{Type: "error", Data: []byte("received binary message with unexpected Content-Type: " + contentType)}
						continue
					}

					// 如果数据为空，跳过
					if len(data) == 0 {
						ch <- TTSChunk{Type: "error", Data: []byte("received binary message, but it is missing the audio data")}
						continue
					}

					// 发送音频数据
					ch <- TTSChunk{
						Type: "audio",
						Data: data,
					}
					continue
				}

				// 处理文本消息（元数据）
				if messageType == websocket.TextMessage {
					// 解析消息头和数据
					parts := bytes.Split(message, []byte("\r\n\r\n"))
					if len(parts) != 2 {
						continue
					}

					headers := string(parts[0])
					data := parts[1]

					// 检查是否是结束消息
					if strings.Contains(headers, "Path:turn.end") {
						ch <- TTSChunk{
							Type: "end",
							Data: nil,
						}
						// 收到结束消息后，直接关闭连接
						conn.Close()
						return
					}

					// 检查是否是元数据消息
					if strings.Contains(headers, "Path:audio.metadata") {
						// 解析元数据
						var metadata struct {
							Metadata []struct {
								Type string `json:"Type"`
								Data struct {
									Offset   int64  `json:"Offset"`
									Duration int64  `json:"Duration"`
									Text     string `json:"text.Text"`
								} `json:"Data"`
							} `json:"Metadata"`
						}

						if err := json.Unmarshal(data, &metadata); err != nil {
							ch <- TTSChunk{Type: "error", Data: []byte(err.Error())}
							return
						}

						// 处理每个元数据项
						for _, meta := range metadata.Metadata {
							if meta.Type == "WordBoundary" {
								ch <- TTSChunk{
									Type:     "WordBoundary",
									Offset:   float64(meta.Data.Offset),
									Duration: float64(meta.Data.Duration),
									Text:     meta.Data.Text,
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

// Save 方法实现
func (c *Communicate) Save(ctx context.Context, audioPath string, subtitlePath string) error {
	ch, err := c.Stream(ctx)
	if err != nil {
		return err
	}

	// 创建音频文件
	audioFile, err := os.Create(audioPath)
	if err != nil {
		return err
	}
	defer audioFile.Close()

	// 创建字幕文件（如果指定）
	var subtitleFile *os.File
	if subtitlePath != "" {
		subtitleFile, err = os.Create(subtitlePath)
		if err != nil {
			return err
		}
		defer subtitleFile.Close()
	}

	// 创建字幕生成器
	submaker := NewSubMaker()

	audioReceived := false

	for chunk := range ch {
		if chunk.Type == "error" {
			return fmt.Errorf("error during streaming: %s", string(chunk.Data))
		}
		if chunk.Type == "audio" {
			audioReceived = true
			// 将音频数据写入缓冲区
			if _, err := audioFile.Write(chunk.Data); err != nil {
				return err
			}
		} else if chunk.Type == "WordBoundary" && subtitleFile != nil {
			submaker.Feed(chunk)
		}
	}

	// 检查是否收到音频数据
	if !audioReceived {
		return ErrNoAudioReceived
	}

	// 生成字幕文件
	if subtitleFile != nil {
		if _, err := subtitleFile.WriteString(submaker.GetSRT()); err != nil {
			return err
		}
	}

	return nil
}

// createSSML 创建SSML字符串
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

// getHeadersAndData 从二进制消息中提取头部和数据
func getHeadersAndData(data []byte, headerLength int) (map[string]string, []byte) {
	headers := make(map[string]string)

	// 如果数据长度小于2字节，返回空结果
	if len(data) < 2 {
		return headers, nil
	}

	// 如果头部长度小于等于2，说明没有头部数据
	if headerLength <= 2 {
		return headers, data[2:]
	}

	// 解析头部数据（从第3个字节开始，到headerLength结束）
	headerData := data[2:headerLength]

	// 解析头部
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

	// 提取内容数据（从headerLength开始到结束）
	var content []byte
	if len(data) > headerLength {
		// 跳过可能的额外换行符
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
