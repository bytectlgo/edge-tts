# Edge TTS

[English](README.md) | 中文

Edge TTS 是一个基于 Microsoft Edge 文本转语音服务的命令行工具，支持多种语言和声音。

## 功能特点

- 支持多种语言和声音
- 可生成音频文件
- 可生成字幕文件
- 简单易用的命令行界面

## 音频试听

您可以在以下地址试听各种声音：[Edge TTS 文本转语音](https://huggingface.co/spaces/innoai/Edge-TTS-Text-to-Speech)

## 安装

```bash
go install github.com/bytectlgo/edge-tts@latest
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