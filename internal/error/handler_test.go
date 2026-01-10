package error

import (
	"bytes"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := NewAppErrorCompat(NetworkError, "network error occurred", originalErr)

	expected := "NETWORK_ERROR: network error occurred (caused by: original error)"
	if appErr.Error() != expected {
		t.Errorf("AppError.Error() = %q, expected %q", appErr.Error(), expected)
	}

	// 元のエラーがない場合
	appErrNoOriginal := NewAppErrorCompat(ConfigError, "config error", nil)
	expectedNoOriginal := "CONFIG_ERROR: config error"
	if appErrNoOriginal.Error() != expectedNoOriginal {
		t.Errorf("AppError.Error() without original = %q, expected %q", appErrNoOriginal.Error(), expectedNoOriginal)
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := NewAppErrorCompat(NetworkError, "network error occurred", originalErr)

	if appErr.Unwrap() != originalErr {
		t.Errorf("AppError.Unwrap() = %v, expected %v", appErr.Unwrap(), originalErr)
	}

	// 元のエラーがない場合
	appErrNoOriginal := NewAppErrorCompat(ConfigError, "config error", nil)
	if appErrNoOriginal.Unwrap() != nil {
		t.Errorf("AppError.Unwrap() without original = %v, expected nil", appErrNoOriginal.Unwrap())
	}
}

func TestNewAppError(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := NewAppErrorCompat(NetworkError, "test error", originalErr)

	if appErr.Type != NetworkError {
		t.Errorf("NewAppError() Type = %v, expected %v", appErr.Type, NetworkError)
	}
	if appErr.Message != "test error" {
		t.Errorf("NewAppError() Message = %q, expected %q", appErr.Message, "test error")
	}
	if appErr.OriginalErr != originalErr {
		t.Errorf("NewAppError() OriginalErr = %v, expected %v", appErr.OriginalErr, originalErr)
	}
	if appErr.Suggestion == "" {
		t.Error("NewAppError() Suggestion should not be empty")
	}
}

func TestNewAppErrorWithSeverity(t *testing.T) {
	appErr := NewAppError(ErrorTypeNetwork, SeverityError, "NETWORK_ERROR", "test message")

	if appErr.Type != ErrorTypeNetwork {
		t.Errorf("NewAppError() Type = %v, expected %v", appErr.Type, ErrorTypeNetwork)
	}
	if appErr.Severity != SeverityError {
		t.Errorf("NewAppError() Severity = %v, expected %v", appErr.Severity, SeverityError)
	}
	if appErr.Code != "NETWORK_ERROR" {
		t.Errorf("NewAppError() Code = %q, expected %q", appErr.Code, "NETWORK_ERROR")
	}
	if appErr.Message != "test message" {
		t.Errorf("NewAppError() Message = %q, expected %q", appErr.Message, "test message")
	}
}

func TestAppError_WithCause(t *testing.T) {
	causeErr := errors.New("cause error")
	appErr := NewAppError(ErrorTypeNetwork, SeverityError, "TEST", "test")

	result := appErr.WithCause(causeErr)

	if result.OriginalErr != causeErr {
		t.Errorf("WithCause() OriginalErr = %v, want %v", result.OriginalErr, causeErr)
	}
	if result != appErr {
		t.Error("WithCause() should return the same instance")
	}
}

func TestAppError_WithContext(t *testing.T) {
	appErr := NewAppError(ErrorTypeNetwork, SeverityError, "TEST", "test")

	result := appErr.WithContext("key", "value")

	if result.Context["key"] != "value" {
		t.Errorf("WithContext() Context[key] = %v, want %v", result.Context["key"], "value")
	}
	if result != appErr {
		t.Error("WithContext() should return the same instance")
	}
}

func TestAppError_WithSolution(t *testing.T) {
	appErr := NewAppError(ErrorTypeNetwork, SeverityError, "TEST", "test")
	solution := "test solution"

	result := appErr.WithSolution(solution)

	if len(result.Solutions) != 1 || result.Solutions[0] != solution {
		t.Errorf("WithSolution() Solutions = %v, want [%v]", result.Solutions, solution)
	}
	if result != appErr {
		t.Error("WithSolution() should return the same instance")
	}
}

func TestAppError_WithHelpURL(t *testing.T) {
	appErr := NewAppError(ErrorTypeNetwork, SeverityError, "TEST", "test")
	helpURL := "https://example.com/help"

	result := appErr.WithHelpURL(helpURL)

	if result.HelpURL != helpURL {
		t.Errorf("WithHelpURL() HelpURL = %v, want %v", result.HelpURL, helpURL)
	}
	if result != appErr {
		t.Error("WithHelpURL() should return the same instance")
	}
}

func TestGetErrorInfo(t *testing.T) {
	tests := []struct {
		name               ErrorType
		expectedSuggestion string
		expectedStatusCode int
	}{
		{NetworkError, "Check your network connection and verify the gateway URL is correct.", 0},
		{AuthenticationError, "Check if your API key is correct and has the necessary permissions.", http.StatusUnauthorized},
		{AuthorizationError, "Your API key may not have permission to access model information.", http.StatusForbidden},
		{NotFoundError, "The endpoint may not support LiteLLM compatibility. Check if the URL is correct.", http.StatusNotFound},
		{RateLimitError, "Rate limit exceeded. Try again later or check your usage limits.", http.StatusTooManyRequests},
		{ServerError, "Server error. The gateway service may be experiencing issues. Try again later.", http.StatusInternalServerError},
		{ConfigError, "Check your configuration file and command line arguments.", 0},
		{ValidationError, "Check your input parameters and try again.", 0},
		{UnknownError, "An unexpected error occurred. Please try again.", 0},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.name)), func(t *testing.T) {
			suggestion, statusCode := getErrorInfo(tt.name)
			if suggestion != tt.expectedSuggestion {
				t.Errorf("getErrorInfo() suggestion = %q, expected %q", suggestion, tt.expectedSuggestion)
			}
			if statusCode != tt.expectedStatusCode {
				t.Errorf("getErrorInfo() statusCode = %d, expected %d", statusCode, tt.expectedStatusCode)
			}
		})
	}
}

