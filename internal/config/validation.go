package config

import (
	"net/url"
	"strings"
)

// isValidURL はURL形式をより厳密に検証
func isValidURL(urlStr string) bool {
	// 明らかに無効なパターンをチェック
	invalidPatterns := []string{
		" ",  // スペースを含む
		"\t", // タブを含む
		"\n", // 改行を含む
	}

	for _, pattern := range invalidPatterns {
		if strings.Contains(urlStr, pattern) {
			return false
		}
	}

	// スキームなしで始まる場合をチェック（//example.com のような形式）
	if strings.HasPrefix(urlStr, "//") {
		return false
	}

	// url.Parseで検証
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// HTTP/HTTPSのみ許可
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	// ホスト名の存在チェック
	if parsed.Host == "" {
		return false
	}

	return true
}

// contains はスライスに文字列が含まれているかチェックする
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}