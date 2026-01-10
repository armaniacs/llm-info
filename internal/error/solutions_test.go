package error

import (
	"runtime"
	"testing"
)

func TestNewSolutionProvider(t *testing.T) {
	provider := NewSolutionProvider()

	if provider.osInfo == "" {
		t.Error("NewSolutionProvider() should set osInfo")
	}

	expectedOSInfo := runtime.GOOS + "/" + runtime.GOARCH
	if provider.osInfo != expectedOSInfo {
		t.Errorf("NewSolutionProvider() osInfo = %v, want %v", provider.osInfo, expectedOSInfo)
	}
}

func TestSolutionProvider_GetGeneralSolutions(t *testing.T) {
	provider := NewSolutionProvider()
	solutions := provider.GetGeneralSolutions()

	if len(solutions) < 3 {
		t.Errorf("GetGeneralSolutions() should return at least 3 solutions, got %d", len(solutions))
	}

	// 基本的な解決策が含まれているか確認
	expectedSolutions := []string{
		"最新バージョンにアップデートしてください: llm-info --version",
		"詳細なログを確認してください: llm-info --verbose",
		"設定ファイルを確認してください: llm-info --check-config",
	}

	for _, expected := range expectedSolutions {
		found := false
		for _, solution := range solutions {
			if solution == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetGeneralSolutions() should include %v", expected)
		}
	}

	// OS固有の解決策が含まれているか確認
	if runtime.GOOS == "windows" {
		found := false
		for _, solution := range solutions {
			if solution == "Windowsファイアウォール設定を確認してください" {
				found = true
				break
			}
		}
		if !found {
			t.Error("GetGeneralSolutions() should include Windows-specific solution on Windows")
		}
	} else {
		found := false
		for _, solution := range solutions {
			if solution == "iptables設定を確認してください" {
				found = true
				break
			}
		}
		if !found {
			t.Error("GetGeneralSolutions() should include iptables solution on non-Windows")
		}
	}
}

func TestSolutionProvider_GetNetworkSolutions(t *testing.T) {
	provider := NewSolutionProvider()

	tests := []struct {
		name     string
		url      string
		expected []string
	}{
		{
			name: "HTTP URL",
			url:  "http://api.example.com",
			expected: []string{
				"ホスト api.example.com に到達できるか確認してください",
				"プロキシ設定を確認してください",
				"DNS設定を確認してください",
			},
		},
		{
			name: "HTTPS URL",
			url:  "https://api.example.com",
			expected: []string{
				"ホスト api.example.com に到達できるか確認してください",
				"プロキシ設定を確認してください",
				"DNS設定を確認してください",
				"TLS/SSL設定を確認してください",
			},
		},
		{
			name: "Invalid URL",
			url:  "invalid-url",
			expected: []string{
				"ホスト  に到達できるか確認してください",
				"プロキシ設定を確認してください",
				"DNS設定を確認してください",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solutions := provider.GetNetworkSolutions(tt.url)

			if len(solutions) < len(tt.expected) {
				t.Errorf("GetNetworkSolutions() should return at least %d solutions, got %d", len(tt.expected), len(solutions))
			}

			for _, expected := range tt.expected {
				found := false
				for _, solution := range solutions {
					if solution == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetNetworkSolutions() should include %v", expected)
				}
			}
		})
	}
}

func TestSolutionProvider_GetConfigSolutions(t *testing.T) {
	provider := NewSolutionProvider()
	configPath := "/path/to/config.yaml"

	solutions := provider.GetConfigSolutions(configPath)

	expectedSolutions := []string{
		"設定ファイル /path/to/config.yaml が存在するか確認してください",
		"設定ファイルのパーミッションを確認してください",
		"設定ファイルのYAML構文を確認してください",
		"llm-info --init-config で初期設定を作成してください",
	}

	if len(solutions) != len(expectedSolutions) {
		t.Errorf("GetConfigSolutions() should return %d solutions, got %d", len(expectedSolutions), len(solutions))
	}

	for _, expected := range expectedSolutions {
		found := false
		for _, solution := range solutions {
			if solution == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetConfigSolutions() should include %v", expected)
		}
	}
}

