package error

import (
	"errors"
	"testing"
)

func TestGetErrorMessage(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		code      string
		expected  string
	}{
		{
			name:      "Network timeout error",
			errorType: ErrorTypeNetwork,
			code:      "connection_timeout",
			expected:  "接続がタイムアウトしました",
		},
		{
			name:      "API authentication error",
			errorType: ErrorTypeAPI,
			code:      "authentication_failed",
			expected:  "認証に失敗しました",
		},
		{
			name:      "Config file not found",
			errorType: ErrorTypeConfig,
			code:      "config_file_not_found",
			expected:  "設定ファイルが見つかりません",
		},
		{
			name:      "Invalid argument",
			errorType: ErrorTypeUser,
			code:      "invalid_argument",
			expected:  "無効な引数です",
		},
		{
			name:      "Permission denied",
			errorType: ErrorTypeSystem,
			code:      "permission_denied",
			expected:  "ファイルアクセス権限がありません",
		},
		{
			name:      "Unknown error type",
			errorType: ErrorTypeUnknown,
			code:      "unknown_code",
			expected:  "不明なエラーが発生しました",
		},
		{
			name:      "Unknown error code",
			errorType: ErrorTypeNetwork,
			code:      "unknown_code",
			expected:  "不明なエラーが発生しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetErrorMessage(tt.errorType, tt.code); got != tt.expected {
				t.Errorf("GetErrorMessage() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCreateNetworkError(t *testing.T) {
	causeErr := errors.New("connection failed")
	url := "https://api.example.com"

	tests := []struct {
		name              string
		code              string
		expectedCode      string
		expectedSolutions int
	}{
		{
			name:              "Connection timeout",
			code:              "connection_timeout",
			expectedCode:      "connection_timeout",
			expectedSolutions: 3,
		},
		{
			name:              "DNS resolution failed",
			code:              "dns_resolution_failed",
			expectedCode:      "dns_resolution_failed",
			expectedSolutions: 3,
		},
		{
			name:              "TLS certificate error",
			code:              "tls_certificate_error",
			expectedCode:      "tls_certificate_error",
			expectedSolutions: 2,
		},
		{
			name:              "Connection refused",
			code:              "connection_refused",
			expectedCode:      "connection_refused",
			expectedSolutions: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := CreateNetworkError(tt.code, url, causeErr)

			if appErr.Type != ErrorTypeNetwork {
				t.Errorf("CreateNetworkError() Type = %v, want %v", appErr.Type, ErrorTypeNetwork)
			}
			if appErr.Code != tt.expectedCode {
				t.Errorf("CreateNetworkError() Code = %v, want %v", appErr.Code, tt.expectedCode)
			}
			if appErr.OriginalErr != causeErr {
				t.Errorf("CreateNetworkError() OriginalErr = %v, want %v", appErr.OriginalErr, causeErr)
			}
			if len(appErr.Solutions) != tt.expectedSolutions {
				t.Errorf("CreateNetworkError() Solutions length = %v, want %v", len(appErr.Solutions), tt.expectedSolutions)
			}
			if appErr.Context["url"] != url {
				t.Errorf("CreateNetworkError() Context[url] = %v, want %v", appErr.Context["url"], url)
			}
			if appErr.HelpURL != "https://github.com/armaniacs/llm-info/wiki/network-errors" {
				t.Errorf("CreateNetworkError() HelpURL = %v, want %v", appErr.HelpURL, "https://github.com/armaniacs/llm-info/wiki/network-errors")
			}
		})
	}
}

func TestCreateAPIError(t *testing.T) {
	causeErr := errors.New("API failed")
	url := "https://api.example.com"
	statusCode := 401

	tests := []struct {
		name              string
		code              string
		expectedSolutions int
	}{
		{
			name:              "Authentication failed",
			code:              "authentication_failed",
			expectedSolutions: 3,
		},
		{
			name:              "Rate limit exceeded",
			code:              "rate_limit_exceeded",
			expectedSolutions: 3,
		},
		{
			name:              "Endpoint not found",
			code:              "endpoint_not_found",
			expectedSolutions: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := CreateAPIError(tt.code, statusCode, url, causeErr)

			if appErr.Type != ErrorTypeAPI {
				t.Errorf("CreateAPIError() Type = %v, want %v", appErr.Type, ErrorTypeAPI)
			}
			if appErr.Code != tt.code {
				t.Errorf("CreateAPIError() Code = %v, want %v", appErr.Code, tt.code)
			}
			if appErr.OriginalErr != causeErr {
				t.Errorf("CreateAPIError() OriginalErr = %v, want %v", appErr.OriginalErr, causeErr)
			}
			if len(appErr.Solutions) != tt.expectedSolutions {
				t.Errorf("CreateAPIError() Solutions length = %v, want %v", len(appErr.Solutions), tt.expectedSolutions)
			}
			if appErr.Context["url"] != url {
				t.Errorf("CreateAPIError() Context[url] = %v, want %v", appErr.Context["url"], url)
			}
			if appErr.Context["status_code"] != statusCode {
				t.Errorf("CreateAPIError() Context[status_code] = %v, want %v", appErr.Context["status_code"], statusCode)
			}
			if appErr.HelpURL != "https://github.com/armaniacs/llm-info/wiki/api-errors" {
				t.Errorf("CreateAPIError() HelpURL = %v, want %v", appErr.HelpURL, "https://github.com/armaniacs/llm-info/wiki/api-errors")
			}
		})
	}
}

