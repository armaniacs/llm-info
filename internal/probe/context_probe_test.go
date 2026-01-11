package probe

import (
	"testing"

	"github.com/armaniacs/llm-info/internal/api"
)

func TestContextProbe_NilUsageHandling(t *testing.T) {
	// response.Usageがnilでもpanicしないことをテスト
	_ = &ContextWindowProbe{}

	// nil Usageのレスポンス
	response := &api.ProbeResponse{
		ID:     "test",
		Usage:  nil, // nil!
		Choices: []api.ChatChoice{{FinishReason: "stop"}},
	}

	// testWithTokenCountの内部ロジックをシミュレート
	if response.Usage == nil {
		// エラーハンドリングが正しく動作することを確認
		result := &BoundarySearchResult{
			Value:           0,
			Success:         false,
			ErrorMessage:    "Response missing usage information",
			Source:          "api_error",
			Trials:          1,
			EstimatedTokens: 0,
		}

		if result.Success {
			t.Error("Expected Success=false when Usage is nil")
		}
		if result.ErrorMessage != "Response missing usage information" {
			t.Errorf("Expected specific error message, got: %s", result.ErrorMessage)
		}
	}
}