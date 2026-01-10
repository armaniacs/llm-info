package error

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
)

// ErrorType はエラーの種類を表します
type ErrorType int

const (
	// ErrorTypeUnknown は不明なエラー
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeNetwork はネットワーク関連のエラー
	ErrorTypeNetwork
	// ErrorTypeAPI はAPI関連のエラー
	ErrorTypeAPI
	// ErrorTypeConfig は設定関連のエラー
	ErrorTypeConfig
	// ErrorTypeUser はユーザーエラー
	ErrorTypeUser
	// ErrorTypeSystem はシステムエラー
	ErrorTypeSystem

	// 互換性のための古い定義
	NetworkError ErrorType = iota + 1
	AuthenticationError
	AuthorizationError
	NotFoundError
	RateLimitError
	ServerError
	ConfigError
	ValidationError
	UnknownError
)

// ErrorSeverity はエラーの重大度を表す
type ErrorSeverity int

const (
	SeverityInfo ErrorSeverity = iota
	SeverityWarning
	SeverityError
	SeverityFatal
)

// AppError はアプリケーションエラーを表します
type AppError struct {
	Type        ErrorType
	Severity    ErrorSeverity
	Code        string
	Message     string
	OriginalErr error
	Suggestion  string
	StatusCode  int
	Context     map[string]interface{}
	Solutions   []string
	HelpURL     string
}

// Error はエラーメッセージを返します
func (e *AppError) Error() string {
	if e.OriginalErr != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.OriginalErr)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap は元のエラーを返します
func (e *AppError) Unwrap() error {
	return e.OriginalErr
}

// NewAppError は新しいアプリケーションエラーを作成します
func NewAppError(errorType ErrorType, severity ErrorSeverity, code, message string) *AppError {
	return &AppError{
		Type:     errorType,
		Severity: severity,
		Code:     code,
		Message:  message,
		Context:  make(map[string]interface{}),
	}
}

// 互換性のための古いNewAppError関数
func NewAppErrorCompat(errorType ErrorType, message string, originalErr error) *AppError {
	suggestion, statusCode := getErrorInfo(errorType)
	return &AppError{
		Type:        errorType,
		Severity:    SeverityError,
		Code:        getErrorCode(errorType),
		Message:     message,
		OriginalErr: originalErr,
		Suggestion:  suggestion,
		StatusCode:  statusCode,
		Context:     make(map[string]interface{}),
	}
}

// WithCause は原因エラーを設定する
func (e *AppError) WithCause(cause error) *AppError {
	e.OriginalErr = cause
	return e
}

// WithContext はコンテキスト情報を追加する
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithSolution は解決策を追加する
func (e *AppError) WithSolution(solution string) *AppError {
	e.Solutions = append(e.Solutions, solution)
	return e
}

// WithHelpURL はヘルプURLを設定する
func (e *AppError) WithHelpURL(url string) *AppError {
	e.HelpURL = url
	return e
}

