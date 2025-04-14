# Edge TTS

[中文](README_ZH.md) | English

Edge TTS is a command-line tool based on Microsoft Edge's text-to-speech service, supporting multiple languages and voices. It can be used both as a command-line tool and as a Go library in your projects.

## Features

- Support for multiple languages and voices
- Generate audio files
- Generate subtitle files
- Simple and easy-to-use command-line interface

## Audio Preview

You can preview the voices at: [Edge TTS Text to Speech](https://huggingface.co/spaces/innoai/Edge-TTS-Text-to-Speech)

## Installation

```bash
go install github.com/bytectlgo/edge-tts@latest
```

## Usage

### List All Available Voices

```bash
edge-tts -list-voices
```

### Text to Speech

Basic usage:

```bash
edge-tts -text "Text to convert" -voice "zh-CN-XiaoxiaoNeural" -write-media output.mp3
```

Parameters:

- `-text`: Text content to convert
- `-voice`: Voice to use (default is "zh-CN-XiaoxiaoNeural")
- `-write-media`: Output audio filename
- `-write-subtitles`: Output subtitle filename

### Examples

1. Generate audio file:

```bash
edge-tts -text "Hello, World!" -voice "zh-CN-XiaoxiaoNeural" -write-media hello.mp3
```

2. Generate audio and subtitle files:

```bash
edge-tts -text "Hello, World!" -voice "zh-CN-XiaoxiaoNeural" -write-media hello.mp3 -write-subtitles hello.srt
```

## Using as a Go Library

You can also use this package as a Go library in your projects:

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/bytectlgo/edge-tts/pkg/edge_tts"
)

func main() {
    // Create a new TTS client with text and voice
    comm := edge_tts.NewCommunicate(
        "Hello, World!",
        "zh-CN-XiaoxiaoNeural",
        edge_tts.WithRate("+0%"),
        edge_tts.WithVolume("+0%"),
        edge_tts.WithPitch("+0Hz"),
    )

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Generate audio file
    err := comm.Save(ctx, "output.mp3", "")
    if err != nil {
        panic(err)
    }

    // Generate audio file with subtitles
    err = comm.Save(ctx, "output.mp3", "output.srt")
    if err != nil {
        panic(err)
    }

    // Stream audio data
    ch, err := comm.Stream(ctx)
    if err != nil {
        panic(err)
    }

    for chunk := range ch {
        switch chunk.Type {
        case "audio":
            // Process audio data
            fmt.Printf("Received audio chunk of size: %d\n", len(chunk.Data))
        case "WordBoundary":
            // Process word boundary metadata
            fmt.Printf("Word: %s, Offset: %f, Duration: %f\n", 
                chunk.Text, chunk.Offset, chunk.Duration)
        case "error":
            fmt.Printf("Error: %s\n", string(chunk.Data))
        }
    }
}
```

## Supported Voices

This project supports multiple languages and voices, including but not limited to:

- Chinese (Mandarin, Cantonese, Taiwanese)
- English (US, UK, Australian, etc.)
- Japanese
- Korean
- French
- German
- Spanish
- And more...

For a complete list of voices, use the `-list-voices` command.

## License

MIT License 

## References

This project is inspired by and references the following project:
- [edge-tts](https://github.com/rany2/edge-tts) - A Python module that allows you to use Microsoft Edge's online text-to-speech service 