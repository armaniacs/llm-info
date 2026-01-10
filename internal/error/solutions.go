package error

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"
)

// SolutionProvider は解決策を提供する
type SolutionProvider struct {
	osInfo string
}

// NewSolutionProvider は新しい解決策プロバイダーを作成する
func NewSolutionProvider() *SolutionProvider {
	return &SolutionProvider{
		osInfo: getOSInfo(),
	}
}

// GetGeneralSolutions は一般的な解決策を返す
func (sp *SolutionProvider) GetGeneralSolutions() []string {
	solutions := []string{
		"最新バージョンにアップデートしてください: llm-info --version",
		"詳細なログを確認してください: llm-info --verbose",
		"設定ファイルを確認してください: llm-info --check-config",
	}

	// OS固有の解決策を追加
	if strings.Contains(sp.osInfo, "windows") {
		solutions = append(solutions, "Windowsファイアウォール設定を確認してください")
	} else {
		solutions = append(solutions, "iptables設定を確認してください")
	}

	return solutions
}

// GetNetworkSolutions はネットワーク関連の解決策を返す
func (sp *SolutionProvider) GetNetworkSolutions(urlStr string) []string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return sp.GetGeneralSolutions()
	}

	solutions := []string{
		fmt.Sprintf("ホスト %s に到達できるか確認してください", parsedURL.Host),
		"プロキシ設定を確認してください",
		"DNS設定を確認してください",
	}

	if parsedURL.Scheme == "https" {
		solutions = append(solutions, "TLS/SSL設定を確認してください")
	}

	return solutions
}

// GetConfigSolutions は設定関連の解決策を返す
func (sp *SolutionProvider) GetConfigSolutions(configPath string) []string {
	return []string{
		fmt.Sprintf("設定ファイル %s が存在するか確認してください", configPath),
		"設定ファイルのパーミッションを確認してください",
		"設定ファイルのYAML構文を確認してください",
		"llm-info --init-config で初期設定を作成してください",
	}
}

// GetAPISolutions はAPI関連の解決策を返す
func (sp *SolutionProvider) GetAPISolutions(statusCode int, urlStr string) []string {
	solutions := []string{}

	switch statusCode {
	case 401:
		solutions = append(solutions, "APIキーが正しいか確認してください")
		solutions = append(solutions, "APIキーの有効期限が切れていないか確認してください")
		solutions = append(solutions, "APIキーの権限設定を確認してください")
	case 403:
		solutions = append(solutions, "APIキーに必要な権限があるか確認してください")
		solutions = append(solutions, "アカウントの利用制限を確認してください")
	case 404:
		solutions = append(solutions, "ゲートウェイURLが正しいか確認してください")
		solutions = append(solutions, "APIバージョンが正しいか確認してください")
		solutions = append(solutions, "エンドポイントパスを確認してください")
	case 429:
		solutions = append(solutions, "しばらく待ってから再試行してください")
		solutions = append(solutions, "APIプランのレート制限を確認してください")
		solutions = append(solutions, "並列リクエスト数を減らしてください")
	case 500, 502, 503, 504:
		solutions = append(solutions, "サーバーが一時的に利用できない可能性があります")
		solutions = append(solutions, "しばらく待ってから再試行してください")
		solutions = append(solutions, "サービスステータスページを確認してください")
	default:
		solutions = append(solutions, "APIドキュメントを確認してください")
		solutions = append(solutions, "サポートにお問い合わせください")
	}

	return solutions
}

// GetUserSolutions はユーザーエラー関連の解決策を返す
func (sp *SolutionProvider) GetUserSolutions(code string, argument string) []string {
	solutions := []string{}

	switch code {
	case "invalid_argument":
		solutions = append(solutions, "コマンドライン引数が正しいか確認してください")
		solutions = append(solutions, "ヘルプを確認してください: llm-info --help")
	case "invalid_filter_syntax":
		solutions = append(solutions, "フィルタ構文を確認してください")
		solutions = append(solutions, "例: --filter \"name:gpt,tokens>1000\"")
		solutions = append(solutions, "ヘルプを確認してください: llm-info --help filter")
	case "invalid_sort_field":
		solutions = append(solutions, "ソートフィールドを確認してください")
		solutions = append(solutions, "使用可能なフィールド: name, tokens, cost, mode")
		solutions = append(solutions, "ヘルプを確認してください: llm-info --help sort")
	case "gateway_not_found":
		solutions = append(solutions, "ゲートウェイ名が正しいか確認してください")
		solutions = append(solutions, "利用可能なゲートウェイを確認してください: llm-info --list-gateways")
		solutions = append(solutions, "設定ファイルにゲートウェイが登録されているか確認してください")
	}

	return solutions
}