func TestCreateConfigError(t *testing.T) {
	causeErr := errors.New("config failed")
	configPath := "/path/to/config.yaml"

	tests := []struct {
		name              string
		code              string
		expectedSolutions int
	}{
		{
			name:              "Config file not found",
			code:              "config_file_not_found",
			expectedSolutions: 3,
		},
		{
			name:              "Invalid config format",
			code:              "invalid_config_format",
			expectedSolutions: 3,
		},
		{
			name:              "Missing required field",
			code:              "missing_required_field",
			expectedSolutions: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := CreateConfigError(tt.code, configPath, causeErr)

			if appErr.Type != ErrorTypeConfig {
				t.Errorf("CreateConfigError() Type = %v, want %v", appErr.Type, ErrorTypeConfig)
			}
			if appErr.Code != tt.code {
				t.Errorf("CreateConfigError() Code = %v, want %v", appErr.Code, tt.code)
			}
			if appErr.OriginalErr != causeErr {
				t.Errorf("CreateConfigError() OriginalErr = %v, want %v", appErr.OriginalErr, causeErr)
			}
			if len(appErr.Solutions) != tt.expectedSolutions {
				t.Errorf("CreateConfigError() Solutions length = %v, want %v", len(appErr.Solutions), tt.expectedSolutions)
			}
			if appErr.Context["config_path"] != configPath {
				t.Errorf("CreateConfigError() Context[config_path] = %v, want %v", appErr.Context["config_path"], configPath)
			}
			if appErr.HelpURL != "https://github.com/armaniacs/llm-info/wiki/config-errors" {
				t.Errorf("CreateConfigError() HelpURL = %v, want %v", appErr.HelpURL, "https://github.com/armaniacs/llm-info/wiki/config-errors")
			}
		})
	}
}

func TestCreateUserError(t *testing.T) {
	causeErr := errors.New("user error")
	argument := "invalid-arg"

	tests := []struct {
		name              string
		code              string
		expectedSolutions int
	}{
		{
			name:              "Invalid argument",
			code:              "invalid_argument",
			expectedSolutions: 2,
		},
		{
			name:              "Invalid filter syntax",
			code:              "invalid_filter_syntax",
			expectedSolutions: 3,
		},
		{
			name:              "Invalid sort field",
			code:              "invalid_sort_field",
			expectedSolutions: 3,
		},
		{
			name:              "Gateway not found",
			code:              "gateway_not_found",
			expectedSolutions: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := CreateUserError(tt.code, argument, causeErr)

			if appErr.Type != ErrorTypeUser {
				t.Errorf("CreateUserError() Type = %v, want %v", appErr.Type, ErrorTypeUser)
			}
			if appErr.Severity != SeverityWarning {
				t.Errorf("CreateUserError() Severity = %v, want %v", appErr.Severity, SeverityWarning)
			}
			if appErr.Code != tt.code {
				t.Errorf("CreateUserError() Code = %v, want %v", appErr.Code, tt.code)
			}
			if appErr.OriginalErr != causeErr {
				t.Errorf("CreateUserError() OriginalErr = %v, want %v", appErr.OriginalErr, causeErr)
			}
			if len(appErr.Solutions) != tt.expectedSolutions {
				t.Errorf("CreateUserError() Solutions length = %v, want %v", len(appErr.Solutions), tt.expectedSolutions)
			}
			if appErr.Context["argument"] != argument {
				t.Errorf("CreateUserError() Context[argument] = %v, want %v", appErr.Context["argument"], argument)
			}
			if appErr.HelpURL != "https://github.com/armaniacs/llm-info/wiki/usage" {
				t.Errorf("CreateUserError() HelpURL = %v, want %v", appErr.HelpURL, "https://github.com/armaniacs/llm-info/wiki/usage")
			}
		})
	}
}

