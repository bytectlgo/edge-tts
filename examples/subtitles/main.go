package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bytectlgo/edge-tts/pkg/edge_tts"
)

func main() {
	// 创建 TTS 实例
	tts := edge_tts.NewCommunicate(
		"Hello World! This is a test of the Edge TTS service with subtitle generation.",
		"en-US-EmmaMultilingualNeural",
		edge_tts.WithRate("+0%"),
		edge_tts.WithVolume("+0%"),
		edge_tts.WithPitch("+0Hz"),
	)

	// 创建上下文
	ctx := context.Background()

	// 保存音频和字幕
	err := tts.Save(ctx, "output.mp3", "output.srt")
	if err != nil {
		log.Fatalf("Error saving audio and subtitles: %v", err)
	}

	fmt.Println("Audio and subtitles have been saved successfully!")
}