// getErrorCode はエラー種別に応じたエラーコードを返します
func getErrorCode(errorType ErrorType) string {
	switch errorType {
	case ErrorTypeNetwork, NetworkError:
		return "NETWORK_ERROR"
	case ErrorTypeAPI:
		return "API_ERROR"
	case AuthenticationError:
		return "AUTHENTICATION_ERROR"
	case AuthorizationError:
		return "AUTHORIZATION_ERROR"
	case NotFoundError:
		return "NOT_FOUND_ERROR"
	case RateLimitError:
		return "RATE_LIMIT_ERROR"
	case ServerError:
		return "SERVER_ERROR"
	case ErrorTypeConfig, ConfigError:
		return "CONFIG_ERROR"
	case ErrorTypeUser, ValidationError:
		return "USER_ERROR"
	case ErrorTypeSystem:
		return "SYSTEM_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}

// getErrorInfo はエラー種別に応じた提案とステータスコードを返します
func getErrorInfo(errorType ErrorType) (string, int) {
	switch errorType {
	case ErrorTypeNetwork, NetworkError:
		return "Check your network connection and verify the gateway URL is correct.", 0
	case AuthenticationError:
		return "Check if your API key is correct and has the necessary permissions.", http.StatusUnauthorized
	case AuthorizationError:
		return "Your API key may not have permission to access model information.", http.StatusForbidden
	case NotFoundError:
		return "The endpoint may not support LiteLLM compatibility. Check if the URL is correct.", http.StatusNotFound
	case RateLimitError:
		return "Rate limit exceeded. Try again later or check your usage limits.", http.StatusTooManyRequests
	case ServerError:
		return "Server error. The gateway service may be experiencing issues. Try again later.", http.StatusInternalServerError
	case ErrorTypeConfig, ConfigError:
		return "Check your configuration file and command line arguments.", 0
	case ErrorTypeUser, ValidationError:
		return "Check your input parameters and try again.", 0
	default:
		return "An unexpected error occurred. Please try again.", 0
	}
}

// Handler はエラーハンドラーを表す
type Handler struct {
	verbose bool
}

// NewHandler は新しいエラーハンドラーを作成する
func NewHandler(verbose bool) *Handler {
	return &Handler{
		verbose: verbose,
	}
}

// Handle はエラーを処理して表示します
func (h *Handler) Handle(err error) int {
	if err == nil {
		return 0
	}

	var appErr *AppError
	if !AsAppError(err, &appErr) {
		// アプリケーションエラーでない場合はラップする
		appErr = NewAppError(ErrorTypeUnknown, SeverityError, "unexpected_error", "予期せぬエラーが発生しました").
			WithCause(err).
			WithSolution("開発者にエラーレポートを送信してください").
			WithHelpURL("https://github.com/your-org/llm-info/issues")
	}

	// エラーメッセージを表示
	fmt.Fprintln(os.Stderr, FormatErrorMessage(appErr))

	// 詳細モードの場合は追加情報を表示
	if h.verbose {
		h.printVerboseInfo(appErr)
	}

	// 重大度に応じて終了コードを返す
	switch appErr.Severity {
	case SeverityInfo:
		return 0
	case SeverityWarning:
		return 1
	case SeverityError:
		return 2
	case SeverityFatal:
		return 3
	default:
		return 2
	}
}

// HandleWithFallback はエラーを処理し、フォールバック処理を実行する
func (h *Handler) HandleWithFallback(err error, fallback func() error) int {
	if err == nil {
		return 0
	}

	// エラーを処理
	exitCode := h.Handle(err)

	// フォールバック処理を実行
	if fallback != nil {
		fmt.Fprintln(os.Stderr, "\n🔄 フォールバック処理を実行します...")
		if fallbackErr := fallback(); fallbackErr != nil {
			fmt.Fprintf(os.Stderr, "❌ フォールバック処理も失敗しました: %v\n", fallbackErr)
			return 3
		}
		fmt.Fprintln(os.Stderr, "✅ フォールバック処理が成功しました")
		return 0
	}

	return exitCode
}

// printVerboseInfo は詳細情報を表示する
func (h *Handler) printVerboseInfo(err *AppError) {
	fmt.Fprintln(os.Stderr, "\n🔍 詳細情報:")

	// スタックトレース
	if err.OriginalErr != nil {
		fmt.Fprintf(os.Stderr, "原因エラー: %v\n", err.OriginalErr)
	}

	// デバッグ情報
	fmt.Fprintf(os.Stderr, "エラータイプ: %v\n", err.Type)
	fmt.Fprintf(os.Stderr, "エラーコード: %s\n", err.Code)
	fmt.Fprintf(os.Stderr, "重大度: %v\n", err.Severity)

	// スタックトレース
	if h.verbose {
		fmt.Fprintln(os.Stderr, "\nスタックトレース:")
		debug.PrintStack()
	}
}

// AsAppError はエラーをAppErrorに変換する
func AsAppError(err error, target **AppError) bool {
	if err == nil {
		return false
	}

	if appErr, ok := err.(*AppError); ok {
		*target = appErr
		return true
	}

	// エラーチェーンをたどってAppErrorを探す
	for {
		if unwrapped := errors.Unwrap(err); unwrapped != nil {
			if appErr, ok := unwrapped.(*AppError); ok {
				*target = appErr
				return true
			}
			err = unwrapped
			continue
		}
		break
	}

	return false
}

// Recover はパニックから回復する
func (h *Handler) Recover() {
	if r := recover(); r != nil {
		err := NewAppError(ErrorTypeSystem, SeverityFatal, "panic", "アプリケーションがクラッシュしました").
			WithCause(fmt.Errorf("panic: %v", r)).
			WithSolution("開発者にバグレポートを送信してください").
			WithHelpURL("https://github.com/your-org/llm-info/issues")

		fmt.Printf("Panic recovered: %v\n", r)
		debug.PrintStack()

		os.Exit(h.Handle(err))
	}
}

// FormatErrorMessage はエラーメッセージをフォーマットする
func FormatErrorMessage(err *AppError) string {
	var builder strings.Builder

	// 基本メッセージ
	builder.WriteString(fmt.Sprintf("❌ %s\n", err.Message))

	// コンテキスト情報
	if len(err.Context) > 0 {
		builder.WriteString("\n📋 詳細情報:\n")
		for key, value := range err.Context {
			builder.WriteString(fmt.Sprintf("   %s: %v\n", key, value))
		}
	}

	// 解決策
	if len(err.Solutions) > 0 {
		builder.WriteString("\n💡 解決策:\n")
		for i, solution := range err.Solutions {
			builder.WriteString(fmt.Sprintf("   %d. %s\n", i+1, solution))
		}
	} else if err.Suggestion != "" {
		// 互換性のためのSuggestionフィールド
		builder.WriteString(fmt.Sprintf("\n💡 解決策: %s\n", err.Suggestion))
	}

	// ヘルプURL
	if err.HelpURL != "" {
		builder.WriteString(fmt.Sprintf("\n📖 詳細なヘルプ: %s\n", err.HelpURL))
	}

	return builder.String()
}

// HandleError はエラーを処理して表示します（互換性のため）
func HandleError(err error) {
	if err == nil {
		return
	}

	var appErr *AppError
	if isErrorType(err, &appErr) {
		handleAppError(appErr)
	} else {
		handleGenericError(err)
	}
}

// isErrorType はエラーが指定された型であるかチェックします
func isErrorType(err error, target interface{}) bool {
	switch err := err.(type) {
	case *AppError:
		if target, ok := target.(**AppError); ok {
			*target = err
			return true
		}
	}
	return false
}

// handleAppError はアプリケーションエラーを処理します
func handleAppError(err *AppError) {
	fmt.Fprintf(os.Stderr, "\n❌ Error: %s\n", err.Message)

	if err.Suggestion != "" {
		fmt.Fprintf(os.Stderr, "💡 Suggestion: %s\n", err.Suggestion)
	}

	// デバッグ情報の表示（環境変数で制御）
	if os.Getenv("LLM_INFO_DEBUG") != "" && err.OriginalErr != nil {
		fmt.Fprintf(os.Stderr, "🔍 Debug: %v\n", err.OriginalErr)
	}
}

// handleGenericError は一般的なエラーを処理します
func handleGenericError(err error) {
	errorMsg := err.Error()
	var suggestion string

	// エラーメッセージに基づいて提案を生成
	switch {
	case strings.Contains(errorMsg, "timeout"):
		suggestion = "Try increasing the timeout with --timeout flag or check your network connection."
	case strings.Contains(errorMsg, "connection refused"):
		suggestion = "Check if the gateway URL is correct and the service is running."
	case strings.Contains(errorMsg, "401"):
		suggestion = "Check if your API key is correct and has the necessary permissions."
	case strings.Contains(errorMsg, "403"):
		suggestion = "Your API key may not have permission to access model information."
	case strings.Contains(errorMsg, "404"):
		suggestion = "The endpoint may not support LiteLLM compatibility. Check if the URL is correct."
	case strings.Contains(errorMsg, "429"):
		suggestion = "Rate limit exceeded. Try again later or check your usage limits."
	case strings.Contains(errorMsg, "500"), strings.Contains(errorMsg, "502"), strings.Contains(errorMsg, "503"), strings.Contains(errorMsg, "504"):
		suggestion = "Server error. The gateway service may be experiencing issues. Try again later."
	case strings.Contains(errorMsg, "no such file"), strings.Contains(errorMsg, "not found"):
		suggestion = "Check if the file path is correct and the file exists."
	case strings.Contains(errorMsg, "permission denied"):
		suggestion = "Check if you have the necessary permissions to access the resource."
	default:
		suggestion = "Check your network connection and verify the gateway URL is correct."
	}

	fmt.Fprintf(os.Stderr, "\n❌ Error: %s\n", errorMsg)
	fmt.Fprintf(os.Stderr, "💡 Suggestion: %s\n", suggestion)

	// デバッグ情報の表示
	if os.Getenv("LLM_INFO_DEBUG") != "" {
		fmt.Fprintf(os.Stderr, "🔍 Debug: This is a generic error. Consider using AppError for better error handling.\n")
	}
}

// NewNetworkError はネットワークエラーを作成します
func NewNetworkError(message string, err error) *AppError {
	return NewAppErrorCompat(NetworkError, message, err)
}

// NewAuthenticationError は認証エラーを作成します
func NewAuthenticationError(message string, err error) *AppError {
	return NewAppErrorCompat(AuthenticationError, message, err)
}

// NewAuthorizationError は認可エラーを作成します
func NewAuthorizationError(message string, err error) *AppError {
	return NewAppErrorCompat(AuthorizationError, message, err)
}

// NewNotFoundError はNotFoundエラーを作成します
func NewNotFoundError(message string, err error) *AppError {
	return NewAppErrorCompat(NotFoundError, message, err)
}

// NewRateLimitError はレート制限エラーを作成します
func NewRateLimitError(message string, err error) *AppError {
	return NewAppErrorCompat(RateLimitError, message, err)
}

// NewServerError はサーバーエラーを作成します
func NewServerError(message string, err error) *AppError {
	return NewAppErrorCompat(ServerError, message, err)
}

// NewConfigError は設定エラーを作成します
func NewConfigError(message string, err error) *AppError {
	return NewAppErrorCompat(ConfigError, message, err)
}

// NewValidationError は検証エラーを作成します
func NewValidationError(message string, err error) *AppError {
	return NewAppErrorCompat(ValidationError, message, err)
}

// WrapError は既存のエラーをAppErrorでラップします
func WrapError(err error, errorType ErrorType, message string) *AppError {
	return NewAppErrorCompat(errorType, message, err)
}

// NewUnknownError は不明なエラーを作成します
func NewUnknownError(message string, err error) *AppError {
	return NewAppErrorCompat(UnknownError, message, err)
}
