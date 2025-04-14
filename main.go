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
	"strings"
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

	// 打印表头
	fmt.Printf("%-35s %-9s %-22s %-35s\n", "Name", "Gender", "ContentCategories", "VoicePersonalities")
	fmt.Println(strings.Repeat("-", 100))

	// 打印语音列表
	for _, voice := range voices {
		personalities := strings.Join(voice.StyleList, ", ")
		if personalities == "" {
			personalities = "Friendly, Positive"
		}
		fmt.Printf("%-35s %-9s %-22s %-35s\n",
			voice.ShortName,
			voice.Gender,
			"General",
			personalities)
	}
	return nil
}

func textToSpeech(text, voice, outputFile, subtitleFile string, rate, volume, pitch string) error {
	// 创建新的 TTS 配置
	opts := []edge_tts.Option{
		edge_tts.WithRate(rate),
		edge_tts.WithVolume(volume),
		edge_tts.WithPitch(pitch),
	}

	// 创建新的 Communicate 实例
	comm := edge_tts.NewCommunicate(text, voice, opts...)

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 保存音频到文件
	err := comm.Save(ctx, outputFile, subtitleFile)
	if err != nil {
		return fmt.Errorf("保存音频失败: %v", err)
	}

	if outputFile != "" {
		fmt.Printf("音频已保存到 %s\n", outputFile)
	}
	if subtitleFile != "" {
		fmt.Printf("字幕已保存到 %s\n", subtitleFile)
	}
	return nil
}

func main() {
	// 定义命令行参数
	listVoicesFlag := flag.Bool("list-voices", false, "列出所有可用的语音")
	text := flag.String("text", "", "要转换的文本")
	voice := flag.String("voice", "zh-CN-XiaoxiaoNeural", "要使用的语音")
	outputMedia := flag.String("write-media", "", "输出音频文件名")
	outputSubtitles := flag.String("write-subtitles", "", "输出字幕文件名")
	rate := flag.String("rate", "+0%", "语速调整")
	volume := flag.String("volume", "+0%", "音量调整")
	pitch := flag.String("pitch", "+0Hz", "音调调整")
	flag.Parse()

	// 根据参数执行相应的功能
	if *listVoicesFlag {
		if err := listVoices(); err != nil {
			log.Fatal(err)
		}
		return
	}

	// 检查必要参数
	if *text == "" {
		log.Fatal("错误: 必须提供 --text 参数")
	}
	if *outputMedia == "" && *outputSubtitles == "" {
		log.Fatal("错误: 必须提供 --write-media 或 --write-subtitles 参数")
	}

	// 检查输出文件是否已存在
	if *outputMedia != "" {
		if _, err := os.Stat(*outputMedia); err == nil {
			fmt.Printf("警告: 文件 %s 已存在，将被覆盖\n", *outputMedia)
		}
	}
	if *outputSubtitles != "" {
		if _, err := os.Stat(*outputSubtitles); err == nil {
			fmt.Printf("警告: 文件 %s 已存在，将被覆盖\n", *outputSubtitles)
		}
	}

	if err := textToSpeech(*text, *voice, *outputMedia, *outputSubtitles, *rate, *volume, *pitch); err != nil {
		log.Fatal(err)
	}
}