func TestHandler_Handle(t *testing.T) {
	// 標準エラー出力をキャプチャ
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// ハンドラーの作成
	handler := NewHandler(false)

	// AppErrorのテスト
	appErr := NewAppErrorCompat(NetworkError, "test network error", errors.New("original"))
	exitCode := handler.Handle(appErr)

	// 出力のキャプチャを終了
	w.Close()
	os.Stderr = oldStderr

	// 出力内容の読み取り
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// 出力の検証
	if !strings.Contains(output, "❌ test network error") {
		t.Errorf("Expected error message not found in output: %s", output)
	}
	if !strings.Contains(output, "💡 解決策: Check your network connection") {
		t.Errorf("Expected suggestion not found in output: %s", output)
	}
	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}
}

func TestHandler_HandleWithVerbose(t *testing.T) {
	// 標準エラー出力をキャプチャ
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// ハンドラーの作成（詳細モード）
	handler := NewHandler(true)

	// AppErrorのテスト
	appErr := NewAppErrorCompat(NetworkError, "test network error", errors.New("original"))
	exitCode := handler.Handle(appErr)

	// 出力のキャプチャを終了
	w.Close()
	os.Stderr = oldStderr

	// 出力内容の読み取り
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// 詳細情報の検証
	if !strings.Contains(output, "🔍 詳細情報:") {
		t.Errorf("Expected verbose info not found in output: %s", output)
	}
	if !strings.Contains(output, "原因エラー: original") {
		t.Errorf("Expected cause error not found in output: %s", output)
	}
	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}
}

