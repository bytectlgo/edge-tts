package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bytectlgo/edge-tts/pkg/edge_tts"
)

func listVoices() error {
	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", edge_tts.VoiceList, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 添加请求头
	for k, v := range edge_tts.VoiceHeaders {
		req.Header.Set(k, v)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var voices []edge_tts.Voice
	if err := json.Unmarshal(body, &voices); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	// 打印语音列表
	fmt.Printf("找到 %d 个语音:\n", len(voices))
	for _, voice := range voices {
		fmt.Printf("名称: %s\n", voice.ShortName)
		fmt.Printf("性别: %s\n", voice.Gender)
		fmt.Printf("语言: %s\n", voice.Locale)
		fmt.Printf("本地名称: %s\n", voice.LocalName)
		fmt.Printf("采样率: %d\n", voice.SampleRate)
		if len(voice.StyleList) > 0 {
			fmt.Printf("样式: %v\n", voice.StyleList)
		}
		fmt.Println("---")
	}
	return nil
}

func textToSpeech(text, voice, outputFile string) error {
	// 创建新的 TTS 配置
	opts := []edge_tts.Option{
		edge_tts.WithRate("+0%"),
		edge_tts.WithVolume("+0%"),
		edge_tts.WithPitch("+0Hz"),
	}

	// 创建新的 Communicate 实例
	comm := edge_tts.NewCommunicate(text, voice, opts...)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 保存音频到文件
	err := comm.Save(ctx, outputFile, "")
	if err != nil {
		return fmt.Errorf("保存音频失败: %v", err)
	}

	fmt.Printf("音频已保存到 %s\n", outputFile)
	return nil
}

func main() {
	// 定义命令行参数
	listFlag := flag.Bool("list", false, "列出所有可用的语音")
	text := flag.String("text", "你好，世界！", "要转换的文本")
	voice := flag.String("voice", "zh-CN-XiaoxiaoNeural", "要使用的语音")
	output := flag.String("output", "output.mp3", "输出文件名")
	flag.Parse()

	// 根据参数执行相应的功能
	if *listFlag {
		if err := listVoices(); err != nil {
			log.Fatal(err)
		}
		return
	}

	// 检查输出文件是否已存在
	if _, err := os.Stat(*output); err == nil {
		fmt.Printf("警告: 文件 %s 已存在，将被覆盖\n", *output)
	}

	if err := textToSpeech(*text, *voice, *output); err != nil {
		log.Fatal(err)
	}
}
