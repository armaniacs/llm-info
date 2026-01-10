package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/llm-info/internal/config"
)

func TestFetchStandardModels_Success(t *testing.T) {
	// テスト用のモックサーバーを作成
	mockResponse := StandardResponse{
		Object: "list",
		Data: []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		}{
			{
				ID:      "gpt-4",
				Object:  "model",
				Created: 1234567890,
				OwnedBy: "openai",
			},
			{
				ID:      "gpt-3.5-turbo",
				Object:  "model",
				Created: 1234567891,
				OwnedBy: "openai",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Errorf("Expected path /v1/models, got %s", r.URL.Path)
		}

		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header 'Bearer test-api-key', got %s", authHeader)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// テスト用クライアントを作成
	cfg := &config.Config{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 10 * time.Second,
	}
	client := NewClient(cfg)

	// テスト実行
	response, err := client.FetchStandardModels()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.Object != "list" {
		t.Errorf("Expected object 'list', got %s", response.Object)
	}

	if len(response.Data) != 2 {
		t.Errorf("Expected 2 models, got %d", len(response.Data))
	}

	if response.Data[0].ID != "gpt-4" {
		t.Errorf("Expected first model ID 'gpt-4', got %s", response.Data[0].ID)
	}
}

func TestFetchStandardModels_Error(t *testing.T) {
	// エラーを返すモックサーバーを作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	// テスト用クライアントを作成
	cfg := &config.Config{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 10 * time.Second,
	}
	client := NewClient(cfg)

	// テスト実行
	_, err := client.FetchStandardModels()
	if err == nil {
		t.Error("Expected error, got nil")
	}

	expectedError := "API request failed with status 404"
	if !contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got %s", expectedError, err.Error())
	}
}

func TestConvertStandardResponse(t *testing.T) {
	// テスト用クライアントを作成
	cfg := &config.Config{
		BaseURL: "https://api.example.com",
		APIKey:  "",
		Timeout: 10 * time.Second,
	}
	client := NewClient(cfg)

	// テスト用レスポンスを作成
	standardResp := &StandardResponse{
		Object: "list",
		Data: []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		}{
			{
				ID:      "gpt-4",
				Object:  "model",
				Created: 1234567890,
				OwnedBy: "openai",
			},
		},
	}

	// テスト実行
	response := client.convertStandardResponse(standardResp)

	if len(response.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(response.Models))
	}

	model := response.Models[0]
	if model.ID != "gpt-4" {
		t.Errorf("Expected model ID 'gpt-4', got %s", model.ID)
	}

	if model.MaxTokens != 0 {
		t.Errorf("Expected MaxTokens 0, got %d", model.MaxTokens)
	}

	if model.Mode != "chat" {
		t.Errorf("Expected Mode 'chat', got %s", model.Mode)
	}

	if model.InputCost != 0 {
		t.Errorf("Expected InputCost 0, got %f", model.InputCost)
	}
}

// contains は文字列が部分文字列を含むかどうかをチェックするヘルパー関数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
