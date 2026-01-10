package error

import (
	"strings"
)

// ErrorMessage はエラーメッセージを管理する
type ErrorMessage struct {
	Type    ErrorType
	Code    string
	Message string
}

var errorMessages = map[ErrorType]map[string]string{
	ErrorTypeNetwork: {
		"connection_timeout":    "接続がタイムアウトしました",
		"dns_resolution_failed": "DNS解決に失敗しました",
		"tls_certificate_error": "TLS証明書エラーが発生しました",
		"connection_refused":    "接続が拒否されました",
		"unknown_host":          "不明なホストです",
	},
	ErrorTypeAPI: {
		"authentication_failed": "認証に失敗しました",
		"rate_limit_exceeded":   "レート制限を超えました",
		"endpoint_not_found":    "エンドポイントが見つかりません",
		"invalid_response":      "無効なレスポンス形式です",
		"server_error":          "サーバーエラーが発生しました",
	},
	ErrorTypeConfig: {
		"config_file_not_found":  "設定ファイルが見つかりません",
		"invalid_config_format":  "設定ファイルの形式が無効です",
		"missing_required_field": "必須項目が設定されていません",
		"invalid_env_variable":   "環境変数の値が無効です",
	},
	ErrorTypeUser: {
		"invalid_argument":      "無効な引数です",
		"invalid_filter_syntax": "フィルタ構文が無効です",
		"invalid_sort_field":    "無効なソートフィールドです",
		"gateway_not_found":     "指定されたゲートウェイが見つかりません",
	},
	ErrorTypeSystem: {
		"permission_denied":   "ファイルアクセス権限がありません",
		"disk_full":           "ディスク容量が不足しています",
		"memory_insufficient": "メモリが不足しています",
		"unexpected_error":    "予期せぬエラーが発生しました",
	},
}

// GetErrorMessage はエラーメッセージを取得する
func GetErrorMessage(errorType ErrorType, code string) string {
	if messages, ok := errorMessages[errorType]; ok {
		if message, ok := messages[code]; ok {
			return message
		}
	}
	return "不明なエラーが発生しました"
}

// CreateNetworkError はネットワークエラーを作成する
func CreateNetworkError(code string, urlStr string, cause error) *AppError {
	message := GetErrorMessage(ErrorTypeNetwork, code)
	err := NewAppError(ErrorTypeNetwork, SeverityError, code, message).
		WithCause(cause).
		WithContext("url", urlStr)

	// エラーコードごとの解決策を追加
	switch code {
	case "connection_timeout":
		err = err.WithSolution("ネットワーク接続を確認してください").
			WithSolution("ファイアウォール設定を確認してください").
			WithSolution("ゲートウェイURLが正しいか確認してください")
	case "dns_resolution_failed":
		err = err.WithSolution("DNS設定を確認してください").
			WithSolution("ホスト名が正しいか確認してください").
			WithSolution("インターネット接続を確認してください")
	case "tls_certificate_error":
		err = err.WithSolution("サーバーの証明書が有効か確認してください").
			WithSolution("システムの時刻が正しいか確認してください")
	case "connection_refused":
		err = err.WithSolution("ゲートウェイサーバーが起動しているか確認してください").
			WithSolution("ポート番号が正しいか確認してください").
			WithSolution("ファイアウォール設定を確認してください")
	}

	return err.WithHelpURL("https://github.com/your-org/llm-info/wiki/network-errors")
}

// CreateAPIError はAPIエラーを作成する
func CreateAPIError(code string, statusCode int, urlStr string, cause error) *AppError {
	message := GetErrorMessage(ErrorTypeAPI, code)
	err := NewAppError(ErrorTypeAPI, SeverityError, code, message).
		WithCause(cause).
		WithContext("url", urlStr).
		WithContext("status_code", statusCode)

	// エラーコードごとの解決策を追加
	switch code {
	case "authentication_failed":
		err = err.WithSolution("APIキーが正しいか確認してください").
			WithSolution("APIキーの有効期限が切れていないか確認してください").
			WithSolution("APIキーの権限設定を確認してください")
	case "rate_limit_exceeded":
		err = err.WithSolution("しばらく待ってから再試行してください").
			WithSolution("APIプランのレート制限を確認してください").
			WithSolution("並列リクエスト数を減らしてください")
	case "endpoint_not_found":
		err = err.WithSolution("ゲートウェイURLが正しいか確認してください").
			WithSolution("APIバージョンが正しいか確認してください").
			WithSolution("エンドポイントパスを確認してください")
	}

	return err.WithHelpURL("https://github.com/your-org/llm-info/wiki/api-errors")
}

// CreateConfigError は設定エラーを作成する
func CreateConfigError(code string, configPath string, cause error) *AppError {
	message := GetErrorMessage(ErrorTypeConfig, code)
	err := NewAppError(ErrorTypeConfig, SeverityError, code, message).
		WithCause(cause).
		WithContext("config_path", configPath)

	// エラーコードごとの解決策を追加
	switch code {
	case "config_file_not_found":
		err = err.WithSolution("設定ファイルを作成してください: llm-info --init-config").
			WithSolution("設定ファイルパスが正しいか確認してください").
			WithSolution("環境変数LLM_INFO_CONFIG_PATHを確認してください")
	case "invalid_config_format":
		err = err.WithSolution("YAML形式が正しいか確認してください").
			WithSolution("設定ファイルの構文を確認してください").
			WithSolution("例: https://github.com/your-org/llm-info/blob/main/configs/example.yaml")
	case "missing_required_field":
		err = err.WithSolution("必須項目（url, api_keyなど）が設定されているか確認してください").
			WithSolution("設定ファイルのテンプレートを確認してください")
	}

	return err.WithHelpURL("https://github.com/your-org/llm-info/wiki/config-errors")
}

