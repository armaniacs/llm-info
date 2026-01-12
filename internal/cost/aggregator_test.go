package cost

import (
	"testing"

	"github.com/armaniacs/llm-info/pkg/config"
)

func TestAggregateFromTrials(t *testing.T) {
	// テスト用のコスト設定
	costConfig := &config.CostConfig{
		WarningThreshold: 0.05,
		Pricing: map[string]config.Pricing{
			"gpt-4": {
				InputPricePer1K:  0.03,
				OutputPricePer1K: 0.06,
			},
		},
	}
	calculator := NewCalculator(costConfig, "gpt-4")

	// テストケース1: 正常な試行データ
	contextTrials := []TrialUsage{
		{PromptTokens: 1000, CompletionTokens: 100},
		{PromptTokens: 2000, CompletionTokens: 200},
	}
	outputTrials := []TrialUsage{
		{PromptTokens: 500, CompletionTokens: 500},
		{PromptTokens: 500, CompletionTokens: 1000},
	}

	summary := AggregateFromTrials(calculator, "gpt-4", contextTrials, outputTrials)

	if summary == nil {
		t.Fatal("Expected summary to be created, got nil")
	}

	if summary.TotalCost <= 0 {
		t.Errorf("Expected positive total cost, got %f", summary.TotalCost)
	}

	if summary.TotalInputTokens != 4500 { // 1000 + 2000 + 500 + 500
		t.Errorf("Expected total input tokens 4500, got %d", summary.TotalInputTokens)
	}

	if summary.TotalOutputTokens != 1800 { // 100 + 200 + 500 + 1000
		t.Errorf("Expected total output tokens 1800, got %d", summary.TotalOutputTokens)
	}

	if len(summary.BreakdownByProbe) != 2 {
		t.Errorf("Expected 2 probe types in breakdown, got %d", len(summary.BreakdownByProbe))
	}
}

func TestEstimateUsage(t *testing.T) {
	// テスト用のコスト設定
	costConfig := &config.CostConfig{
		WarningThreshold: 0.05,
		Pricing: map[string]config.Pricing{
			"gpt-4": {
				InputPricePer1K:  0.03,
				OutputPricePer1K: 0.06,
			},
		},
	}
	calculator := NewCalculator(costConfig, "gpt-4")

	// テストケース1: 両方のprobeを含む見積もり
	summary := EstimateUsage(calculator, "gpt-4", false, false)

	if summary == nil {
		t.Fatal("Expected summary to be created, got nil")
	}

	if summary.TotalCost <= 0 {
		t.Errorf("Expected positive total cost, got %f", summary.TotalCost)
	}

	if len(summary.BreakdownByProbe) != 2 {
		t.Errorf("Expected 2 probe types in breakdown, got %d", len(summary.BreakdownByProbe))
	}

	// テストケース2: Context onlyの見積もり
	summary = EstimateUsage(calculator, "gpt-4", false, true)

	if len(summary.BreakdownByProbe) != 1 {
		t.Errorf("Expected 1 probe type in breakdown for context only, got %d", len(summary.BreakdownByProbe))
	}

	if _, exists := summary.BreakdownByProbe["context"]; !exists {
		t.Error("Expected context probe in breakdown")
	}

	// テストケース3: Output onlyの見積もり
	summary = EstimateUsage(calculator, "gpt-4", true, false)

	if len(summary.BreakdownByProbe) != 1 {
		t.Errorf("Expected 1 probe type in breakdown for output only, got %d", len(summary.BreakdownByProbe))
	}

	if _, exists := summary.BreakdownByProbe["max_output"]; !exists {
		t.Error("Expected max_output probe in breakdown")
	}
}

func TestAggregateProbeTrials(t *testing.T) {
	// テスト用のコスト設定
	costConfig := &config.CostConfig{
		Pricing: map[string]config.Pricing{
			"gpt-4": {
				InputPricePer1K:  0.03,
				OutputPricePer1K: 0.06,
			},
		},
	}
	calculator := NewCalculator(costConfig, "gpt-4")

	// テストケース: 正常な試行データ
	trials := []TrialUsage{
		{PromptTokens: 1000, CompletionTokens: 100},
		{PromptTokens: 2000, CompletionTokens: 200},
		{PromptTokens: 0, CompletionTokens: 0}, // 無効な試行（スキップされるべき）
	}

	usage := aggregateProbeTrials(calculator, "gpt-4", trials)

	if usage == nil {
		t.Fatal("Expected usage to be created, got nil")
	}

	if usage.Trials != 3 { // 全試行数（無効な試行も含む）
		t.Errorf("Expected 3 trials, got %d", usage.Trials)
	}

	if usage.InputTokens != 3000 { // 1000 + 2000
		t.Errorf("Expected input tokens 3000, got %d", usage.InputTokens)
	}

	if usage.OutputTokens != 300 { // 100 + 200
		t.Errorf("Expected output tokens 300, got %d", usage.OutputTokens)
	}

	if usage.Cost <= 0 {
		t.Errorf("Expected positive cost, got %f", usage.Cost)
	}
}
