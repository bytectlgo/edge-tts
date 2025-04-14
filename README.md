# Edge TTS

Edge TTS 是一个基于 Microsoft Edge 浏览器的文本转语音工具，使用 Go 语言实现。它提供了简单易用的 API 来将文本转换为自然流畅的语音。

## 功能特点

- 支持多种语言和语音
- 可调节语速、音量和音调
- 支持获取可用语音列表
- 简单易用的命令行界面
- 支持将语音保存为 MP3 文件

## 安装

```bash
go get github.com/bytectlgo/edge-tts
```

## 使用方法

### 命令行工具

项目提供了一个命令行工具，支持以下功能：

1. 列出所有可用的语音：
```bash
go run .
```

2. 将文本转换为语音（使用默认参数）：
```bash
go run .
```

3. 使用自定义参数转换文本：
```bash
go run . -text "Hello, World!" -voice "en-US-EmmaMultilingualNeural" -output "hello.mp3"
```

### 命令行参数

- `-list`：列出所有可用的语音
- `-text`：要转换的文本（默认为"你好，世界！"）
- `-voice`：要使用的语音（默认为"zh-CN-XiaoxiaoNeural"）
- `-output`：输出文件名（默认为"output.mp3"）

### 在代码中使用

```go
package main

import (
    "context"
    "time"
    "github.com/bytectlgo/edge-tts/pkg/edge_tts"
)

func main() {
    // 创建新的 TTS 配置
    opts := []edge_tts.Option{
        edge_tts.WithRate("+0%"),    // 语速
        edge_tts.WithVolume("+0%"),  // 音量
        edge_tts.WithPitch("+0Hz"),  // 音调
    }

    // 创建新的 Communicate 实例
    comm := edge_tts.NewCommunicate("你好，世界！", "zh-CN-XiaoxiaoNeural", opts...)

    // 设置超时上下文
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // 保存音频到文件
    err := comm.Save(ctx, "output.mp3", "")
    if err != nil {
        panic(err)
    }
}
```

## 常用语音列表

### 中文语音

#### 普通话 (zh-CN)
- `zh-CN-XiaoxiaoNeural`（女声）
- `zh-CN-XiaoyiNeural`（女声）
- `zh-CN-YunjianNeural`（男声）
- `zh-CN-YunxiNeural`（男声）
- `zh-CN-YunxiaNeural`（男声）
- `zh-CN-YunyangNeural`（男声）
- `zh-CN-liaoning-XiaobeiNeural`（女声，辽宁方言）
- `zh-CN-shaanxi-XiaoniNeural`（女声，陕西方言）

#### 香港粤语 (zh-HK)
- `zh-HK-HiuGaaiNeural`（女声）
- `zh-HK-HiuMaanNeural`（女声）
- `zh-HK-WanLungNeural`（男声）

#### 台湾中文 (zh-TW)
- `zh-TW-HsiaoChenNeural`（女声）
- `zh-TW-YunJheNeural`（男声）
- `zh-TW-HsiaoYuNeural`（女声）

### 英语语音

#### 美国英语 (en-US)
- `en-US-AvaNeural`（女声）
- `en-US-AndrewNeural`（男声）
- `en-US-EmmaNeural`（女声）
- `en-US-BrianNeural`（男声）
- `en-US-AnaNeural`（女声）
- `en-US-AndrewMultilingualNeural`（男声，多语言）
- `en-US-AriaNeural`（女声）
- `en-US-AvaMultilingualNeural`（女声，多语言）
- `en-US-BrianMultilingualNeural`（男声，多语言）
- `en-US-ChristopherNeural`（男声）
- `en-US-EmmaMultilingualNeural`（女声，多语言）
- `en-US-EricNeural`（男声）
- `en-US-GuyNeural`（男声）
- `en-US-JennyNeural`（女声）
- `en-US-MichelleNeural`（女声）
- `en-US-RogerNeural`（男声）
- `en-US-SteffanNeural`（男声）

#### 英国英语 (en-GB)
- `en-GB-LibbyNeural`（女声）
- `en-GB-MaisieNeural`（女声）
- `en-GB-RyanNeural`（男声）
- `en-GB-SoniaNeural`（女声）
- `en-GB-ThomasNeural`（男声）

#### 澳大利亚英语 (en-AU)
- `en-AU-NatashaNeural`（女声）
- `en-AU-WilliamNeural`（男声）

#### 加拿大英语 (en-CA)
- `en-CA-ClaraNeural`（女声）
- `en-CA-LiamNeural`（男声）

#### 印度英语 (en-IN)
- `en-IN-NeerjaExpressiveNeural`（女声，富有表现力）
- `en-IN-NeerjaNeural`（女声）
- `en-IN-PrabhatNeural`（男声）

### 其他语言

#### 日语 (ja-JP)
- `ja-JP-KeitaNeural`（男声）
- `ja-JP-NanamiNeural`（女声）

#### 韩语 (ko-KR)
- `ko-KR-HyunsuMultilingualNeural`（男声，多语言）
- `ko-KR-InJoonNeural`（男声）
- `ko-KR-SunHiNeural`（女声）

#### 法语 (fr-FR)
- `fr-FR-VivienneMultilingualNeural`（女声，多语言）
- `fr-FR-RemyMultilingualNeural`（男声，多语言）
- `fr-FR-DeniseNeural`（女声）
- `fr-FR-EloiseNeural`（女声）
- `fr-FR-HenriNeural`（男声）

#### 德语 (de-DE)
- `de-DE-SeraphinaMultilingualNeural`（女声，多语言）
- `de-DE-FlorianMultilingualNeural`（男声，多语言）
- `de-DE-AmalaNeural`（女声）
- `de-DE-ConradNeural`（男声）
- `de-DE-KatjaNeural`（女声）
- `de-DE-KillianNeural`（男声）

#### 西班牙语 (es-ES)
- `es-ES-XimenaNeural`（女声）
- `es-ES-AlvaroNeural`（男声）
- `es-ES-ElviraNeural`（女声）

#### 俄语 (ru-RU)
- `ru-RU-DmitryNeural`（男声）
- `ru-RU-SvetlanaNeural`（女声）

## 注意事项

1. 确保有稳定的网络连接
2. 转换大段文本时可能需要较长时间
3. 某些语音可能需要特定的语言环境支持
4. 多语言语音（Multilingual）支持更多语言，但可能在某些语言上的发音不如专门的语音自然

## 许可证

MIT License 