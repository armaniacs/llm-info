package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/armaniacs/llm-info/pkg/config"
)

func TestProbeClient_TimeoutWorks(t *testing.T) {
	// タイムアウトが実際に機能することをテスト
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // 遅いAPIをシミュレート
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.AppConfig{
		BaseURL: server.URL,
		APIKey:  "test",
		Timeout: 100 * time.Millisecond, // 短いタイムアウト
	}

	client := NewProbeClient(cfg)
	_, err := client.ProbeModel("test-model")

	// タイムアウトエラーが発生することを確認
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "deadline exceeded") &&
		!strings.Contains(err.Error(), "context deadline") {
		t.Errorf("Expected deadline exceeded error, got: %v", err)
	}
}

func TestProbeClient_NormalRequestStillWorks(t *testing.T) {
	// 正常なリクエストが引き続き動作することをテスト
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ProbeResponse{
			ID:      "test-123",
			Model:   "test-model",
			Choices: []ChatChoice{{FinishReason: "stop"}},
			Usage: &UsageInfo{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &config.AppConfig{
		BaseURL: server.URL,
		APIKey:  "test",
		Timeout: 10 * time.Second,
	}

	client := NewProbeClient(cfg)
	result, err := client.ProbeModel("test-model")

	// 正常に完了することを確認
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Error("Expected result, got nil")
	}
	if result.Usage == nil {
		t.Error("Expected Usage field, got nil")
	}
	if result.Usage.PromptTokens != 10 {
		t.Errorf("Expected PromptTokens=10, got %d", result.Usage.PromptTokens)
	}
}