package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bytectlgo/edge-tts/pkg/edge_tts"
)

func main() {
	// 创建新的 TTS 配置
	opts := []edge_tts.Option{
		edge_tts.WithRate("+0%"),
		edge_tts.WithVolume("+0%"),
		edge_tts.WithPitch("+0Hz"),
	}
	// 创建新的 Communicate 实例
	comm := edge_tts.NewCommunicate("你好，世界, 你好，中国！", "zh-CN-XiaoxiaoNeural", opts...)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 保存音频到文件
	err := comm.Save(ctx, "output.mp3", "")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("音频已保存到 output.mp3")

}