func TestHandler_HandleWithFallback(t *testing.T) {
	// 標準エラー出力をキャプチャ
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// ハンドラーの作成
	handler := NewHandler(false)

	// エラーとフォールバック関数
	originalErr := errors.New("original error")
	fallbackCalled := false
	fallback := func() error {
		fallbackCalled = true
		return nil
	}

	exitCode := handler.HandleWithFallback(originalErr, fallback)

	// 出力のキャプチャを終了
	w.Close()
	os.Stderr = oldStderr

	// 出力内容の読み取り
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// フォールバック処理の検証
	if !fallbackCalled {
		t.Error("Fallback function was not called")
	}
	if !strings.Contains(output, "🔄 フォールバック処理を実行します...") {
		t.Errorf("Expected fallback message not found in output: %s", output)
	}
	if !strings.Contains(output, "✅ フォールバック処理が成功しました") {
		t.Errorf("Expected fallback success message not found in output: %s", output)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestFormatErrorMessage(t *testing.T) {
	appErr := NewAppError(ErrorTypeNetwork, SeverityError, "NETWORK_ERROR", "test message").
		WithContext("url", "https://example.com").
		WithSolution("Check network connection").
		WithSolution("Verify URL").
		WithHelpURL("https://example.com/help")

	formatted := FormatErrorMessage(appErr)

	expectedParts := []string{
		"❌ test message",
		"📋 詳細情報:",
		"url: https://example.com",
		"💡 解決策:",
		"1. Check network connection",
		"2. Verify URL",
		"📖 詳細なヘルプ: https://example.com/help",
	}

	for _, part := range expectedParts {
		if !strings.Contains(formatted, part) {
			t.Errorf("Expected part '%s' not found in formatted message: %s", part, formatted)
		}
	}
}

func TestAsAppError(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := NewAppErrorCompat(NetworkError, "test error", originalErr)

	var target *AppError
	result := AsAppError(appErr, &target)

	if !result {
		t.Error("AsAppError() should return true for AppError")
	}
	if target != appErr {
		t.Error("AsAppError() should set target to the AppError")
	}

	// 一般的なエラーの場合
	genericErr := errors.New("generic error")
	var target2 *AppError
	result2 := AsAppError(genericErr, &target2)

	if result2 {
		t.Error("AsAppError() should return false for generic error")
	}
	if target2 != nil {
		t.Error("AsAppError() should not set target for generic error")
	}
}

func TestHandleError(t *testing.T) {
	// 標準エラー出力をキャプチャ
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// AppErrorのテスト
	appErr := NewAppErrorCompat(NetworkError, "test network error", errors.New("original"))
	HandleError(appErr)

	// 出力のキャプチャを終了
	w.Close()
	os.Stderr = oldStderr

	// 出力内容の読み取り
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// 出力の検証
	if !strings.Contains(output, "❌ Error: test network error") {
		t.Errorf("Expected error message not found in output: %s", output)
	}
	if !strings.Contains(output, "💡 Suggestion: Check your network connection") {
		t.Errorf("Expected suggestion not found in output: %s", output)
	}
}

func TestHandleErrorWithDebug(t *testing.T) {
	// デバッグ環境変数を設定
	os.Setenv("LLM_INFO_DEBUG", "1")
	defer os.Unsetenv("LLM_INFO_DEBUG")

	// 標準エラー出力をキャプチャ
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// AppErrorのテスト
	appErr := NewAppErrorCompat(NetworkError, "test network error", errors.New("original"))
	HandleError(appErr)

	// 出力のキャプチャを終了
	w.Close()
	os.Stderr = oldStderr

	// 出力内容の読み取り
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// デバッグ情報の検証
	if !strings.Contains(output, "🔍 Debug: original") {
		t.Errorf("Expected debug message not found in output: %s", output)
	}
}

func TestHandleGenericError(t *testing.T) {
	// 標準エラー出力をキャプチャ
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// 一般的なエラーのテスト
	genericErr := errors.New("timeout occurred")
	HandleError(genericErr)

	// 出力のキャプチャを終了
	w.Close()
	os.Stderr = oldStderr

	// 出力内容の読み取り
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// 出力の検証
	if !strings.Contains(output, "❌ Error: timeout occurred") {
		t.Errorf("Expected error message not found in output: %s", output)
	}
	if !strings.Contains(output, "💡 Suggestion: Try increasing the timeout") {
		t.Errorf("Expected suggestion not found in output: %s", output)
	}
}

func TestErrorConstructors(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name         string
		constructor  func(string, error) *AppError
		expectedType ErrorType
	}{
		{"NewNetworkError", NewNetworkError, NetworkError},
		{"NewAuthenticationError", NewAuthenticationError, AuthenticationError},
		{"NewAuthorizationError", NewAuthorizationError, AuthorizationError},
		{"NewNotFoundError", NewNotFoundError, NotFoundError},
		{"NewRateLimitError", NewRateLimitError, RateLimitError},
		{"NewServerError", NewServerError, ServerError},
		{"NewConfigError", NewConfigError, ConfigError},
		{"NewValidationError", NewValidationError, ValidationError},
		{"NewUnknownError", NewUnknownError, UnknownError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := tt.constructor("test message", originalErr)
			if appErr.Type != tt.expectedType {
				t.Errorf("%s() Type = %v, expected %v", tt.name, appErr.Type, tt.expectedType)
			}
			if appErr.Message != "test message" {
				t.Errorf("%s() Message = %q, expected %q", tt.name, appErr.Message, "test message")
			}
			if appErr.OriginalErr != originalErr {
				t.Errorf("%s() OriginalErr = %v, expected %v", tt.name, appErr.OriginalErr, originalErr)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := WrapError(originalErr, ConfigError, "wrapped message")

	if wrappedErr.Type != ConfigError {
		t.Errorf("WrapError() Type = %v, expected %v", wrappedErr.Type, ConfigError)
	}
	if wrappedErr.Message != "wrapped message" {
		t.Errorf("WrapError() Message = %q, expected %q", wrappedErr.Message, "wrapped message")
	}
	if wrappedErr.OriginalErr != originalErr {
		t.Errorf("WrapError() OriginalErr = %v, expected %v", wrappedErr.OriginalErr, originalErr)
	}
}