func TestSolutionProvider_GetAPISolutions(t *testing.T) {
	provider := NewSolutionProvider()
	url := "https://api.example.com"

	tests := []struct {
		name       string
		statusCode int
		expected   []string
	}{
		{
			name:       "401 Unauthorized",
			statusCode: 401,
			expected: []string{
				"APIキーが正しいか確認してください",
				"APIキーの有効期限が切れていないか確認してください",
				"APIキーの権限設定を確認してください",
			},
		},
		{
			name:       "403 Forbidden",
			statusCode: 403,
			expected: []string{
				"APIキーに必要な権限があるか確認してください",
				"アカウントの利用制限を確認してください",
			},
		},
		{
			name:       "404 Not Found",
			statusCode: 404,
			expected: []string{
				"ゲートウェイURLが正しいか確認してください",
				"APIバージョンが正しいか確認してください",
				"エンドポイントパスを確認してください",
			},
		},
		{
			name:       "429 Rate Limit",
			statusCode: 429,
			expected: []string{
				"しばらく待ってから再試行してください",
				"APIプランのレート制限を確認してください",
				"並列リクエスト数を減らしてください",
			},
		},
		{
			name:       "500 Server Error",
			statusCode: 500,
			expected: []string{
				"サーバーが一時的に利用できない可能性があります",
				"しばらく待ってから再試行してください",
				"サービスステータスページを確認してください",
			},
		},
		{
			name:       "Unknown Status",
			statusCode: 0,
			expected: []string{
				"APIドキュメントを確認してください",
				"サポートにお問い合わせください",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solutions := provider.GetAPISolutions(tt.statusCode, url)

			if len(solutions) < len(tt.expected) {
				t.Errorf("GetAPISolutions() should return at least %d solutions, got %d", len(tt.expected), len(solutions))
			}

			for _, expected := range tt.expected {
				found := false
				for _, solution := range solutions {
					if solution == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetAPISolutions() should include %v", expected)
				}
			}
		})
	}
}

func TestSolutionProvider_GetUserSolutions(t *testing.T) {
	provider := NewSolutionProvider()

	tests := []struct {
		name     string
		code     string
		argument string
		expected []string
	}{
		{
			name:     "Invalid argument",
			code:     "invalid_argument",
			argument: "test-arg",
			expected: []string{
				"コマンドライン引数が正しいか確認してください",
				"ヘルプを確認してください: llm-info --help",
			},
		},
		{
			name:     "Invalid filter syntax",
			code:     "invalid_filter_syntax",
			argument: "invalid-filter",
			expected: []string{
				"フィルタ構文を確認してください",
				"例: --filter \"name:gpt,tokens>1000\"",
				"ヘルプを確認してください: llm-info --help filter",
			},
		},
		{
			name:     "Invalid sort field",
			code:     "invalid_sort_field",
			argument: "invalid-field",
			expected: []string{
				"ソートフィールドを確認してください",
				"使用可能なフィールド: name, tokens, cost, mode",
				"ヘルプを確認してください: llm-info --help sort",
			},
		},
		{
			name:     "Gateway not found",
			code:     "gateway_not_found",
			argument: "test-gateway",
			expected: []string{
				"ゲートウェイ名が正しいか確認してください",
				"利用可能なゲートウェイを確認してください: llm-info --list-gateways",
				"設定ファイルにゲートウェイが登録されているか確認してください",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solutions := provider.GetUserSolutions(tt.code, tt.argument)

			if len(solutions) != len(tt.expected) {
				t.Errorf("GetUserSolutions() should return %d solutions, got %d", len(tt.expected), len(solutions))
			}

			for _, expected := range tt.expected {
				found := false
				for _, solution := range solutions {
					if solution == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetUserSolutions() should include %v", expected)
				}
			}
		})
	}
}

