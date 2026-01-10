package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/your-org/llm-info/internal/config"
)

// Client はAPIクライアントです
type Client struct {
	baseURL string
	apiKey  string
	timeout time.Duration
	client  *http.Client
}

// NewClient は新しいAPIクライアントを作成します
func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		timeout: cfg.Timeout,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// GetModelInfo はモデル情報を取得します
func (c *Client) GetModelInfo() (*ModelInfoResponse, error) {
	url := fmt.Sprintf("%s/model/info", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 認証ヘッダーの設定
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}
	req.Header.Set("Content-Type", "application/json")

	// リクエスト送信
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// ステータスコードのチェック
	if resp.StatusCode != http.StatusOK {
		// エラーレスポンスのボディを読み取って詳細なエラーメッセージを取得
		errorBody, _ := io.ReadAll(resp.Body)
		errorMsg := string(errorBody)

		// エラーメッセージが空の場合はステータスコードに基づくデフォルトメッセージを使用
		if errorMsg == "" {
			errorMsg = getDefaultStatusMessage(resp.StatusCode)
		}

		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, errorMsg)
	}

	// レスポンスボディを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 空のレスポンスのチェック
	if len(body) == 0 {
		return nil, fmt.Errorf("API returned empty response")
	}

	// レスポンスのパース
	var response ModelInfoResponse
	if err := json.Unmarshal(body, &response); err != nil {
		// JSONパースエラーの場合、レスポンスの一部を表示してデバッグを容易にする
		preview := string(body)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return nil, fmt.Errorf("failed to decode JSON response: %w. Response preview: %s", err, preview)
	}

	return &response, nil
}

// getDefaultStatusMessage はHTTPステータスコードに基づくデフォルトエラーメッセージを返します
func getDefaultStatusMessage(statusCode int) string {
	switch statusCode {
	case 400:
		return "Bad Request - Invalid parameters or request format"
	case 401:
		return "Unauthorized - Invalid or missing API key"
	case 403:
		return "Forbidden - Insufficient permissions to access this resource"
	case 404:
		return "Not Found - The requested endpoint does not exist"
	case 429:
		return "Too Many Requests - Rate limit exceeded"
	case 500:
		return "Internal Server Error - The server encountered an unexpected error"
	case 502:
		return "Bad Gateway - The server received an invalid response"
	case 503:
		return "Service Unavailable - The server is temporarily unavailable"
	case 504:
		return "Gateway Timeout - The server took too long to respond"
	default:
		return fmt.Sprintf("HTTP Error %d", statusCode)
	}
}

// FetchModelsWithFallback はLiteLLMエンドポイントを試行し、失敗した場合はOpenAI標準エンドポイントにフォールバックする
func (c *Client) FetchModelsWithFallback() (*ModelInfoResponse, error) {
	// まずLiteLLMエンドポイントを試行
	models, err := c.GetModelInfo()
	if err == nil {
		return models, nil
	}

	fmt.Printf("⚠️  LiteLLM endpoint failed, falling back to OpenAI standard endpoint: %v\n", err)

	// OpenAI標準エンドポイントを試行
	standardResp, err := c.FetchStandardModels()
	if err != nil {
		return nil, fmt.Errorf("both endpoints failed: LiteLLM error: %v, Standard error: %w", err, err)
	}

	// 標準レスポンスを内部形式に変換
	return c.convertStandardResponse(standardResp), nil
}
