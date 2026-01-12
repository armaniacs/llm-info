package cost

import (
	"testing"

	"github.com/armaniacs/llm-info/pkg/config"
)

func TestNewCalculator(t *testing.T) {
	// テストケース1: nilのコスト設定で作成
	calculator := NewCalculator(nil, "test-model")
	if calculator == nil {
		t.Error("Expected calculator to be created, got nil")
	}
	if calculator.warningThreshold != 0.05 {
		t.Errorf("Expected default threshold 0.05, got %f", calculator.warningThreshold)
	}

	// テストケース2: 有効なコスト設定で作成
	costConfig := &config.CostConfig{
		WarningThreshold: 0.1,
		Pricing: map[string]config.Pricing{
			"gpt-4": {
				InputPricePer1K:  0.03,
				OutputPricePer1K: 0.06,
			},
		},
	}
	calculator = NewCalculator(costConfig, "gpt-4")
	if calculator.warningThreshold != 0.1 {
		t.Errorf("Expected threshold 0.1, got %f", calculator.warningThreshold)
	}
}

func TestCalculateTrialCost(t *testing.T) {
	costConfig := &config.CostConfig{
		Pricing: map[string]config.Pricing{
			"gpt-4": {
				InputPricePer1K:  0.03,
				OutputPricePer1K: 0.06,
			},
		},
	}
	calculator := NewCalculator(costConfig, "gpt-4")

	// テストケース1: 既知のモデル
	cost := calculator.CalculateTrialCost(1000, 500, "gpt-4")
	expected := 0.03 + 0.06*0.5 // 0.03 + 0.03 = 0.06
	if cost != expected {
		t.Errorf("Expected cost %f, got %f", expected, cost)
	}

	// テストケース2: 未知のモデル（デフォルト料金を使用）
	cost = calculator.CalculateTrialCost(1000, 500, "unknown-model")
	expected = 0.00015 + 0.0006*0.5 // 0.00015 + 0.0003 = 0.00045
	if cost != expected {
		t.Errorf("Expected default cost %f, got %f", expected, cost)
	}
}

func TestEstimateProbeCost(t *testing.T) {
	costConfig := &config.CostConfig{
		Pricing: map[string]config.Pricing{
			"gpt-4": {
				InputPricePer1K:  0.03,
				OutputPricePer1K: 0.06,
			},
		},
	}
	calculator := NewCalculator(costConfig, "gpt-4")

	// テストケース1: Context probeの見積もり
	cost, inputTokens, outputTokens := calculator.EstimateProbeCost("gpt-4", "context")
	if cost <= 0 {
		t.Errorf("Expected positive cost, got %f", cost)
	}
	if inputTokens <= 0 || outputTokens <= 0 {
		t.Errorf("Expected positive token counts, got input=%d, output=%d", inputTokens, outputTokens)
	}

	// テストケース2: Max output probeの見積もり
	cost, inputTokens, outputTokens = calculator.EstimateProbeCost("gpt-4", "max_output")
	if cost <= 0 {
		t.Errorf("Expected positive cost, got %f", cost)
	}
	if inputTokens <= 0 || outputTokens <= 0 {
		t.Errorf("Expected positive token counts, got input=%d, output=%d", inputTokens, outputTokens)
	}

	// テストケース3: 未知のモデル（デフォルト料金を使用）
	cost, inputTokens, outputTokens = calculator.EstimateProbeCost("unknown-model", "context")
	if cost <= 0 {
		t.Errorf("Expected positive cost for unknown model, got %f", cost)
	}
}

func TestFormatCost(t *testing.T) {
	calculator := NewCalculator(nil, "test-model")

	// テストケース: コストのフォーマット
	cost := 0.123456
	formatted := calculator.FormatCost(cost, 3)
	expected := "$0.123"
	if formatted != expected {
		t.Errorf("Expected %s, got %s", expected, formatted)
	}
}

func TestFormatTokenCount(t *testing.T) {
	calculator := NewCalculator(nil, "test-model")

	// テストケース1: 小さな数値
	count := 1234
	formatted := calculator.FormatTokenCount(count)
	expected := "1,234"
	if formatted != expected {
		t.Errorf("Expected %s, got %s", expected, formatted)
	}

	// テストケース2: 大きな数値
	count = 1234567
	formatted = calculator.FormatTokenCount(count)
	expected = "1,234,567"
	if formatted != expected {
		t.Errorf("Expected %s, got %s", expected, formatted)
	}
}
