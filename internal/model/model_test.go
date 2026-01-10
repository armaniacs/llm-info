package model

import (
	"testing"

	"github.com/your-org/llm-info/internal/api"
)

func TestFromAPIResponse(t *testing.T) {
	tests := []struct {
		name      string
		apiModels []api.ModelInfo
		expected  []Model
	}{
		{
			name: "single model",
			apiModels: []api.ModelInfo{
				{
					ID:        "gpt-4",
					MaxTokens: 8192,
					Mode:      "chat",
					InputCost: 0.00003,
				},
			},
			expected: []Model{
				{
					Name:      "gpt-4",
					MaxTokens: 8192,
					Mode:      "chat",
					InputCost: 0.00003,
				},
			},
		},
		{
			name: "multiple models",
			apiModels: []api.ModelInfo{
				{
					ID:        "gpt-4",
					MaxTokens: 8192,
					Mode:      "chat",
					InputCost: 0.00003,
				},
				{
					ID:        "claude-3-opus",
					MaxTokens: 200000,
					Mode:      "chat",
					InputCost: 0.000015,
				},
				{
					ID:        "gemini-1.5-pro",
					MaxTokens: 1000000,
					Mode:      "chat",
					InputCost: 0,
				},
			},
			expected: []Model{
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
				{
					Name:      "gemini-1.5-pro",
					MaxTokens: 1000000,
					Mode:      "chat",
					InputCost: 0,
				},
			},
		},
		{
			name:      "empty array",
			apiModels: []api.ModelInfo{},
			expected:  []Model{},
		},
		{
			name: "models with missing data",
			apiModels: []api.ModelInfo{
				{
					ID:        "model-with-zero-tokens",
					MaxTokens: 0,
					Mode:      "",
					InputCost: 0,
				},
				{
					ID:        "model-with-negative-cost",
					MaxTokens: 4096,
					Mode:      "completion",
					InputCost: -0.00001,
				},
			},
			expected: []Model{
				{
					Name:      "model-with-zero-tokens",
					MaxTokens: 0,
					Mode:      "",
					InputCost: 0,
				},
				{
					Name:      "model-with-negative-cost",
					MaxTokens: 4096,
					Mode:      "completion",
					InputCost: -0.00001,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromAPIResponse(tt.apiModels)

			if len(got) != len(tt.expected) {
				t.Errorf("FromAPIResponse() returned %d models, expected %d", len(got), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				if got[i].Name != expected.Name {
					t.Errorf("Model %d Name = %q, expected %q", i, got[i].Name, expected.Name)
				}
				if got[i].MaxTokens != expected.MaxTokens {
					t.Errorf("Model %d MaxTokens = %d, expected %d", i, got[i].MaxTokens, expected.MaxTokens)
				}
				if got[i].Mode != expected.Mode {
					t.Errorf("Model %d Mode = %q, expected %q", i, got[i].Mode, expected.Mode)
				}
				if got[i].InputCost != expected.InputCost {
					t.Errorf("Model %d InputCost = %f, expected %f", i, got[i].InputCost, expected.InputCost)
				}
			}
		})
	}
}

func TestFromAPIResponseNilInput(t *testing.T) {
	// nil入力のテスト
	got := FromAPIResponse(nil)

	if got != nil {
		t.Errorf("FromAPIResponse() with nil input should return nil, got %v", got)
	}
}
