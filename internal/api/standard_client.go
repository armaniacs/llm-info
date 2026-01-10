package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// StandardResponse はOpenAI標準APIのレスポンス形式を表す
type StandardResponse struct {
	Object string `json:"object"`
	Data   []struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

// FetchStandardModels はOpenAI標準エンドポイントからモデル情報を取得する
func (c *Client) FetchStandardModels() (*StandardResponse, error) {
	endpoint := fmt.Sprintf("%s/v1/models", c.baseURL)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

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

	var result StandardResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// convertStandardResponse はOpenAI標準レスポンスを内部形式に変換する
func (c *Client) convertStandardResponse(resp *StandardResponse) *ModelInfoResponse {
	var models []ModelInfo
	for _, data := range resp.Data {
		models = append(models, ModelInfo{
			ID:        data.ID,
			MaxTokens: 0,      // 標準APIでは提供されない
			Mode:      "chat", // デフォルト値
			InputCost: 0,      // 標準APIでは提供されない
		})
	}
	return &ModelInfoResponse{Models: models}
}
