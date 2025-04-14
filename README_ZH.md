# Edge TTS

[English](README.md) | 中文

Edge TTS 是一个基于 Microsoft Edge 文本转语音服务的命令行工具，支持多种语言和声音。它既可以作为命令行工具使用，也可以作为 Go 库在您的项目中使用。

## 功能特点

- 支持多种语言和声音
- 可生成音频文件
- 可生成字幕文件
- 简单易用的命令行界面

## 音频试听

您可以在以下地址试听各种声音：[Edge TTS 文本转语音](https://huggingface.co/spaces/innoai/Edge-TTS-Text-to-Speech)

## 安装

### 使用 Go 安装

```bash
go install github.com/bytectlgo/edge-tts@latest
```

### 使用 Homebrew 安装

1. 添加 tap 源：
```bash
brew tap bytectlgo/homebrew-tap
```

2. 安装 edge-tts：
```bash
brew install edge-tts
```

## 使用方法

### 列出所有可用语音

```bash
edge-tts -list-voices
```

### 文本转语音

基本用法：

```bash
edge-tts -text "要转换的文本" -voice "zh-CN-XiaoxiaoNeural" -write-media output.mp3
```

参数说明：

- `-text`: 要转换的文本内容
- `-voice`: 要使用的语音（默认为 "zh-CN-XiaoxiaoNeural"）
- `-write-media`: 输出音频文件名
- `-write-subtitles`: 输出字幕文件名

### 示例

1. 生成音频文件：

```bash
edge-tts -text "你好，世界！" -voice "zh-CN-XiaoxiaoNeural" -write-media hello.mp3
```

2. 生成音频和字幕文件：

```bash
edge-tts -text "你好，世界！" -voice "zh-CN-XiaoxiaoNeural" -write-media hello.mp3 -write-subtitles hello.srt
```

## 作为 Go 库使用

您也可以在您的 Go 项目中将此包作为库使用：

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/bytectlgo/edge-tts/pkg/edge_tts"
)

func main() {
    // 创建新的 TTS 客户端，设置文本和语音
    comm := edge_tts.NewCommunicate(
        "你好，世界！",
        "zh-CN-XiaoxiaoNeural",
        edge_tts.WithRate("+0%"),
        edge_tts.WithVolume("+0%"),
        edge_tts.WithPitch("+0Hz"),
    )

    // 创建带超时的上下文
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // 生成音频文件
    err := comm.Save(ctx, "output.mp3", "")
    if err != nil {
        panic(err)
    }

    // 生成带字幕的音频文件
    err = comm.Save(ctx, "output.mp3", "output.srt")
    if err != nil {
        panic(err)
    }

    // 流式处理音频数据
    ch, err := comm.Stream(ctx)
    if err != nil {
        panic(err)
    }

    for chunk := range ch {
        switch chunk.Type {
        case "audio":
            // 处理音频数据
            fmt.Printf("收到音频数据块，大小：%d\n", len(chunk.Data))
        case "WordBoundary":
            // 处理字幕元数据
            fmt.Printf("单词：%s，偏移：%f，持续时间：%f\n", 
                chunk.Text, chunk.Offset, chunk.Duration)
        case "error":
            fmt.Printf("错误：%s\n", string(chunk.Data))
        }
    }
}
```

## 支持的语音

本项目支持多种语言和声音，包括但不限于：

- 中文（普通话、粤语、台湾话）
- 英语（美国、英国、澳大利亚等）
- 日语
- 韩语
- 法语
- 德语
- 西班牙语
- 等等...

完整的语音列表可以通过 `-list-voices` 命令查看。

## 许可证

MIT License 

## 参考

本项目参考了以下项目：
- [edge-tts](https://github.com/rany2/edge-tts) - 一个允许使用 Microsoft Edge 在线文本转语音服务的 Python 模块 