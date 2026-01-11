package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/armaniacs/llm-info/pkg/config"
)

// ProbeClient はモデル制約値を探索するためのクライアント
type ProbeClient struct {
	client  *http.Client
	config  *config.AppConfig
}

// NewProbeClient は新しいProbeClientを作成する
func NewProbeClient(cfg *config.AppConfig) *ProbeClient {
	return &ProbeClient{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		config: cfg,
	}
}

// GetConfig は設定を返す
func (pc *ProbeClient) GetConfig() *config.AppConfig {
	return pc.config
}

// ProbeRequest はAPIリクエストの構造体
type ProbeRequest struct {
	Model       string `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// Message はメッセージの構造体
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ProbeResponse はAPIレスポンスの構造体
type ProbeResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []ChatChoice   `json:"choices"`
	Usage   *UsageInfo    `json:"usage"`
	Error   *OpenAIError   `json:"error"`
}

// ChatChoice は選択肢の構造体
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatMessage はチャットメッセージの構造体
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIError はエラー情報の構造体
type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// UsageInfo は使用量情報の構造体
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ProbeModel はモデルの制約値を探索する
func (pc *ProbeClient) ProbeModel(modelID string) (*ProbeResponse, error) {
	// リクエストを作成
	req := ProbeRequest{
		Model: modelID,
		Messages: []Message{
			{Role: "user", Content: "test"},
		},
		MaxTokens:   16,
		Temperature: 0,
	}

	// JSONにエンコード
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// HTTPリクエストを作成
	url := fmt.Sprintf("%s/v1/chat/completions", pc.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		url,
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// ヘッダーを設定
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pc.config.APIKey))

	// リクエストを送信
	resp, err := pc.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスを読み込む
	var probeResp ProbeResponse
	if err := json.NewDecoder(resp.Body).Decode(&probeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// ステータスコードをチェック
	if resp.StatusCode != http.StatusOK {
		if probeResp.Error != nil {
			return &probeResp, fmt.Errorf("API error (%s): %s", probeResp.Error.Type, probeResp.Error.Message)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return &probeResp, nil
}