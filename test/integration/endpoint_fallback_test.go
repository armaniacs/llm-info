package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/armaniacs/llm-info/internal/api"
	"github.com/armaniacs/llm-info/internal/config"
)

func TestEndpointFallbackIntegration(t *testing.T) {
	tests := []struct {
		name              string
		litellmAvailable  bool
		standardAvailable bool
		expectedModels    int
		expectFallback    bool
		expectError       bool
	}{
		{
			name:              "Both endpoints available, LiteLLM enhances",
			litellmAvailable:  true,
			standardAvailable: true,
			expectedModels:    2,
			expectFallback:    false,
			expectError:       false,
		},
		{
			name:              "Standard available, LiteLLM unavailable",
			litellmAvailable:  false,
			standardAvailable: true,
			expectedModels:    1,
			expectFallback:    false,
			expectError:       false,
		},
		{
			name:              "Standard unavailable, LiteLLM available",
			litellmAvailable:  true,
			standardAvailable: false,
			expectedModels:    2,
			expectFallback:    true,
			expectError:       false,
		},
		{
			name:              "Both endpoints unavailable",
			litellmAvailable:  false,
			standardAvailable: false,
			expectedModels:    0,
			expectFallback:    true,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用サーバーのセットアップ
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/model/info" {
					if tt.litellmAvailable {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{
							"models": [
								{"id": "gpt-4", "max_tokens": 8192, "mode": "chat", "input_cost": 0.00003},
								{"id": "claude-3-opus", "max_tokens": 200000, "mode": "chat", "input_cost": 0.000015}
							]
						}`))
					} else {
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte(`{"error": "Not found"}`))
					}
				} else if r.URL.Path == "/v1/models" {
					if tt.standardAvailable {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{
							"object": "list",
							"data": [
								{"id": "gpt-3.5-turbo", "object": "model", "created": 1234567890, "owned_by": "openai"}
							]
						}`))
					} else {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(`{"error": "Internal server error"}`))
					}
				}
			}))
			defer server.Close()

			// クライアントの作成
			cfg := config.New(server.URL, "test-api-key", 5*time.Second)
			client := api.NewClient(cfg)

			// API呼び出し
			response, err := client.FetchModelsWithFallback()

			// エラーチェック
			if (err != nil) != tt.expectError {
				t.Errorf("FetchModelsWithFallback() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// 成功時のレスポンスチェック
			if !tt.expectError {
				if response == nil {
					t.Error("FetchModelsWithFallback() returned nil response")
					return
				}

				if len(response.Models) != tt.expectedModels {
					t.Errorf("FetchModelsWithFallback() returned %d models, expected %d", len(response.Models), tt.expectedModels)
				}
			}
		})
	}
}

func TestEndpointPriorityIntegration(t *testing.T) {
	// OpenAI標準を先に試すが、両方が利用可能な場合、詳細情報を持つLiteLLMのレスポンスが返されることをテスト
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/model/info" {
			// LiteLLMエンドポイント
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"models": [
					{"id": "litellm-model", "max_tokens": 4096, "mode": "chat", "input_cost": 0.00001}
				]
			}`))
		} else if r.URL.Path == "/v1/models" {
			// OpenAI標準エンドポイント
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"object": "list",
				"data": [
					{"id": "standard-model", "object": "model", "created": 1234567890, "owned_by": "openai"}
				]
			}`))
		}
	}))
	defer server.Close()

	// クライアントの作成
	cfg := config.New(server.URL, "test-api-key", 5*time.Second)
	client := api.NewClient(cfg)

	// API呼び出し
	response, err := client.FetchModelsWithFallback()
	if err != nil {
		t.Fatalf("FetchModelsWithFallback() error = %v", err)
	}

	if response == nil {
		t.Fatal("FetchModelsWithFallback() returned nil response")
	}

	// LiteLLMのモデルが返されることを確認
	if len(response.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(response.Models))
	}

	if response.Models[0].ID != "litellm-model" {
		t.Errorf("Expected model ID 'litellm-model', got %s", response.Models[0].ID)
	}

	// LiteLLM特有のフィールドが設定されていることを確認
	if response.Models[0].MaxTokens != 4096 {
		t.Errorf("Expected MaxTokens 4096, got %d", response.Models[0].MaxTokens)
	}

	if response.Models[0].InputCost != 0.00001 {
		t.Errorf("Expected InputCost 0.00001, got %f", response.Models[0].InputCost)
	}
}
