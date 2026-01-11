package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/armaniacs/llm-info/internal/probe"
)

func TestTableFormatter_FormatContextWindowResult(t *testing.T) {
	formatter := NewTableFormatter()

	result := &probe.ContextWindowResult{
		Model:              "GLM-4.6",
		MaxContextTokens:   127000,
		MethodConfidence:   "high",
		Trials:             12,
		Duration:           45*time.Second + 300*time.Millisecond,
		MaxInputAtSuccess:  126800,
		Success:            true,
		ErrorMessage:       "",
	}

	output := formatter.FormatContextWindowResult(result)

	// 期待される要素が含まれることを確認
	if !strings.Contains(output, "GLM-4.6") {
		t.Error("Output should contain model name")
	}
	if !strings.Contains(output, "127,000") {
		t.Error("Output should contain formatted context tokens")
	}
	if !strings.Contains(output, "✓ Success") {
		t.Error("Output should contain success status")
	}
	if !strings.Contains(output, "45.3s") {
		t.Error("Output should contain formatted duration")
	}
	if !strings.Contains(output, "126,800") {
		t.Error("Output should contain max input tokens")
	}
}

func TestTableFormatter_FormatContextWindowResult_Error(t *testing.T) {
	formatter := NewTableFormatter()

	result := &probe.ContextWindowResult{
		Model:              "invalid-model",
		MaxContextTokens:   0,
		MethodConfidence:   "low",
		Trials:             5,
		Duration:           30 * time.Second,
		MaxInputAtSuccess:  0,
		Success:            false,
		ErrorMessage:       "API timeout after 30s",
	}

	output := formatter.FormatContextWindowResult(result)

	if !strings.Contains(output, "invalid-model") {
		t.Error("Output should contain model name")
	}
	if !strings.Contains(output, "✗ Failed") {
		t.Error("Output should contain failed status")
	}
	if !strings.Contains(output, "API timeout after 30s") {
		t.Error("Output should contain error message")
	}
}

func TestTableFormatter_FormatMaxOutputResult(t *testing.T) {
	formatter := NewTableFormatter()

	result := &probe.MaxOutputResult{
		Model:                   "GLM-4.6",
		MaxOutputTokens:         16384,
		Evidence:                "validation_error",
		Trials:                  8,
		Duration:                30*time.Second + 200*time.Millisecond,
		MaxSuccessfullyGenerated: 8192,
		Success:                 true,
		ErrorMessage:            "",
	}

	output := formatter.FormatMaxOutputResult(result)

	if !strings.Contains(output, "GLM-4.6") {
		t.Error("Output should contain model name")
	}
	if !strings.Contains(output, "16,384") {
		t.Error("Output should contain formatted max output tokens")
	}
	if !strings.Contains(output, "validation_error") {
		t.Error("Output should contain evidence")
	}
	if !strings.Contains(output, "30.2s") {
		t.Error("Output should contain formatted duration")
	}
	if !strings.Contains(output, "8,192") {
		t.Error("Output should contain max successfully generated tokens")
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{123, "123"},
		{1234, "1,234"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
	}

	for _, tt := range tests {
		result := formatNumber(tt.input)
		if result != tt.expected {
			t.Errorf("formatNumber(%d) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{500 * time.Millisecond, "500ms"},
		{1 * time.Second, "1.0s"},
		{45*time.Second + 300*time.Millisecond, "45.3s"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.input)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestTableFormatter_FormatVerboseHistory(t *testing.T) {
	formatter := NewTableFormatter()

	trials := []probe.TrialInfo{
		{TokenCount: 1000, Success: true, Message: "Success"},
		{TokenCount: 10000, Success: false, Message: "Context length exceeded"},
		{TokenCount: 8000, Success: true, Message: "Success with smaller input"},
	}

	output := formatter.FormatVerboseHistory(trials)

	if !strings.Contains(output, "Search History") {
		t.Error("Output should contain search history header")
	}
	if !strings.Contains(output, "1,000") {
		t.Error("Output should contain formatted token count")
	}
	if !strings.Contains(output, "✓") {
		t.Error("Output should contain success marking")
	}
	if !strings.Contains(output, "✗") {
		t.Error("Output should contain failure marking")
	}
	if !strings.Contains(output, "Context length exceeded") {
		t.Error("Output should contain trial message")
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abc", "cba"},
		{"123456", "654321"},
		{"", ""},
	}

	for _, tt := range tests {
		result := reverse(tt.input)
		if result != tt.expected {
			t.Errorf("reverse(%s) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}