func TestSolutionProvider_GetSystemSolutions(t *testing.T) {
	provider := NewSolutionProvider()

	tests := []struct {
		name     string
		code     string
		context  string
		expected []string
	}{
		{
			name:    "Permission denied",
			code:    "permission_denied",
			context: "file access",
			expected: []string{
				"ファイルのアクセス権限を確認してください",
				"ファイルの所有者を確認してください",
			},
		},
		{
			name:    "Disk full",
			code:    "disk_full",
			context: "storage",
			expected: []string{
				"ディスク容量を確認してください",
				"不要なファイルを削除してください",
				"別のストレージを使用してください",
			},
		},
		{
			name:    "Memory insufficient",
			code:    "memory_insufficient",
			context: "memory",
			expected: []string{
				"メモリ使用量を確認してください",
				"他のアプリケーションを終了してください",
				"システムを再起動してください",
			},
		},
		{
			name:    "Unexpected error",
			code:    "unexpected_error",
			context: "system",
			expected: []string{
				"開発者にバグレポートを送信してください",
				"詳細なログを確認してください: llm-info --verbose",
				"最新バージョンにアップデートしてください",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solutions := provider.GetSystemSolutions(tt.code, tt.context)

			// permission_deniedの場合はOS固有の解決策が追加されるため、最小数をチェック
			minExpected := len(tt.expected)
			if tt.code == "permission_denied" {
				minExpected = 3 // 基本解決策2つ + OS固有の解決策1つ
			}

			if len(solutions) < minExpected {
				t.Errorf("GetSystemSolutions() should return at least %d solutions, got %d", minExpected, len(solutions))
			}

			// Windowsの場合は管理者権限のメッセージが異なる
			if tt.code == "permission_denied" {
				if runtime.GOOS == "windows" {
					found := false
					for _, solution := range solutions {
						if solution == "管理者として実行してください" {
							found = true
							break
						}
					}
					if !found {
						t.Error("GetSystemSolutions() should include Windows-specific admin solution on Windows")
					}
				} else {
					found := false
					for _, solution := range solutions {
						if solution == "sudoを使用して実行してください" {
							found = true
							break
						}
					}
					if !found {
						t.Error("GetSystemSolutions() should include sudo solution on non-Windows")
					}
				}
			}

			for _, expected := range tt.expected {
				found := false
				for _, solution := range solutions {
					if solution == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetSystemSolutions() should include %v", expected)
				}
			}
		})
	}
}

func TestSolutionProvider_GetContextualSolutions(t *testing.T) {
	provider := NewSolutionProvider()

	tests := []struct {
		name         string
		appErr       *AppError
		expectedType ErrorType
	}{
		{
			name: "Network error with URL",
			appErr: &AppError{
				Type: ErrorTypeNetwork,
				Context: map[string]interface{}{
					"url": "https://api.example.com",
				},
			},
			expectedType: ErrorTypeNetwork,
		},
		{
			name: "API error with status code",
			appErr: &AppError{
				Type: ErrorTypeAPI,
				Context: map[string]interface{}{
					"url":         "https://api.example.com",
					"status_code": 401,
				},
			},
			expectedType: ErrorTypeAPI,
		},
		{
			name: "Config error with path",
			appErr: &AppError{
				Type: ErrorTypeConfig,
				Context: map[string]interface{}{
					"config_path": "/path/to/config.yaml",
				},
			},
			expectedType: ErrorTypeConfig,
		},
		{
			name: "User error with argument",
			appErr: &AppError{
				Type: ErrorTypeUser,
				Code: "invalid_argument",
				Context: map[string]interface{}{
					"argument": "test-arg",
				},
			},
			expectedType: ErrorTypeUser,
		},
		{
			name: "System error with context",
			appErr: &AppError{
				Type: ErrorTypeSystem,
				Code: "permission_denied",
				Context: map[string]interface{}{
					"context": "file access",
				},
			},
			expectedType: ErrorTypeSystem,
		},
		{
			name: "Unknown error",
			appErr: &AppError{
				Type: ErrorTypeUnknown,
			},
			expectedType: ErrorTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solutions := provider.GetContextualSolutions(tt.appErr)

			if len(solutions) == 0 {
				t.Error("GetContextualSolutions() should return at least one solution")
			}
		})
	}
}

