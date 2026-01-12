package cost

// UsageSummary はAPI使用量の集計結果
type UsageSummary struct {
	TotalCost         float64
	TotalInputTokens  int
	TotalOutputTokens int
	TotalTokens       int
	WarningTriggered  bool
	UnknownModelUsed  bool
	BreakdownByProbe  map[string]*ProbeUsage
}

// ProbeUsage はプローブごとの使用量
type ProbeUsage struct {
	Cost         float64
	InputTokens  int
	OutputTokens int
	Trials       int
}

// TrialUsage は試行ごとの使用量情報（import cycleを避けるための独立構造体）
type TrialUsage struct {
	PromptTokens     int
	CompletionTokens int
}

// AggregateFromTrials は試行履歴から実際のコストを集計する
func AggregateFromTrials(
	calculator *Calculator,
	modelName string,
	contextTrials []TrialUsage,
	outputTrials []TrialUsage,
) *UsageSummary {
	summary := &UsageSummary{
		BreakdownByProbe: make(map[string]*ProbeUsage),
	}

	// モデルが料金表に存在するかチェック
	_, exists := calculator.models[modelName]
	summary.UnknownModelUsed = !exists

	// Context Window Probe の集計
	if len(contextTrials) > 0 {
		contextUsage := aggregateProbeTrials(calculator, modelName, contextTrials)
		summary.BreakdownByProbe["context"] = contextUsage
		summary.TotalInputTokens += contextUsage.InputTokens
		summary.TotalOutputTokens += contextUsage.OutputTokens
		summary.TotalCost += contextUsage.Cost
	}

	// Max Output Probe の集計
	if len(outputTrials) > 0 {
		outputUsage := aggregateProbeTrials(calculator, modelName, outputTrials)
		summary.BreakdownByProbe["max_output"] = outputUsage
		summary.TotalInputTokens += outputUsage.InputTokens
		summary.TotalOutputTokens += outputUsage.OutputTokens
		summary.TotalCost += outputUsage.Cost
	}

	summary.TotalTokens = summary.TotalInputTokens + summary.TotalOutputTokens

	// 警告判定
	summary.WarningTriggered = summary.TotalCost > calculator.warningThreshold

	return summary
}

// aggregateProbeTrials は単一プローブの試行履歴を集計
func aggregateProbeTrials(calculator *Calculator, modelName string, trials []TrialUsage) *ProbeUsage {
	usage := &ProbeUsage{
		Trials: len(trials),
	}

	for _, trial := range trials {
		// トークン数が0の場合はスキップ（Usage情報がない試行）
		if trial.PromptTokens == 0 && trial.CompletionTokens == 0 {
			continue
		}

		usage.InputTokens += trial.PromptTokens
		usage.OutputTokens += trial.CompletionTokens

		// 各試行のコストを計算
		trialCost := calculator.CalculateTrialCost(
			trial.PromptTokens,
			trial.CompletionTokens,
			modelName,
		)
		usage.Cost += trialCost
	}

	return usage
}

// EstimateUsage はDry-run用の概算を計算
func EstimateUsage(
	calculator *Calculator,
	modelName string,
	contextOnly, outputOnly bool,
) *UsageSummary {
	summary := &UsageSummary{
		BreakdownByProbe: make(map[string]*ProbeUsage),
	}

	// モデルが料金表に存在するかチェック
	_, exists := calculator.models[modelName]
	summary.UnknownModelUsed = !exists

	if !outputOnly {
		// Context Window Probe の概算
		contextCost, contextInput, contextOutput := calculator.EstimateProbeCost(modelName, "context")
		summary.BreakdownByProbe["context"] = &ProbeUsage{
			Cost:         contextCost,
			InputTokens:  contextInput,
			OutputTokens: contextOutput,
			Trials:       12, // 概算試行回数
		}
		summary.TotalInputTokens += contextInput
		summary.TotalOutputTokens += contextOutput
		summary.TotalCost += contextCost
	}

	if !contextOnly {
		// Max Output Probe の概算
		outputCost, outputInput, outputOutput := calculator.EstimateProbeCost(modelName, "max_output")
		summary.BreakdownByProbe["max_output"] = &ProbeUsage{
			Cost:         outputCost,
			InputTokens:  outputInput,
			OutputTokens: outputOutput,
			Trials:       12, // 概算試行回数
		}
		summary.TotalInputTokens += outputInput
		summary.TotalOutputTokens += outputOutput
		summary.TotalCost += outputCost
	}

	summary.TotalTokens = summary.TotalInputTokens + summary.TotalOutputTokens

	// 警告判定
	summary.WarningTriggered = summary.TotalCost > calculator.warningThreshold

	return summary
}
