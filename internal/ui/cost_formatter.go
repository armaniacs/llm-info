package ui

import (
	"fmt"
	"strings"

	"github.com/armaniacs/llm-info/internal/cost"
)

// FormatDryRunCostEstimate はDry-run時のコスト概算を表示用にフォーマットする
func FormatDryRunCostEstimate(summary *cost.UsageSummary, calculator *cost.Calculator) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("Estimated API Usage:\n")
	sb.WriteString(strings.Repeat("─", 40) + "\n")

	// Context Window Probeの見積もり
	if contextUsage, ok := summary.BreakdownByProbe["context"]; ok {
		sb.WriteString(fmt.Sprintf("  Context Window Probe : ~%d trials, ~%s\n",
			contextUsage.Trials,
			calculator.FormatCost(contextUsage.Cost, 3)))
	}

	// Max Output Probeの見積もり
	if outputUsage, ok := summary.BreakdownByProbe["max_output"]; ok {
		sb.WriteString(fmt.Sprintf("  Max Output Probe     : ~%d trials, ~%s\n",
			outputUsage.Trials,
			calculator.FormatCost(outputUsage.Cost, 3)))
	}

	sb.WriteString(strings.Repeat("─", 40) + "\n")
	sb.WriteString(fmt.Sprintf("  Total Estimated Cost : ~%s\n",
		calculator.FormatCost(summary.TotalCost, 3)))

	// 不明モデルの警告
	if summary.UnknownModelUsed {
		sb.WriteString("\n")
		sb.WriteString("⚠️  Model pricing unknown. Using default rate.\n")
	}

	sb.WriteString("\n")

	return sb.String()
}

// FormatAPIUsageSummary は実行後のAPI使用量サマリーをテーブル形式でフォーマットする
func FormatAPIUsageSummary(summary *cost.UsageSummary, calculator *cost.Calculator) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("API Usage Summary\n")
	sb.WriteString(strings.Repeat("═", 40) + "\n")

	// 総計
	sb.WriteString(fmt.Sprintf("  Total Trials       : %d\n", getTotalTrials(summary)))
	sb.WriteString(fmt.Sprintf("  Input Tokens       : %s\n",
		calculator.FormatTokenCount(summary.TotalInputTokens)))
	sb.WriteString(fmt.Sprintf("  Output Tokens      : %s\n",
		calculator.FormatTokenCount(summary.TotalOutputTokens)))
	sb.WriteString(fmt.Sprintf("  Total Cost         : %s\n",
		calculator.FormatCost(summary.TotalCost, 4)))

	// 内訳
	if len(summary.BreakdownByProbe) > 1 {
		sb.WriteString("\n")
		sb.WriteString("  Breakdown:\n")

		if contextUsage, ok := summary.BreakdownByProbe["context"]; ok {
			sb.WriteString(fmt.Sprintf("    Context Probe    : %s (%d trials)\n",
				calculator.FormatCost(contextUsage.Cost, 4),
				contextUsage.Trials))
		}

		if outputUsage, ok := summary.BreakdownByProbe["max_output"]; ok {
			sb.WriteString(fmt.Sprintf("    Max Output Probe : %s (%d trials)\n",
				calculator.FormatCost(outputUsage.Cost, 4),
				outputUsage.Trials))
		}
	}

	// 不明モデルの警告
	if summary.UnknownModelUsed {
		sb.WriteString("\n")
		sb.WriteString("  ⚠️  Using default pricing (model not in config)\n")
	}

	sb.WriteString(strings.Repeat("═", 40) + "\n")

	return sb.String()
}

// AddCostToJSON はJSON出力にコスト情報を追加する
func AddCostToJSON(result map[string]interface{}, summary *cost.UsageSummary) map[string]interface{} {
	costData := map[string]interface{}{
		"total":             summary.TotalCost,
		"input_tokens":      summary.TotalInputTokens,
		"output_tokens":     summary.TotalOutputTokens,
		"total_tokens":      summary.TotalTokens,
		"warning_triggered": summary.WarningTriggered,
		"unknown_model":     summary.UnknownModelUsed,
	}

	// 内訳情報
	if len(summary.BreakdownByProbe) > 0 {
		breakdown := make(map[string]interface{})
		for probeType, usage := range summary.BreakdownByProbe {
			breakdown[probeType] = map[string]interface{}{
				"cost":          usage.Cost,
				"input_tokens":  usage.InputTokens,
				"output_tokens": usage.OutputTokens,
				"trials":        usage.Trials,
			}
		}
		costData["breakdown"] = breakdown
	}

	result["cost"] = costData
	return result
}

// getTotalTrials は総試行回数を計算する
func getTotalTrials(summary *cost.UsageSummary) int {
	total := 0
	for _, usage := range summary.BreakdownByProbe {
		total += usage.Trials
	}
	return total
}