func TestSolutionProvider_GetHelpURL(t *testing.T) {
	provider := NewSolutionProvider()

	tests := []struct {
		name      string
		errorType ErrorType
		expected  string
	}{
		{
			name:      "Network error",
			errorType: ErrorTypeNetwork,
			expected:  "https://github.com/armaniacs/llm-info/wiki/network-errors",
		},
		{
			name:      "API error",
			errorType: ErrorTypeAPI,
			expected:  "https://github.com/armaniacs/llm-info/wiki/api-errors",
		},
		{
			name:      "Config error",
			errorType: ErrorTypeConfig,
			expected:  "https://github.com/armaniacs/llm-info/wiki/config-errors",
		},
		{
			name:      "User error",
			errorType: ErrorTypeUser,
			expected:  "https://github.com/armaniacs/llm-info/wiki/usage",
		},
		{
			name:      "System error",
			errorType: ErrorTypeSystem,
			expected:  "https://github.com/armaniacs/llm-info/wiki/troubleshooting",
		},
		{
			name:      "Unknown error",
			errorType: ErrorTypeUnknown,
			expected:  "https://github.com/armaniacs/llm-info/wiki/errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpURL := provider.GetHelpURL(tt.errorType)

			if helpURL != tt.expected {
				t.Errorf("GetHelpURL() = %v, want %v", helpURL, tt.expected)
			}
		})
	}
}

func TestSolutionProvider_EnhanceError(t *testing.T) {
	provider := NewSolutionProvider()

	tests := []struct {
		name         string
		appErr       *AppError
		hasSolutions bool
		hasHelpURL   bool
	}{
		{
			name: "Error without solutions",
			appErr: &AppError{
				Type:    ErrorTypeNetwork,
				Code:    "connection_timeout",
				Message: "Connection timeout",
			},
			hasSolutions: true,
			hasHelpURL:   true,
		},
		{
			name: "Error with existing solutions",
			appErr: &AppError{
				Type:      ErrorTypeNetwork,
				Code:      "connection_timeout",
				Message:   "Connection timeout",
				Solutions: []string{"Existing solution"},
			},
			hasSolutions: true,
			hasHelpURL:   true,
		},
		{
			name: "Error with existing help URL",
			appErr: &AppError{
				Type:    ErrorTypeNetwork,
				Code:    "connection_timeout",
				Message: "Connection timeout",
				HelpURL: "https://example.com/help",
			},
			hasSolutions: true,
			hasHelpURL:   true,
		},
		{
			name:         "Nil error",
			appErr:       nil,
			hasSolutions: false,
			hasHelpURL:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.EnhanceError(tt.appErr)

			if tt.appErr == nil {
				if result != nil {
					t.Error("EnhanceError() should return nil for nil error")
				}
				return
			}

			if result != tt.appErr {
				t.Error("EnhanceError() should return the same error instance")
			}

			if tt.hasSolutions && len(result.Solutions) == 0 {
				t.Error("EnhanceError() should add solutions when none exist")
			}

			if tt.hasHelpURL && result.HelpURL == "" {
				t.Error("EnhanceError() should add help URL when none exists")
			}
		})
	}
}

func TestGetOSInfo(t *testing.T) {
	osInfo := getOSInfo()
	expected := runtime.GOOS + "/" + runtime.GOARCH

	if osInfo != expected {
		t.Errorf("getOSInfo() = %v, want %v", osInfo, expected)
	}
}
