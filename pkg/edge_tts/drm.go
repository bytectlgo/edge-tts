package edge_tts

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// clockSkewSeconds 存储时钟偏差（秒）
var clockSkewSeconds float64

// clockSkewLock 用于保护 clockSkewSeconds 的并发访问
var clockSkewLock sync.RWMutex

// adjustClockSkew 调整时钟偏差
func adjustClockSkew(date string) error {
	// 解析 RFC 2616 格式的日期
	serverTime, err := parseRFC2616Date(date)
	if err != nil {
		return err
	}

	// 计算时钟偏差
	clientTime := time.Now().UTC().Unix()
	clockSkewLock.Lock()
	clockSkewSeconds = float64(serverTime - clientTime)
	clockSkewLock.Unlock()

	return nil
}

// parseRFC2616Date 解析 RFC 2616 格式的日期
func parseRFC2616Date(date string) (int64, error) {
	// 尝试解析 RFC 2616 格式的日期
	t, err := time.Parse(time.RFC1123, date)
	if err != nil {
		return 0, err
	}

	return t.Unix(), nil
}

// getUnixTimestamp 获取当前 Unix 时间戳（秒）
func getUnixTimestamp() int64 {
	clockSkewLock.RLock()
	skew := clockSkewSeconds
	clockSkewLock.RUnlock()
	return time.Now().UTC().Unix() + int64(skew)
}

// generateSecMsGec 生成 Sec-MS-GEC token
func generateSecMsGec() string {
	// 获取当前时间戳（Unix 时间戳，秒）
	timestamp := getUnixTimestamp()

	// 转换为 Windows 文件时间（从 1601-01-01 开始的 100 纳秒间隔）
	ticks := (timestamp + 11644473600) * 10000000

	// 向下取整到最近的 5 分钟（300 秒）
	ticks = ticks - (ticks % (300 * 10000000))

	// 创建要哈希的字符串
	strToHash := fmt.Sprintf("%d%s", ticks, TrustedClientToken)

	// 计算 SHA256 哈希
	hash := sha256.Sum256([]byte(strToHash))

	// 返回大写的十六进制字符串
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// handleClientResponseError 处理客户端响应错误
func handleClientResponseError(resp *http.Response) error {
	// 获取服务器日期
	date := resp.Header.Get("Date")
	if date == "" {
		return fmt.Errorf("no server date in headers")
	}

	// 调整时钟偏差
	return adjustClockSkew(date)
}
