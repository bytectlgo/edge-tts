package edge_tts

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// ListVoices 获取所有可用的语音列表
func ListVoices(ctx context.Context) ([]Voice, error) {
	// 获取系统代理
	proxy := os.Getenv("HTTP_PROXY")
	if proxy == "" {
		proxy = os.Getenv("HTTPS_PROXY")
	}

	// 创建HTTP客户端
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// 如果有代理，设置代理
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("parse proxy URL failed: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Transport: transport,
	}

	// 生成安全令牌
	secMsGec := generateSecMsGec()

	// 构建请求URL
	reqURL := fmt.Sprintf("%s&Sec-MS-GEC=%s&Sec-MS-GEC-Version=%s",
		VoiceList, secMsGec, SEC_MS_GEC_VERSION)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	for k, v := range BaseHeaders {
		req.Header.Set(k, v)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		// 如果是403错误，可能需要调整时钟偏差
		if resp.StatusCode == http.StatusForbidden {
			if err := handleClientResponseError(resp); err != nil {
				return nil, err
			}
			// 重试请求
			return ListVoices(ctx)
		}
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	// 解析响应
	var voices []Voice
	if err := json.Unmarshal(body, &voices); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	// 清理语音标签中的空白字符
	for i := range voices {
		for j := range voices[i].VoiceTag.ContentCategories {
			voices[i].VoiceTag.ContentCategories[j] = strings.TrimSpace(voices[i].VoiceTag.ContentCategories[j])
		}
		for j := range voices[i].VoiceTag.VoicePersonalities {
			voices[i].VoiceTag.VoicePersonalities[j] = strings.TrimSpace(voices[i].VoiceTag.VoicePersonalities[j])
		}
	}

	return voices, nil
}

// ListVoicesWithProxy 使用代理获取所有可用的语音列表
func ListVoicesWithProxy(ctx context.Context, proxyURL string) ([]Voice, error) {
	// 解析代理URL
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("parse proxy URL failed: %w", err)
	}

	// 创建HTTP客户端
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Proxy: http.ProxyURL(proxy),
		},
	}

	// 生成安全令牌
	secMsGec := generateSecMsGec()

	// 构建请求URL
	reqURL := fmt.Sprintf("%s&Sec-MS-GEC=%s&Sec-MS-GEC-Version=%s",
		VoiceList, secMsGec, SEC_MS_GEC_VERSION)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	for k, v := range BaseHeaders {
		req.Header.Set(k, v)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		// 如果是403错误，可能需要调整时钟偏差
		if resp.StatusCode == http.StatusForbidden {
			if err := handleClientResponseError(resp); err != nil {
				return nil, err
			}
			// 重试请求
			return ListVoicesWithProxy(ctx, proxyURL)
		}
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	// 解析响应
	var voices []Voice
	if err := json.Unmarshal(body, &voices); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	// 清理语音标签中的空白字符
	for i := range voices {
		for j := range voices[i].VoiceTag.ContentCategories {
			voices[i].VoiceTag.ContentCategories[j] = strings.TrimSpace(voices[i].VoiceTag.ContentCategories[j])
		}
		for j := range voices[i].VoiceTag.VoicePersonalities {
			voices[i].VoiceTag.VoicePersonalities[j] = strings.TrimSpace(voices[i].VoiceTag.VoicePersonalities[j])
		}
	}

	return voices, nil
}