// CreateUserError はユーザーエラーを作成する
func CreateUserError(code string, argument string, cause error) *AppError {
	message := GetErrorMessage(ErrorTypeUser, code)
	err := NewAppError(ErrorTypeUser, SeverityWarning, code, message).
		WithCause(cause).
		WithContext("argument", argument)

	// エラーコードごとの解決策を追加
	switch code {
	case "invalid_argument":
		err = err.WithSolution("コマンドライン引数が正しいか確認してください").
			WithSolution("ヘルプを確認してください: llm-info --help")
	case "invalid_filter_syntax":
		err = err.WithSolution("フィルタ構文を確認してください").
			WithSolution("例: --filter \"name:gpt,tokens>1000\"").
			WithSolution("ヘルプを確認してください: llm-info --help filter")
	case "invalid_sort_field":
		err = err.WithSolution("ソートフィールドを確認してください").
			WithSolution("使用可能なフィールド: name, tokens, cost, mode").
			WithSolution("ヘルプを確認してください: llm-info --help sort")
	case "gateway_not_found":
		err = err.WithSolution("ゲートウェイ名が正しいか確認してください").
			WithSolution("利用可能なゲートウェイを確認してください: llm-info --list-gateways")
	}

	return err.WithHelpURL("https://github.com/your-org/llm-info/wiki/usage")
}

// CreateSystemError はシステムエラーを作成する
func CreateSystemError(code string, context string, cause error) *AppError {
	message := GetErrorMessage(ErrorTypeSystem, code)
	err := NewAppError(ErrorTypeSystem, SeverityError, code, message).
		WithCause(cause).
		WithContext("context", context)

	// エラーコードごとの解決策を追加
	switch code {
	case "permission_denied":
		err = err.WithSolution("ファイルのアクセス権限を確認してください").
			WithSolution("管理者権限で実行してください").
			WithSolution("ファイルの所有者を確認してください")
	case "disk_full":
		err = err.WithSolution("ディスク容量を確認してください").
			WithSolution("不要なファイルを削除してください").
			WithSolution("別のストレージを使用してください")
	case "memory_insufficient":
		err = err.WithSolution("メモリ使用量を確認してください").
			WithSolution("他のアプリケーションを終了してください").
			WithSolution("システムを再起動してください")
	case "unexpected_error":
		err = err.WithSolution("開発者にバグレポートを送信してください").
			WithSolution("詳細なログを確認してください: llm-info --verbose").
			WithSolution("最新バージョンにアップデートしてください")
	}

	return err.WithHelpURL("https://github.com/your-org/llm-info/issues")
}

// DetectErrorType はエラーメッセージからエラー種別を検出する
func DetectErrorType(err error) (ErrorType, string) {
	if err == nil {
		return ErrorTypeUnknown, "unknown_error"
	}

	errorMsg := strings.ToLower(err.Error())

	// ネットワークエラーの検出
	if strings.Contains(errorMsg, "timeout") || strings.Contains(errorMsg, "deadline exceeded") {
		return ErrorTypeNetwork, "connection_timeout"
	}
	if strings.Contains(errorMsg, "connection refused") {
		return ErrorTypeNetwork, "connection_refused"
	}
	if strings.Contains(errorMsg, "no such host") || strings.Contains(errorMsg, "dns") {
		return ErrorTypeNetwork, "dns_resolution_failed"
	}
	if strings.Contains(errorMsg, "certificate") || strings.Contains(errorMsg, "tls") {
		return ErrorTypeNetwork, "tls_certificate_error"
	}

	// APIエラーの検出
	if strings.Contains(errorMsg, "401") || strings.Contains(errorMsg, "unauthorized") {
		return ErrorTypeAPI, "authentication_failed"
	}
	if strings.Contains(errorMsg, "403") || strings.Contains(errorMsg, "forbidden") {
		return ErrorTypeAPI, "authorization_failed"
	}
	if strings.Contains(errorMsg, "404") || strings.Contains(errorMsg, "not found") {
		return ErrorTypeAPI, "endpoint_not_found"
	}
	if strings.Contains(errorMsg, "429") || strings.Contains(errorMsg, "rate limit") {
		return ErrorTypeAPI, "rate_limit_exceeded"
	}
	if strings.Contains(errorMsg, "500") || strings.Contains(errorMsg, "502") || strings.Contains(errorMsg, "503") {
		return ErrorTypeAPI, "server_error"
	}

	// 設定エラーの検出
	if strings.Contains(errorMsg, "no such file") || strings.Contains(errorMsg, "file not found") {
		return ErrorTypeConfig, "config_file_not_found"
	}
	if strings.Contains(errorMsg, "yaml") || strings.Contains(errorMsg, "parse") {
		return ErrorTypeConfig, "invalid_config_format"
	}

	// システムエラーの検出
	if strings.Contains(errorMsg, "permission denied") {
		return ErrorTypeSystem, "permission_denied"
	}

	return ErrorTypeUnknown, "unexpected_error"
}

// WrapErrorWithDetection はエラーを検出して適切なAppErrorでラップする
func WrapErrorWithDetection(err error, context string) *AppError {
	if err == nil {
		return nil
	}

	errorType, code := DetectErrorType(err)

	switch errorType {
	case ErrorTypeNetwork:
		return CreateNetworkError(code, context, err)
	case ErrorTypeAPI:
		return CreateAPIError(code, 0, context, err)
	case ErrorTypeConfig:
		return CreateConfigError(code, context, err)
	case ErrorTypeUser:
		return CreateUserError(code, context, err)
	case ErrorTypeSystem:
		return CreateSystemError(code, context, err)
	default:
		return CreateSystemError("unexpected_error", context, err)
	}
}
