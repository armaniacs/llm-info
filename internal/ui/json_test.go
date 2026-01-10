package ui

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/your-org/llm-info/internal/model"
)

func TestRenderJSON(t *testing.T) {
	tests := []struct {
		name     string
		models   []model.Model
		expected string
	}{
		{
			name: "single model",
			models: []model.Model{
				{
					Name:      "gpt-4",
					MaxTokens: 8192,
					Mode:      "chat",
					InputCost: 0.00003,
				},
			},
			expected: `[
  {
    "name": "gpt-4",
    "max_tokens": 8192,
    "mode": "chat",
    "input_cost": 0.00003
  }
]`,
		},
		{
			name: "multiple models",
			models: []model.Model{
				{
					Name:      "gpt-4",
					MaxTokens: 8192,
					Mode:      "chat",
					InputCost: 0.00003,
				},
				{
					Name:      "claude-3-opus",
					MaxTokens: 200000,
					Mode:      "chat",
					InputCost: 0.000015,
				},
			},
			expected: `[
  {
    "name": "gpt-4",
    "max_tokens": 8192,
    "mode": "chat",
    "input_cost": 0.00003
  },
  {
    "name": "claude-3-opus",
    "max_tokens": 200000,
    "mode": "chat",
    "input_cost": 0.000015
  }
]`,
		},
		{
			name:     "empty models",
			models:   []model.Model{},
			expected: "{}",
		},
		{
			name:     "nil models",
			models:   nil,
			expected: "{}",
		},
		{
			name: "model with missing fields",
			models: []model.Model{
				{
					Name:      "minimal-model",
					MaxTokens: 0,
					Mode:      "",
					InputCost: 0,
				},
			},
			expected: `[
  {
    "name": "minimal-model"
  }
]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 標準出力をキャプチャ
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// テスト対象関数の実行
			err := RenderJSON(tt.models)

			// 出力のキャプチャを終了
			w.Close()
			os.Stdout = oldStdout

			// 出力内容の読み取り
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// エラーチェック
			if err != nil {
				t.Errorf("RenderJSON() error = %v", err)
				return
			}

			// 出力の検証
			if strings.TrimSpace(output) != strings.TrimSpace(tt.expected) {
				t.Errorf("RenderJSON() output = %q, expected %q", strings.TrimSpace(output), strings.TrimSpace(tt.expected))
			}
		})
	}
}

func TestRenderCompactJSON(t *testing.T) {
	tests := []struct {
		name     string
		models   []model.Model
		expected string
	}{
		{
			name: "single model",
			models: []model.Model{
				{
					Name:      "gpt-4",
					MaxTokens: 8192,
					Mode:      "chat",
					InputCost: 0.00003,
				},
			},
			expected: `[{"name":"gpt-4","max_tokens":8192,"mode":"chat","input_cost":0.00003}]`,
		},
		{
			name:     "empty models",
			models:   []model.Model{},
			expected: "[]",
		},
		{
			name:     "nil models",
			models:   nil,
			expected: "[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 標準出力をキャプチャ
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// テスト対象関数の実行
			err := RenderCompactJSON(tt.models)

			// 出力のキャプチャを終了
			w.Close()
			os.Stdout = oldStdout

			// 出力内容の読み取り
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// エラーチェック
			if err != nil {
				t.Errorf("RenderCompactJSON() error = %v", err)
				return
			}

			// 出力の検証
			if strings.TrimSpace(output) != strings.TrimSpace(tt.expected) {
				t.Errorf("RenderCompactJSON() output = %q, expected %q", strings.TrimSpace(output), strings.TrimSpace(tt.expected))
			}
		})
	}
}

func TestRenderJSONWithMetadata(t *testing.T) {
	models := []model.Model{
		{
			Name:      "gpt-4",
			MaxTokens: 8192,
			Mode:      "chat",
			InputCost: 0.00003,
		},
	}

	metadata := map[string]interface{}{
		"total":  1,
		"source": "test",
	}

	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// テスト対象関数の実行
	err := RenderJSONWithMetadata(models, metadata)

	// 出力のキャプチャを終了
	w.Close()
	os.Stdout = oldStdout

	// 出力内容の読み取り
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// エラーチェック
	if err != nil {
		t.Errorf("RenderJSONWithMetadata() error = %v", err)
		return
	}

	// JSONのパースして検証
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
		return
	}

	// メタデータの検証
	if metadata, ok := result["metadata"].(map[string]interface{}); ok {
		if metadata["total"] != float64(1) {
			t.Errorf("Expected metadata.total = 1, got %v", metadata["total"])
		}
		if metadata["source"] != "test" {
			t.Errorf("Expected metadata.source = 'test', got %v", metadata["source"])
		}
	} else {
		t.Error("metadata field not found or not a map")
	}

	// モデルデータの検証
	if models, ok := result["models"].([]interface{}); ok {
		if len(models) != 1 {
			t.Errorf("Expected 1 model, got %d", len(models))
		}
	} else {
		t.Error("models field not found or not an array")
	}
}

func TestJSONModel(t *testing.T) {
	// JSONModel構造体のテスト
	model := JSONModel{
		Name:      "test-model",
		MaxTokens: 4096,
		Mode:      "chat",
		InputCost: 0.00001,
	}

	// JSONにシリアライズ
	data, err := json.Marshal(model)
	if err != nil {
		t.Errorf("Failed to marshal JSONModel: %v", err)
		return
	}

	// JSONからデシリアライズ
	var unmarshaled JSONModel
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal JSONModel: %v", err)
		return
	}

	// 検証
	if unmarshaled.Name != model.Name {
		t.Errorf("Expected Name %q, got %q", model.Name, unmarshaled.Name)
	}
	if unmarshaled.MaxTokens != model.MaxTokens {
		t.Errorf("Expected MaxTokens %d, got %d", model.MaxTokens, unmarshaled.MaxTokens)
	}
	if unmarshaled.Mode != model.Mode {
		t.Errorf("Expected Mode %q, got %q", model.Mode, unmarshaled.Mode)
	}
	if unmarshaled.InputCost != model.InputCost {
		t.Errorf("Expected InputCost %f, got %f", model.InputCost, unmarshaled.InputCost)
	}
}
