package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/armaniacs/llm-info/internal/probe"
)

// TableFormatter はテーブル形式の出力を整形する
type TableFormatter struct {
	width int // テーブル幅（デフォルト80）
}

// NewTableFormatter は新しいTableFormatterを作成
func NewTableFormatter() *TableFormatter {
	return &TableFormatter{
		width: 80,
	}
}

// FormatIntegratedResult は統合探索結果を整形する
func (tf *TableFormatter) FormatIntegratedResult(model string, contextResult *probe.ContextWindowResult, outputResult *probe.MaxOutputResult, duration time.Duration, trials int) string {
	var sb strings.Builder

	// ヘッダー
	sb.WriteString("Model Constraints Probe Results\n")
	sb.WriteString(strings.Repeat("=", 31) + "\n")

	// データ行
	sb.WriteString(fmt.Sprintf("%-22s %s\n", "Model:", model))

	if contextResult != nil {
		sb.WriteString(fmt.Sprintf("%-22s %s tokens\n", "Context Window:", formatNumber(contextResult.MaxContextTokens)))
		sb.WriteString(fmt.Sprintf("%-22s %s\n", "Context Confidence:", contextResult.MethodConfidence))
	} else {
		sb.WriteString(fmt.Sprintf("%-22s %s\n", "Context Window:", "Failed"))
		sb.WriteString(fmt.Sprintf("%-22s %s\n", "Context Confidence:", "-"))
	}

	if outputResult != nil {
		sb.WriteString(fmt.Sprintf("%-22s %s tokens\n", "Max Output Tokens:", formatNumber(outputResult.MaxOutputTokens)))
		sb.WriteString(fmt.Sprintf("%-22s %s\n", "Output Confidence:", outputResult.MethodConfidence))
	} else {
		sb.WriteString(fmt.Sprintf("%-22s %s\n", "Max Output Tokens:", "Failed"))
		sb.WriteString(fmt.Sprintf("%-22s %s\n", "Output Confidence:", "-"))
	}

	sb.WriteString(fmt.Sprintf("%-22s %d\n", "Total Trials:", trials))
	sb.WriteString(fmt.Sprintf("%-22s %s\n", "Total Duration:", formatDuration(duration)))

	sb.WriteString("\n")

	// ステータス
	successCount := 0
	if contextResult != nil && contextResult.Success {
		successCount++
	}
	if outputResult != nil && outputResult.Success {
		successCount++
	}

	if successCount == 2 {
		sb.WriteString("Status: ✓ All probes succeeded\n")
	} else if successCount == 1 {
		sb.WriteString("Status: ⚠ Partial success\n")
	} else {
		sb.WriteString("Status: ✗ All probes failed\n")
	}

	return sb.String()
}

// FormatContextWindowResult はContext Window探索結果を整形
func (tf *TableFormatter) FormatContextWindowResult(result *probe.ContextWindowResult) string {
	var sb strings.Builder

	// ヘッダー
	sb.WriteString("Context Window Probe Results\n")
	sb.WriteString(strings.Repeat("=", 28) + "\n")

	// データ行
	sb.WriteString(fmt.Sprintf("%-22s %s\n", "Model:", result.Model))
	sb.WriteString(fmt.Sprintf("%-22s %s tokens\n", "Estimated Context:", formatNumber(result.MaxContextTokens)))
	sb.WriteString(fmt.Sprintf("%-22s %s\n", "Method Confidence:", result.MethodConfidence))
	sb.WriteString(fmt.Sprintf("%-22s %d\n", "Trials:", result.Trials))
	sb.WriteString(fmt.Sprintf("%-22s %s\n", "Duration:", formatDuration(result.Duration)))

	if result.MaxInputAtSuccess > 0 {
		sb.WriteString(fmt.Sprintf("%-22s %s tokens\n", "Max Input at Success:", formatNumber(result.MaxInputAtSuccess)))
	}

	sb.WriteString("\n")

	// ステータス
	if result.Success {
		sb.WriteString("Status: ✓ Success\n")
	} else {
		sb.WriteString("Status: ✗ Failed\n")
		if result.ErrorMessage != "" {
			sb.WriteString(fmt.Sprintf("Error:  %s\n", result.ErrorMessage))
		}
	}

	return sb.String()
}

// FormatMaxOutputResult はMax Output探索結果を整形
func (tf *TableFormatter) FormatMaxOutputResult(result *probe.MaxOutputResult) string {
	var sb strings.Builder

	// ヘッダー
	sb.WriteString("Max Output Tokens Probe Results\n")
	sb.WriteString(strings.Repeat("=", 32) + "\n")

	// データ行
	sb.WriteString(fmt.Sprintf("%-22s %s\n", "Model:", result.Model))
	sb.WriteString(fmt.Sprintf("%-22s %s\n", "Max Output Tokens:", formatNumber(result.MaxOutputTokens)))
	sb.WriteString(fmt.Sprintf("%-22s %s\n", "Evidence:", result.Evidence))
	sb.WriteString(fmt.Sprintf("%-22s %d\n", "Trials:", result.Trials))
	sb.WriteString(fmt.Sprintf("%-22s %s\n", "Duration:", formatDuration(result.Duration)))

	if result.MaxSuccessfullyGenerated > 0 {
		sb.WriteString(fmt.Sprintf("%-22s %s tokens\n", "Max Successfully Gen:", formatNumber(result.MaxSuccessfullyGenerated)))
	}

	sb.WriteString("\n")

	// ステータス
	if result.Success {
		sb.WriteString("Status: ✓ Success\n")
	} else {
		sb.WriteString("Status: ✗ Failed\n")
		if result.ErrorMessage != "" {
			sb.WriteString(fmt.Sprintf("Error:  %s\n", result.ErrorMessage))
		}
	}

	return sb.String()
}

// formatNumber は数値を3桁区切りで整形
func formatNumber(n int) string {
	if n == 0 {
		return "0"
	}

	s := fmt.Sprintf("%d", n)
	var result []rune

	for i, r := range reverse(s) {
		if i > 0 && i%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, r)
	}

	return string(reverse(string(result)))
}

// formatDuration は時間を読みやすく整形
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// reverse は文字列を反転
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// FormatVerboseHistory は探索履歴を整形（verbose用）
func (tf *TableFormatter) FormatVerboseHistory(trials []probe.TrialInfo) string {
	var sb strings.Builder

	sb.WriteString("\nSearch History:\n")
	sb.WriteString(strings.Repeat("-", 60) + "\n")
	sb.WriteString(fmt.Sprintf("%-8s %-15s %-12s %s\n", "Trial", "Tokens", "Result", "Message"))
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	for i, trial := range trials {
		status := "✓"
		if !trial.Success {
			status = "✗"
		}

		msg := trial.Message
		if len(msg) > 30 {
			msg = msg[:27] + "..."
		}

		sb.WriteString(fmt.Sprintf("%-8d %-15s %-12s %s\n",
			i+1,
			formatNumber(trial.TokenCount),
			status,
			msg,
		))
	}

	return sb.String()
}