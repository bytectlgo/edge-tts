package edge_tts

import (
	"bytes"
	"context"
	"os"
	"testing"
)

// TestNewCommunicate 测试 NewCommunicate 函数
func TestNewCommunicate(t *testing.T) {
	// 保存原始环境变量
	originalHTTPProxy := os.Getenv("HTTP_PROXY")
	originalHTTPSProxy := os.Getenv("HTTPS_PROXY")
	defer func() {
		os.Setenv("HTTP_PROXY", originalHTTPProxy)
		os.Setenv("HTTPS_PROXY", originalHTTPSProxy)
	}()

	// 测试用例
	tests := []struct {
		name      string
		text      string
		voice     string
		proxy     string
		wantProxy string
	}{
		{
			name:      "基本创建",
			text:      "Hello, world!",
			voice:     "en-US-JennyNeural",
			proxy:     "",
			wantProxy: "",
		},
		{
			name:      "使用 HTTP_PROXY",
			text:      "Hello, world!",
			voice:     "en-US-JennyNeural",
			proxy:     "http://proxy.example.com:8080",
			wantProxy: "http://proxy.example.com:8080",
		},
		{
			name:      "使用 HTTPS_PROXY",
			text:      "Hello, world!",
			voice:     "en-US-JennyNeural",
			proxy:     "https://proxy.example.com:8443",
			wantProxy: "https://proxy.example.com:8443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置环境变量
			if tt.proxy != "" {
				os.Setenv("HTTP_PROXY", tt.proxy)
				os.Setenv("HTTPS_PROXY", "")
			} else {
				os.Setenv("HTTP_PROXY", "")
				os.Setenv("HTTPS_PROXY", "")
			}

			// 创建 Communicate 实例
			c := NewCommunicate(tt.text, tt.voice)

			// 验证结果
			if c.config.Text != tt.text {
				t.Errorf("NewCommunicate() text = %v, want %v", c.config.Text, tt.text)
			}
			if c.config.Voice != tt.voice {
				t.Errorf("NewCommunicate() voice = %v, want %v", c.config.Voice, tt.voice)
			}
			if c.proxy != tt.wantProxy {
				t.Errorf("NewCommunicate() proxy = %v, want %v", c.proxy, tt.wantProxy)
			}
			if c.wsURL != WSSURL {
				t.Errorf("NewCommunicate() wsURL = %v, want %v", c.wsURL, WSSURL)
			}
			if !bytes.Equal(c.state.PartialText, []byte(tt.text)) {
				t.Errorf("NewCommunicate() state.PartialText = %v, want %v", c.state.PartialText, []byte(tt.text))
			}
		})
	}
}

// TestOptionFunctions 测试选项函数
func TestOptionFunctions(t *testing.T) {
	// 测试 WithRate
	rate := "+10%"
	config := NewTTSConfig("test", "en-US-JennyNeural")
	WithRate(rate)(config)
	if config.Rate != rate {
		t.Errorf("WithRate() = %v, want %v", config.Rate, rate)
	}

	// 测试 WithVolume
	volume := "+20%"
	config = NewTTSConfig("test", "en-US-JennyNeural")
	WithVolume(volume)(config)
	if config.Volume != volume {
		t.Errorf("WithVolume() = %v, want %v", config.Volume, volume)
	}

	// 测试 WithPitch
	pitch := "+5Hz"
	config = NewTTSConfig("test", "en-US-JennyNeural")
	WithPitch(pitch)(config)
	if config.Pitch != pitch {
		t.Errorf("WithPitch() = %v, want %v", config.Pitch, pitch)
	}
}

