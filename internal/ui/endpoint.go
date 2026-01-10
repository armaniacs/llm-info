package ui

import (
	"net/url"
	"strings"
)

// MaskURL 敏捷性情報をマスキングしてURLを表示する
func MaskURL(urlStr string) string {
	if urlStr == "" {
		return ""
	}

	// URLを解析
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		// 解析できない場合はそのまま返す
		return urlStr
	}

	// クエリパラメータを処理
	if parsedURL.RawQuery != "" {
		queryParams := parsedURL.Query()
		maskedParams := url.Values{}

		// 敏捷性の高いパラメータをマスキング
		sensitiveParams := map[string]bool{
			"api_key":    true,
			"apikey":     true,
			"token":      true,
			"access_token": true,
			"auth":       true,
			"key":        true,
			"secret":     true,
		}

		for key, values := range queryParams {
			lowerKey := strings.ToLower(key)
			if sensitiveParams[lowerKey] {
				// 敏捷性情報をマスキング
				maskedParams[key] = []string{"****"}
			} else {
				maskedParams[key] = values
			}
		}

		parsedURL.RawQuery = maskedParams.Encode()
	}

	return parsedURL.String()
}

// DisplayEndpoint はエンドポイントURLを表示する
func DisplayEndpoint(urlStr string) error {
	if urlStr == "" {
		return nil
	}

	maskedURL := MaskURL(urlStr)

	// エンドポイント情報を表示
	printEndpoint(maskedURL)

	return nil
}

// printEndpoint はエンドポイント情報を実際に出力する
func printEndpoint(urlStr string) {
	println()
	println("Endpoint:", urlStr)
	println()
}