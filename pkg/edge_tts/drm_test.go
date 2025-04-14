package edge_tts

import (
	"testing"
	"time"
)

func TestAdjustClockSkew(t *testing.T) {
	// 测试调整时钟偏差
	date := time.Now().UTC().Format(time.RFC1123)
	err := adjustClockSkew(date)
	if err != nil {
		t.Errorf("adjustClockSkew failed: %v", err)
	}

	// 测试正偏差
	clockSkewSeconds = 0
	date = time.Now().Add(10 * time.Second).UTC().Format(time.RFC1123)
	err = adjustClockSkew(date)
	if err != nil {
		t.Errorf("adjustClockSkew failed: %v", err)
	}
	if clockSkewSeconds != 10 {
		t.Errorf("expected clockSkewSeconds to be 10, got %f", clockSkewSeconds)
	}

	// 测试负偏差
	clockSkewSeconds = 0
	date = time.Now().Add(-5 * time.Second).UTC().Format(time.RFC1123)
	err = adjustClockSkew(date)
	if err != nil {
		t.Errorf("adjustClockSkew failed: %v", err)
	}
	if clockSkewSeconds != -5 {
		t.Errorf("expected clockSkewSeconds to be -5, got %f", clockSkewSeconds)
	}
}

func TestGetUnixTimestamp(t *testing.T) {
	// 测试获取时间戳
	now := time.Now().Unix()
	timestamp := getUnixTimestamp()

	// 检查时间戳是否在合理范围内（当前时间前后5秒）
	if timestamp < now-5 || timestamp > now+5 {
		t.Errorf("timestamp %d is not within expected range (%d to %d)", timestamp, now-5, now+5)
	}
}

func TestGenerateSecMsGec(t *testing.T) {
	// 测试生成 Sec-MS-GEC token
	token := generateSecMsGec()

	// 检查 token 格式
	if len(token) != 64 {
		t.Errorf("token length should be 64, got %d", len(token))
	}

	// 检查 token 是否只包含十六进制字符
	for _, c := range token {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
			t.Errorf("token contains invalid character: %c", c)
		}
	}
}

func TestParseRFC2616Date(t *testing.T) {
	// 测试有效的 RFC 2616 日期
	date := "Mon, 15 Jan 2023 12:34:56 GMT"
	timestamp, err := parseRFC2616Date(date)
	if err != nil {
		t.Errorf("failed to parse valid date: %v", err)
	}

	// 检查时间戳是否在预期范围内
	expected := time.Date(2023, time.January, 15, 12, 34, 56, 0, time.UTC).Unix()
	if timestamp != expected {
		t.Errorf("expected timestamp %d, got %d", expected, timestamp)
	}

	// 测试无效的日期格式
	invalidDate := "Invalid Date"
	_, err = parseRFC2616Date(invalidDate)
	if err == nil {
		t.Error("expected error for invalid date format")
	}
}
