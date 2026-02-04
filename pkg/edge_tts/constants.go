package edge_tts

import (
	"strings"
)

const (
	BaseURL              = "speech.platform.bing.com/consumer/speech/synthesize/readaloud"
	TrustedClientToken   = "6A5AA1D4EAFF4E9FB37E23D68491D6F4"
	WSSURL               = "wss://" + BaseURL + "/edge/v1?TrustedClientToken=" + TrustedClientToken
	VoiceList            = "https://" + BaseURL + "/voices/list?trustedclienttoken=" + TrustedClientToken
	DefaultVoice         = "en-US-EmmaMultilingualNeural"
	ChromiumFullVersion  = "143.0.3650.75"
	ChromiumMajorVersion = "143"
	SEC_MS_GEC_VERSION   = "1-" + ChromiumFullVersion
)

var (
	BaseHeaders = map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" +
			" (KHTML, like Gecko) Chrome/" + ChromiumMajorVersion + ".0.0.0 Safari/537.36" +
			" Edg/" + ChromiumMajorVersion + ".0.0.0",
		"Accept-Encoding": "gzip, deflate, br, zstd",
		"Accept-Language": "en-US,en;q=0.9",
		"Sec-CH-UA": `" Not;A Brand";v="99", "Microsoft Edge";v="` + ChromiumMajorVersion + `",` +
			` "Chromium";v="` + ChromiumMajorVersion + `"`,
		"Sec-CH-UA-Mobile":           "?0",
		"Sec-CH-UA-Platform":         "Windows",
		"Sec-CH-UA-Platform-Version": "10.0.0",
		"Sec-CH-UA-Arch":             "x86_64",
		"Sec-CH-UA-Bitness":          "64",
		"Sec-CH-UA-Full-Version":     ChromiumFullVersion,
		"Sec-CH-UA-Full-Version-List": `" Not;A Brand";v="99.0.0.0", "Microsoft Edge";v="` + ChromiumFullVersion + `",` +
			` "Chromium";v="` + ChromiumFullVersion + `"`,
		"Sec-CH-UA-Model": "",
	}

	// Basic WebSocket headers
	WSSHeaders = map[string]string{
		"Pragma":          "no-cache",
		"Cache-Control":   "no-cache",
		"Origin":          "chrome-extension://jdiccldimpdaibmpdkjnbmckianbfold",
		"Accept-Language": "en-US,en;q=0.9",
	}

	// WebSocket specific headers
	WSProtocolHeaders = map[string]string{
		// "Upgrade": "websocket",
		// "Connection": "Upgrade",
		// "Sec-WebSocket-Version":    "13",
		// "Sec-WebSocket-Extensions": "permessage-deflate; client_max_window_bits",
	}

	VoiceHeaders = map[string]string{
		"Authority":      "speech.platform.bing.com",
		"Accept":         "*/*",
		"Sec-Fetch-Site": "none",
		"Sec-Fetch-Mode": "cors",
		"Sec-Fetch-Dest": "empty",
	}
)

func init() {
	// Add BaseHeaders to WSSHeaders, excluding WebSocket specific headers
	for k, v := range BaseHeaders {
		// Skip WebSocket specific headers
		if k == "Upgrade" || k == "Connection" || strings.HasPrefix(k, "Sec-WebSocket") {
			continue
		}
		WSSHeaders[k] = v
	}

	// Add WebSocket specific headers
	for k, v := range WSProtocolHeaders {
		WSSHeaders[k] = v
	}

	// Add BaseHeaders to VoiceHeaders
	for k, v := range BaseHeaders {
		VoiceHeaders[k] = v
	}
}
