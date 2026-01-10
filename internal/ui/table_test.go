package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/armaniacs/llm-info/internal/model"
)

func TestRenderTable(t *testing.T) {
	tests := []struct {
		name     string
		models   []model.Model
		expected []string
	}{
		{
			name: "single model with all fields",
			models: []model.Model{
				{
					Name:      "gpt-4",
					MaxTokens: 8192,
					Mode:      "chat",
					InputCost: 0.00003,
				},
			},
			expected: []string{
				"Model Name  Max Tokens  Mode  Input Cost",
				"----------  ----------  ----  ----------",
				"gpt-4       8192        chat  0.000030",
			},
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
			expected: []string{
				"Model Name  Max Tokens  Mode  Input Cost",
				"----------  ----------  ----  ----------",
				"gpt-4       8192        chat  0.000030",
				"claude-3-opus  200000      chat  0.000015",
			},
		},
		{
			name: "models with missing max tokens",
			models: []model.Model{
				{
					Name:      "model-no-tokens",
					MaxTokens: 0,
					Mode:      "chat",
					InputCost: 0.00001,
				},
				{
					Name:      "model-with-tokens",
					MaxTokens: 4096,
					Mode:      "chat",
					InputCost: 0.00002,
				},
			},
			expected: []string{
				"Model Name  Max Tokens  Mode  Input Cost",
				"----------  ----------  ----  ----------",
				"model-no-tokens  0           chat  0.000010",
				"model-with-tokens  4096        chat  0.000020",
			},
		},
		{
			name: "models with missing mode",
			models: []model.Model{
				{
					Name:      "model-no-mode",
					MaxTokens: 8192,
					Mode:      "",
					InputCost: 0.00001,
				},
				{
					Name:      "model-with-mode",
					MaxTokens: 4096,
					Mode:      "completion",
					InputCost: 0.00002,
				},
			},
			expected: []string{
				"Model Name  Max Tokens  Mode        Input Cost",
				"----------  ----------  ----------  ----------",
				"model-no-mode  8192                    0.000010",
				"model-with-mode  4096        completion  0.000020",
			},
		},
		{
			name: "models with missing input cost",
			models: []model.Model{
				{
					Name:      "model-no-cost",
					MaxTokens: 8192,
					Mode:      "chat",
					InputCost: 0,
				},
				{
					Name:      "model-with-cost",
					MaxTokens: 4096,
					Mode:      "chat",
					InputCost: 0.00001,
				},
			},
			expected: []string{
				"Model Name  Max Tokens  Mode  Input Cost",
				"----------  ----------  ----  ----------",
				"model-no-cost  8192        chat  0.000000",
				"model-with-cost  4096        chat  0.000010",
			},
		},
		{
			name: "models with only name",
			models: []model.Model{
				{
					Name:      "minimal-model",
					MaxTokens: 0,
					Mode:      "",
					InputCost: 0,
				},
			},
			expected: []string{
				"Model Name",
				"----------",
				"minimal-model",
			},
		},
		{
			name:     "empty models array",
			models:   []model.Model{},
			expected: []string{"No models found."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 標準出力をキャプチャ
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// テスト対象関数の実行
			RenderTable(tt.models)

			// 出力のキャプチャを終了
			w.Close()
			os.Stdout = oldStdout

			// 出力内容の読み取り
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// 出力の検証
			outputLines := strings.Split(strings.TrimSpace(output), "\n")

			if len(outputLines) != len(tt.expected) {
				t.Errorf("RenderTable() output lines = %d, expected %d", len(outputLines), len(tt.expected))
				t.Logf("Actual output:\n%s", output)
				t.Logf("Expected output:\n%s", strings.Join(tt.expected, "\n"))
				return
			}

			for i, expectedLine := range tt.expected {
				if !strings.Contains(outputLines[i], expectedLine) {
					t.Errorf("RenderTable() line %d = %q, expected to contain %q", i, outputLines[i], expectedLine)
				}
			}
		})
	}
}

func TestRenderTableNilInput(t *testing.T) {
	// nil入力のテスト
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	RenderTable(nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expected := "No models found."
	if strings.TrimSpace(output) != expected {
		t.Errorf("RenderTable() with nil input = %q, expected %q", strings.TrimSpace(output), expected)
	}
}

func TestPrintTable(t *testing.T) {
	headers := []string{"Name", "Age"}
	rows := [][]string{
		{"Alice", "30"},
		{"Bob", "25"},
	}
	colWidths := []int{5, 3}

	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printTable(headers, rows, colWidths)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expectedLines := []string{
		"Name   Age",
		"-----  ---",
		"Alice  30",
		"Bob    25",
	}

	outputLines := strings.Split(strings.TrimSpace(output), "\n")

	for i, expectedLine := range expectedLines {
		if i >= len(outputLines) {
			t.Errorf("Expected line %d not found in output", i)
			continue
		}
		if !strings.Contains(outputLines[i], expectedLine) {
			t.Errorf("printTable() line %d = %q, expected to contain %q", i, outputLines[i], expectedLine)
		}
	}
}

func TestPrintRow(t *testing.T) {
	row := []string{"Alice", "30"}
	colWidths := []int{5, 3}

	// 標準出力をキャプチャ
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printRow(row, colWidths)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	expected := "Alice  30 \n"
	if output != expected {
		t.Errorf("printRow() = %q, expected %q", output, expected)
	}
}