// GetSystemSolutions はシステムエラー関連の解決策を返す
func (sp *SolutionProvider) GetSystemSolutions(code string, context string) []string {
	solutions := []string{}

	switch code {
	case "permission_denied":
		solutions = append(solutions, "ファイルのアクセス権限を確認してください")
		if strings.Contains(sp.osInfo, "windows") {
			solutions = append(solutions, "管理者として実行してください")
		} else {
			solutions = append(solutions, "sudoを使用して実行してください")
		}
		solutions = append(solutions, "ファイルの所有者を確認してください")
	case "disk_full":
		solutions = append(solutions, "ディスク容量を確認してください")
		solutions = append(solutions, "不要なファイルを削除してください")
		solutions = append(solutions, "別のストレージを使用してください")
	case "memory_insufficient":
		solutions = append(solutions, "メモリ使用量を確認してください")
		solutions = append(solutions, "他のアプリケーションを終了してください")
		solutions = append(solutions, "システムを再起動してください")
	case "unexpected_error":
		solutions = append(solutions, "開発者にバグレポートを送信してください")
		solutions = append(solutions, "詳細なログを確認してください: llm-info --verbose")
		solutions = append(solutions, "最新バージョンにアップデートしてください")
	}

	return solutions
}

// GetContextualSolutions はコンテキストに応じた解決策を返す
func (sp *SolutionProvider) GetContextualSolutions(err *AppError) []string {
	switch err.Type {
	case ErrorTypeNetwork:
		if url, ok := err.Context["url"].(string); ok {
			return sp.GetNetworkSolutions(url)
		}
		return sp.GetNetworkSolutions("")
	case ErrorTypeAPI:
		if statusCode, ok := err.Context["status_code"].(int); ok {
			if url, ok := err.Context["url"].(string); ok {
				solutions := sp.GetAPISolutions(statusCode, url)
				if len(solutions) > 0 {
					return solutions
				}
			}
			return sp.GetAPISolutions(statusCode, "")
		}
		return sp.GetAPISolutions(0, "")
	case ErrorTypeConfig:
		if configPath, ok := err.Context["config_path"].(string); ok {
			return sp.GetConfigSolutions(configPath)
		}
		return sp.GetConfigSolutions("")
	case ErrorTypeUser:
		if argument, ok := err.Context["argument"].(string); ok {
			return sp.GetUserSolutions(err.Code, argument)
		}
		return sp.GetUserSolutions(err.Code, "")
	case ErrorTypeSystem:
		if context, ok := err.Context["context"].(string); ok {
			return sp.GetSystemSolutions(err.Code, context)
		}
		return sp.GetSystemSolutions(err.Code, "")
	default:
		return sp.GetGeneralSolutions()
	}
}

// GetHelpURL はエラー種別に応じたヘルプURLを返す
func (sp *SolutionProvider) GetHelpURL(errorType ErrorType) string {
	switch errorType {
	case ErrorTypeNetwork:
		return "https://github.com/armaniacs/llm-info/wiki/network-errors"
	case ErrorTypeAPI:
		return "https://github.com/armaniacs/llm-info/wiki/api-errors"
	case ErrorTypeConfig:
		return "https://github.com/armaniacs/llm-info/wiki/config-errors"
	case ErrorTypeUser:
		return "https://github.com/armaniacs/llm-info/wiki/usage"
	case ErrorTypeSystem:
		return "https://github.com/armaniacs/llm-info/wiki/troubleshooting"
	default:
		return "https://github.com/armaniacs/llm-info/wiki/errors"
	}
}

// getOSInfo はOS情報を取得する
func getOSInfo() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}

// EnhanceError はエラーに解決策とヘルプURLを追加する
func (sp *SolutionProvider) EnhanceError(err *AppError) *AppError {
	if err == nil {
		return nil
	}

	// 解決策が設定されていない場合は自動生成
	if len(err.Solutions) == 0 {
		solutions := sp.GetContextualSolutions(err)
		for _, solution := range solutions {
			err = err.WithSolution(solution)
		}
	}

	// ヘルプURLが設定されていない場合は自動設定
	if err.HelpURL == "" {
		err = err.WithHelpURL(sp.GetHelpURL(err.Type))
	}

	return err
}
