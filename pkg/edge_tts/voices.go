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

// ListVoices gets all available voices
func ListVoices(ctx context.Context) ([]Voice, error) {
	// Get system proxy
	proxy := os.Getenv("HTTP_PROXY")
	if proxy == "" {
		proxy = os.Getenv("HTTPS_PROXY")
	}

	// Create HTTP client
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// If proxy exists, set it
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

	// Generate security token
	secMsGec := generateSecMsGec()

	// Build request URL
	reqURL := fmt.Sprintf("%s&Sec-MS-GEC=%s&Sec-MS-GEC-Version=%s",
		VoiceList, secMsGec, SEC_MS_GEC_VERSION)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// Set request headers
	for k, v := range BaseHeaders {
		req.Header.Set(k, v)
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		// If 403 error, may need to adjust clock skew
		if resp.StatusCode == http.StatusForbidden {
			if err := handleClientResponseError(resp); err != nil {
				return nil, err
			}
			// Retry request
			return ListVoices(ctx)
		}
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	// Read response content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	// Parse response
	var voices []Voice
	if err := json.Unmarshal(body, &voices); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	// Clean whitespace in voice tags
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

// ListVoicesWithProxy gets all available voices using a proxy
func ListVoicesWithProxy(ctx context.Context, proxyURL string) ([]Voice, error) {
	// Parse proxy URL
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("parse proxy URL failed: %w", err)
	}

	// Create HTTP client
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Proxy: http.ProxyURL(proxy),
		},
	}

	// Generate security token
	secMsGec := generateSecMsGec()

	// Build request URL
	reqURL := fmt.Sprintf("%s&Sec-MS-GEC=%s&Sec-MS-GEC-Version=%s",
		VoiceList, secMsGec, SEC_MS_GEC_VERSION)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// Set request headers
	for k, v := range BaseHeaders {
		req.Header.Set(k, v)
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		// If 403 error, may need to adjust clock skew
		if resp.StatusCode == http.StatusForbidden {
			if err := handleClientResponseError(resp); err != nil {
				return nil, err
			}
			// Retry request
			return ListVoicesWithProxy(ctx, proxyURL)
		}
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	// Read response content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	// Parse response
	var voices []Voice
	if err := json.Unmarshal(body, &voices); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	// Clean whitespace in voice tags
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
