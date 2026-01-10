package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/llm-info/internal/config"
)

func TestClient_GetModelInfo(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		mockResponse   string
		statusCode     int
		expectedModels []ModelInfo
		wantErr        bool
	}{
		{
			name:   "successful response without API key",
			apiKey: "",
			mockResponse: `{
				"models": [
					{
						"id": "gpt-4",
						"max_tokens": 8192,
						"mode": "chat",
						"input_cost": 0.00003
					},
					{
						"id": "claude-3-opus",
						"max_tokens": 200000,
						"mode": "chat",
						"input_cost": 0.000015
					}
				]
			}`,
			statusCode: http.StatusOK,
			expectedModels: []ModelInfo{
				{ID: "gpt-4", MaxTokens: 8192, Mode: "chat", InputCost: 0.00003},
				{ID: "claude-3-opus", MaxTokens: 200000, Mode: "chat", InputCost: 0.000015},
			},
			wantErr: false,
		},
		{
			name:   "successful response with API key",
			apiKey: "test-api-key",
			mockResponse: `{
				"models": [
					{
						"id": "gpt-3.5-turbo",
						"max_tokens": 4096,
						"mode": "chat",
						"input_cost": 0.000002
					}
				]
			}`,
			statusCode: http.StatusOK,
			expectedModels: []ModelInfo{
				{ID: "gpt-3.5-turbo", MaxTokens: 4096, Mode: "chat", InputCost: 0.000002},
			},
			wantErr: false,
		},
		{
			name:           "API error response",
			apiKey:         "",
			mockResponse:   `{"error": "Internal server error"}`,
			statusCode:     http.StatusInternalServerError,
			expectedModels: nil,
			wantErr:        true,
		},
		{
			name:           "invalid JSON response",
			apiKey:         "",
			mockResponse:   `{"invalid": json}`,
			statusCode:     http.StatusOK,
			expectedModels: nil,
			wantErr:        true,
		},
		{
			name:           "empty models array",
			apiKey:         "",
			mockResponse:   `{"models": []}`,
			statusCode:     http.StatusOK,
			expectedModels: []ModelInfo{},
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用サーバーのセットアップ
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// APIキーのチェック
				if tt.apiKey != "" {
					authHeader := r.Header.Get("Authorization")
					expectedAuth := "Bearer " + tt.apiKey
					if authHeader != expectedAuth {
						t.Errorf("Expected Authorization header %q, got %q", expectedAuth, authHeader)
					}
				}

				// Content-Typeヘッダーのチェック
				expectedContentType := "application/json"
				if contentType := r.Header.Get("Content-Type"); contentType != expectedContentType {
					t.Errorf("Expected Content-Type header %q, got %q", expectedContentType, contentType)
				}

				// レスポンスの設定
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// クライアントの作成
			cfg := config.New(server.URL, tt.apiKey, 5*time.Second)
			client := NewClient(cfg)

			// API呼び出し
			got, err := client.GetModelInfo()

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("GetModelInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 成功時のレスポンスチェック
			if !tt.wantErr {
				if got == nil {
					t.Error("GetModelInfo() returned nil response")
					return
				}

				if len(got.Models) != len(tt.expectedModels) {
					t.Errorf("GetModelInfo() returned %d models, expected %d", len(got.Models), len(tt.expectedModels))
					return
				}

				for i, expected := range tt.expectedModels {
					if got.Models[i].ID != expected.ID {
						t.Errorf("Model %d ID = %q, expected %q", i, got.Models[i].ID, expected.ID)
					}
					if got.Models[i].MaxTokens != expected.MaxTokens {
						t.Errorf("Model %d MaxTokens = %d, expected %d", i, got.Models[i].MaxTokens, expected.MaxTokens)
					}
					if got.Models[i].Mode != expected.Mode {
						t.Errorf("Model %d Mode = %q, expected %q", i, got.Models[i].Mode, expected.Mode)
					}
					if got.Models[i].InputCost != expected.InputCost {
						t.Errorf("Model %d InputCost = %f, expected %f", i, got.Models[i].InputCost, expected.InputCost)
					}
				}
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	baseURL := "https://api.example.com"
	apiKey := "test-key"
	timeout := 10 * time.Second

	cfg := config.New(baseURL, apiKey, timeout)
	client := NewClient(cfg)

	if client.baseURL != baseURL {
		t.Errorf("NewClient() baseURL = %q, expected %q", client.baseURL, baseURL)
	}

	if client.apiKey != apiKey {
		t.Errorf("NewClient() apiKey = %q, expected %q", client.apiKey, apiKey)
	}

	if client.timeout != timeout {
		t.Errorf("NewClient() timeout = %v, expected %v", client.timeout, timeout)
	}

	if client.client == nil {
		t.Error("NewClient() client is nil")
	}

	if client.client.Timeout != timeout {
		t.Errorf("NewClient() client timeout = %v, expected %v", client.client.Timeout, timeout)
	}
}

func TestClient_FetchModelsWithFallback(t *testing.T) {
	tests := []struct {
		name               string
		litellmStatusCode  int
		litellmResponse    string
		standardStatusCode int
		standardResponse   string
		expectedModels     []ModelInfo
		wantErr            bool
		expectFallback     bool
	}{
		{
			name:               "LiteLLM succeeds, no fallback",
			litellmStatusCode:  http.StatusOK,
			litellmResponse:    `{"models": [{"id": "gpt-4", "max_tokens": 8192, "mode": "chat", "input_cost": 0.00003}]}`,
			standardStatusCode: http.StatusOK,
			standardResponse:   `{"object": "list", "data": []}`,
			expectedModels:     []ModelInfo{{ID: "gpt-4", MaxTokens: 8192, Mode: "chat", InputCost: 0.00003}},
			wantErr:            false,
			expectFallback:     false,
		},
		{
			name:               "LiteLLM fails, fallback to standard succeeds",
			litellmStatusCode:  http.StatusNotFound,
			litellmResponse:    `{"error": "Not found"}`,
			standardStatusCode: http.StatusOK,
			standardResponse:   `{"object": "list", "data": [{"id": "gpt-3.5-turbo", "object": "model", "created": 1234567890, "owned_by": "openai"}]}`,
			expectedModels:     []ModelInfo{{ID: "gpt-3.5-turbo", MaxTokens: 0, Mode: "chat", InputCost: 0}},
			wantErr:            false,
			expectFallback:     true,
		},
		{
			name:               "Both endpoints fail",
			litellmStatusCode:  http.StatusNotFound,
			litellmResponse:    `{"error": "Not found"}`,
			standardStatusCode: http.StatusInternalServerError,
			standardResponse:   `{"error": "Internal server error"}`,
			expectedModels:     nil,
			wantErr:            true,
			expectFallback:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用サーバーのセットアップ
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/model/info" {
					// LiteLLMエンドポイント
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.litellmStatusCode)
					w.Write([]byte(tt.litellmResponse))
				} else if r.URL.Path == "/v1/models" {
					// OpenAI標準エンドポイント
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.standardStatusCode)
					w.Write([]byte(tt.standardResponse))
				}
			}))
			defer server.Close()

			// クライアントの作成
			cfg := config.New(server.URL, "test-api-key", 5*time.Second)
			client := NewClient(cfg)

			// API呼び出し
			got, err := client.FetchModelsWithFallback()

			// エラーチェック
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchModelsWithFallback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 成功時のレスポンスチェック
			if !tt.wantErr {
				if got == nil {
					t.Error("FetchModelsWithFallback() returned nil response")
					return
				}

				if len(got.Models) != len(tt.expectedModels) {
					t.Errorf("FetchModelsWithFallback() returned %d models, expected %d", len(got.Models), len(tt.expectedModels))
					return
				}

				for i, expected := range tt.expectedModels {
					if got.Models[i].ID != expected.ID {
						t.Errorf("Model %d ID = %q, expected %q", i, got.Models[i].ID, expected.ID)
					}
					if got.Models[i].MaxTokens != expected.MaxTokens {
						t.Errorf("Model %d MaxTokens = %d, expected %d", i, got.Models[i].MaxTokens, expected.MaxTokens)
					}
					if got.Models[i].Mode != expected.Mode {
						t.Errorf("Model %d Mode = %q, expected %q", i, got.Models[i].Mode, expected.Mode)
					}
					if got.Models[i].InputCost != expected.InputCost {
						t.Errorf("Model %d InputCost = %f, expected %f", i, got.Models[i].InputCost, expected.InputCost)
					}
				}
			}
		})
	}
}