// TestCreateSSML 测试 createSSML 方法
func TestCreateSSML(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		voice  string
		rate   string
		volume string
		pitch  string
		want   string
	}{
		{
			name:   "基本 SSML",
			text:   "Hello, world!",
			voice:  "en-US-JennyNeural",
			rate:   "+0%",
			volume: "+0%",
			pitch:  "+0Hz",
			want:   "<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='en-US'><voice name='en-US-JennyNeural'><prosody pitch='+0Hz' rate='+0%' volume='+0%'>Hello, world!</prosody></voice></speak>",
		},
		{
			name:   "自定义参数",
			text:   "Hello, world!",
			voice:  "zh-CN-XiaoxiaoNeural",
			rate:   "+10%",
			volume: "+20%",
			pitch:  "+5Hz",
			want:   "<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='en-US'><voice name='zh-CN-XiaoxiaoNeural'><prosody pitch='+5Hz' rate='+10%' volume='+20%'>Hello, world!</prosody></voice></speak>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Communicate{
				config: &TTSConfig{
					Text:   tt.text,
					Voice:  tt.voice,
					Rate:   tt.rate,
					Volume: tt.volume,
					Pitch:  tt.pitch,
				},
			}

			got := c.createSSML()
			if got != tt.want {
				t.Errorf("createSSML() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetHeadersAndData 测试 getHeadersAndData 函数
func TestGetHeadersAndData(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		headerLength int
		wantHeaders  map[string]string
		wantContent  []byte
	}{
		{
			name: "基本头部和数据",
			data: []byte{
				0x00, 0x1E, // 头部长度 (30 字节)
				// 头部数据
				'P', 'a', 't', 'h', ':', ' ', 'a', 'u', 'd', 'i', 'o', '\r', '\n',
				'C', 'o', 'n', 't', 'e', 'n', 't', '-', 'T', 'y', 'p', 'e', ':', ' ', 'a', 'u', 'd', 'i', 'o', '/', 'm', 'p', 'e', 'g', '\r', '\n',
				// 内容数据
				'H', 'e', 'l', 'l', 'o',
			},
			headerLength: 41, // 2字节长度 + 39字节头部数据
			wantHeaders: map[string]string{
				"Path":         "audio",
				"Content-Type": "audio/mpeg",
			},
			wantContent: []byte("Hello"),
		},
		{
			name: "空头部",
			data: []byte{
				0x00, 0x02, // 头部长度 (2 字节)
				// 内容数据
				'H', 'e', 'l', 'l', 'o',
			},
			headerLength: 2, // 只有长度信息
			wantHeaders:  map[string]string{},
			wantContent:  []byte("Hello"),
		},
		{
			name: "多个头部",
			data: []byte{
				0x00, 0x3C, // 头部长度 (60 字节)
				// 头部数据
				'P', 'a', 't', 'h', ':', ' ', 'a', 'u', 'd', 'i', 'o', '\r', '\n',
				'C', 'o', 'n', 't', 'e', 'n', 't', '-', 'T', 'y', 'p', 'e', ':', ' ', 'a', 'u', 'd', 'i', 'o', '/', 'm', 'p', 'e', 'g', '\r', '\n',
				'X', '-', 'T', 'i', 'm', 'e', 's', 't', 'a', 'm', 'p', ':', ' ', '1', '2', '3', '\r', '\n',
				// 内容数据
				'H', 'e', 'l', 'l', 'o',
			},
			headerLength: 57, // 2字节长度 + 55字节头部数据
			wantHeaders: map[string]string{
				"Path":         "audio",
				"Content-Type": "audio/mpeg",
				"X-Timestamp":  "123",
			},
			wantContent: []byte("Hello"),
		},
		{
			name:         "数据太短",
			data:         []byte{0x00},
			headerLength: 1,
			wantHeaders:  map[string]string{},
			wantContent:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 打印测试数据以便调试
			t.Logf("Test data: %v", tt.data)
			t.Logf("Header length: %d", tt.headerLength)

			gotHeaders, gotContent := getHeadersAndData(tt.data, tt.headerLength)

			// 打印实际结果以便调试
			t.Logf("Got headers: %v", gotHeaders)
			t.Logf("Got content: %v", gotContent)

			// 检查头部
			if len(gotHeaders) != len(tt.wantHeaders) {
				t.Errorf("getHeadersAndData() headers count = %v, want %v", len(gotHeaders), len(tt.wantHeaders))
				t.Errorf("Got headers: %v", gotHeaders)
				t.Errorf("Want headers: %v", tt.wantHeaders)
			}
			for k, v := range tt.wantHeaders {
				if gotHeaders[k] != v {
					t.Errorf("getHeadersAndData() headers[%s] = %v, want %v", k, gotHeaders[k], v)
				}
			}

			// 检查内容
			if !bytes.Equal(gotContent, tt.wantContent) {
				t.Errorf("getHeadersAndData() content = %v, want %v", gotContent, tt.wantContent)
				t.Errorf("Got content string: %s", string(gotContent))
				t.Errorf("Want content string: %s", string(tt.wantContent))
			}
		})
	}
}

// TestCommunicateSave 测试 Save 方法
func TestCommunicateSave(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "edge-tts-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建临时文件路径
	audioPath := tempDir + "/test.mp3"
	// 创建 Communicate 实例
	c := NewCommunicate("Hello, world!", "en-US-JennyNeural")

	// 由于 Save 方法依赖于 Stream 方法，而 Stream 方法需要网络连接，
	// 这里我们只测试文件创建和错误处理逻辑
	// 在实际应用中，应该使用 mock 来模拟 Stream 方法

	// 测试无效的音频路径
	err = c.Save(context.Background(), "/invalid/path/test.mp3", "")
	if err == nil {
		t.Error("Save() with invalid audio path should return error")
	}

	// 测试无效的字幕路径
	err = c.Save(context.Background(), audioPath, "/invalid/path/test.srt")
	if err == nil {
		t.Error("Save() with invalid subtitle path should return error")
	}

	// 注意：完整的 Save 方法测试需要 mock Stream 方法，
	// 这超出了简单测试的范围，需要更复杂的测试框架
}