func TestCreateSystemError(t *testing.T) {
	causeErr := errors.New("system error")
	context := "system context"

	tests := []struct {
		name              string
		code              string
		expectedSolutions int
	}{
		{
			name:              "Permission denied",
			code:              "permission_denied",
			expectedSolutions: 3,
		},
		{
			name:              "Disk full",
			code:              "disk_full",
			expectedSolutions: 3,
		},
		{
			name:              "Memory insufficient",
			code:              "memory_insufficient",
			expectedSolutions: 3,
		},
		{
			name:              "Unexpected error",
			code:              "unexpected_error",
			expectedSolutions: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := CreateSystemError(tt.code, context, causeErr)

			if appErr.Type != ErrorTypeSystem {
				t.Errorf("CreateSystemError() Type = %v, want %v", appErr.Type, ErrorTypeSystem)
			}
			if appErr.Code != tt.code {
				t.Errorf("CreateSystemError() Code = %v, want %v", appErr.Code, tt.code)
			}
			if appErr.OriginalErr != causeErr {
				t.Errorf("CreateSystemError() OriginalErr = %v, want %v", appErr.OriginalErr, causeErr)
			}
			if len(appErr.Solutions) != tt.expectedSolutions {
				t.Errorf("CreateSystemError() Solutions length = %v, want %v", len(appErr.Solutions), tt.expectedSolutions)
			}
			if appErr.Context["context"] != context {
				t.Errorf("CreateSystemError() Context[context] = %v, want %v", appErr.Context["context"], context)
			}
			if appErr.HelpURL != "https://github.com/armaniacs/llm-info/issues" {
				t.Errorf("CreateSystemError() HelpURL = %v, want %v", appErr.HelpURL, "https://github.com/armaniacs/llm-info/issues")
			}
		})
	}
}

func TestDetectErrorType(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedType ErrorType
		expectedCode string
	}{
		{
			name:         "Timeout error",
			err:          errors.New("request timeout"),
			expectedType: ErrorTypeNetwork,
			expectedCode: "connection_timeout",
		},
		{
			name:         "Connection refused",
			err:          errors.New("connection refused"),
			expectedType: ErrorTypeNetwork,
			expectedCode: "connection_refused",
		},
		{
			name:         "DNS error",
			err:          errors.New("no such host"),
			expectedType: ErrorTypeNetwork,
			expectedCode: "dns_resolution_failed",
		},
		{
			name:         "401 error",
			err:          errors.New("401 unauthorized"),
			expectedType: ErrorTypeAPI,
			expectedCode: "authentication_failed",
		},
		{
			name:         "403 error",
			err:          errors.New("403 forbidden"),
			expectedType: ErrorTypeAPI,
			expectedCode: "authorization_failed",
		},
		{
			name:         "404 error",
			err:          errors.New("404 not found"),
			expectedType: ErrorTypeAPI,
			expectedCode: "endpoint_not_found",
		},
		{
			name:         "429 error",
			err:          errors.New("429 rate limit"),
			expectedType: ErrorTypeAPI,
			expectedCode: "rate_limit_exceeded",
		},
		{
			name:         "500 error",
			err:          errors.New("500 server error"),
			expectedType: ErrorTypeAPI,
			expectedCode: "server_error",
		},
		{
			name:         "File not found",
			err:          errors.New("no such file or directory"),
			expectedType: ErrorTypeConfig,
			expectedCode: "config_file_not_found",
		},
		{
			name:         "Permission denied",
			err:          errors.New("permission denied"),
			expectedType: ErrorTypeSystem,
			expectedCode: "permission_denied",
		},
		{
			name:         "Unknown error",
			err:          errors.New("unknown error"),
			expectedType: ErrorTypeUnknown,
			expectedCode: "unexpected_error",
		},
		{
			name:         "Nil error",
			err:          nil,
			expectedType: ErrorTypeUnknown,
			expectedCode: "unknown_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorType, code := DetectErrorType(tt.err)
			if errorType != tt.expectedType {
				t.Errorf("DetectErrorType() Type = %v, want %v", errorType, tt.expectedType)
			}
			if code != tt.expectedCode {
				t.Errorf("DetectErrorType() Code = %v, want %v", code, tt.expectedCode)
			}
		})
	}
}

func TestWrapErrorWithDetection(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		context      string
		expectedType ErrorType
	}{
		{
			name:         "Network error",
			err:          errors.New("connection timeout"),
			context:      "https://api.example.com",
			expectedType: ErrorTypeNetwork,
		},
		{
			name:         "API error",
			err:          errors.New("401 unauthorized"),
			context:      "https://api.example.com",
			expectedType: ErrorTypeAPI,
		},
		{
			name:         "Config error",
			err:          errors.New("no such file"),
			context:      "/path/to/config.yaml",
			expectedType: ErrorTypeConfig,
		},
		{
			name:         "System error",
			err:          errors.New("permission denied"),
			context:      "file access",
			expectedType: ErrorTypeSystem,
		},
		{
			name:         "Nil error",
			err:          nil,
			context:      "test",
			expectedType: ErrorTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := WrapErrorWithDetection(tt.err, tt.context)

			if tt.err == nil {
				if appErr != nil {
					t.Errorf("WrapErrorWithDetection() should return nil for nil error")
				}
				return
			}

			if appErr.Type != tt.expectedType {
				t.Errorf("WrapErrorWithDetection() Type = %v, want %v", appErr.Type, tt.expectedType)
			}
			if appErr.OriginalErr != tt.err {
				t.Errorf("WrapErrorWithDetection() OriginalErr = %v, want %v", appErr.OriginalErr, tt.err)
			}
		})
	}
}
