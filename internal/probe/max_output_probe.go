package probe

import (
	"fmt"
	"regexp"
	"time"

	"github.com/armaniacs/llm-info/internal/api"
	"github.com/armaniacs/llm-info/pkg/config"
)

// MaxOutputTokensProbe はmax output tokensを探索する
type MaxOutputTokensProbe struct {
	client    *api.ProbeClient
	generator *TestDataGenerator
	searcher  *BoundarySearcher
}

// NewMaxOutputTokensProbe は新しいMaxOutputTokensProbeを作成する
func NewMaxOutputTokensProbe(client *api.ProbeClient) *MaxOutputTokensProbe {
	return &MaxOutputTokensProbe{
		client:    client,
		generator: NewTestDataGenerator(),
		searcher:  NewBoundarySearcher(),
	}
}

// ProbeOutputTokens は指定されたモデルのmax output tokensを推定する
func (p *MaxOutputTokensProbe) ProbeOutputTokens(model string, verbose bool) (*MaxOutputResult, error) {
	startTime := time.Now()

	// 十分な入力長を確保する（推定：context windowの50%）
	// TODO: Context Window探索結果を利用する改善ポイント
	inputTokens := 1000 // 一時的な固定値

	// 第1段階: 指数探索で上限を特定
	upperLimit, err := p.searcher.ExponentialSearch(func(tokens int) (*BoundarySearchResult, error) {
		return p.testWithMaxTokens(model, inputTokens, tokens, verbose)
	})

	if err != nil {
		return nil, fmt.Errorf("exponential search phase failed: %w", err)
	}

	// 上限が見つからなかった場合
	if !upperLimit.Success {
		return &MaxOutputResult{
			Model:             model,
			MaxOutputTokens:   upperLimit.Value,
			MethodConfidence:  "low",
			Trials:            upperLimit.Trials,
			Duration:          time.Since(startTime),
			Error:             upperLimit.ErrorMessage,
			InputTokensUsed:   inputTokens,
		}, nil
	}

	// バリデーションエラーから値を抽出した場合
	if tokenLimit, found := p.extractMaxTokensFromError(upperLimit.ErrorMessage); found {
		return &MaxOutputResult{
			Model:             model,
			MaxOutputTokens:   tokenLimit,
			MethodConfidence:  p.searcher.CalculateConfidence(upperLimit.Trials, "validation_error", tokenLimit),
			Trials:            upperLimit.Trials,
			Duration:          time.Since(startTime),
			Error:             upperLimit.ErrorMessage,
			InputTokensUsed:   inputTokens,
			Evidence:          "validation_error",
		}, nil
	}

	// 第2段階: 二分探索で境界を絞る
	boundaryResult, err := p.searcher.Search(upperLimit.Value/2, upperLimit.Value, func(tokens int) (*BoundarySearchResult, error) {
		return p.testWithMaxTokens(model, inputTokens, tokens, verbose)
	})

	if err != nil {
		return nil, fmt.Errorf("binary search phase failed: %w", err)
	}

	// 結果の整形
	result := &MaxOutputResult{
		Model:                 model,
		MaxOutputTokens:       boundaryResult.Value,
		MethodConfidence:      p.searcher.CalculateConfidence(boundaryResult.Trials, boundaryResult.Source, boundaryResult.Value),
		Trials:                upperLimit.Trials + boundaryResult.Trials + 1,
		Duration:              time.Since(startTime),
		InputTokensUsed:       inputTokens,
		MaxSuccessfullyGenerated: boundaryResult.Value,
		Evidence:              boundaryResult.Source,
	}

	return result, nil
}

// testWithMaxTokens は指定されたmax tokensでテストを実行する
func (p *MaxOutputTokensProbe) testWithMaxTokens(model string, inputTokens, maxTokens int, verbose bool) (*BoundarySearchResult, error) {
	// APIクライアント設定
	cfg := p.client.GetConfig()
	adjustedCfg := config.NewAppConfig()
	adjustedCfg.BaseURL = cfg.BaseURL
	adjustedCfg.APIKey = cfg.APIKey
	adjustedCfg.Timeout = cfg.Timeout

	client := api.NewProbeClient(adjustedCfg)

	// APIリクエストを送信
	response, err := client.ProbeModel(model)
	if err != nil {
		// エラーメッセージから情報を抽出
		errorMessage := ""
		if response != nil && response.Error != nil {
			errorMessage = response.Error.Message
		}

		// max_output_tokensエラーかチェック
		if tokenLimit, found := p.extractMaxTokensFromError(errorMessage); found {
			return &BoundarySearchResult{
				Value:           tokenLimit,
				Success:         false,
				ErrorMessage:    errorMessage,
				Source:          "validation_error",
				Trials:          1,
				EstimatedTokens: tokenLimit,
			}, nil
		}

		return &BoundarySearchResult{
			Value:           0,
			Success:         false,
			ErrorMessage:    err.Error(),
			Source:          "error",
			Trials:          1,
			EstimatedTokens: 0,
		}, nil
	}

	// レスポンスをチェック
	if response.Error != nil {
		// max_output_tokensエラーかチェック
		if tokenLimit, found := p.extractMaxTokensFromError(response.Error.Message); found {
			return &BoundarySearchResult{
				Value:           tokenLimit,
				Success:         false,
				ErrorMessage:    response.Error.Message,
				Source:          "validation_error",
				Trials:          1,
				EstimatedTokens: tokenLimit,
			}, nil
		}

		return &BoundarySearchResult{
			Value:        0,
			Success:      false,
			ErrorMessage: response.Error.Message,
			Source:      "api_error",
			Trials:        1,
			EstimatedTokens: 0,
		}, nil
	}

	// incomplete statusをチェック
	if len(response.Choices) > 0 {
		choice := response.Choices[0]
		if choice.FinishReason == "length" {
			// max_output_tokensで打ち切られた場合
			actualTokens := 0
			if response.Usage != nil {
				actualTokens = response.Usage.CompletionTokens
			}
			return &BoundarySearchResult{
				Value:           actualTokens,
				Success:         false,
				ErrorMessage:    "Response truncated due to max_output_tokens",
				Source:          "max_output_incomplete",
				Trials:          1,
				EstimatedTokens: actualTokens,
			}, nil
		}
	}

	// 成功した場合
	actualTokens := maxTokens
	if response.Usage != nil {
		actualTokens = response.Usage.CompletionTokens
	}

	return &BoundarySearchResult{
		Value:           actualTokens,
		Success:         true,
		ErrorMessage:    "",
		Source:          "success",
		Trials:          1,
		EstimatedTokens: actualTokens,
	}, nil
}

// extractMaxTokensFromError はエラーメッセージからmax_output_tokens制限を抽出する
func (p *MaxOutputTokensProbe) extractMaxTokensFromError(errorMessage string) (int, bool) {
	patterns := []string{
		`max_output_tokens must be <= (\d+)`,
		`maximum output tokens is (\d+)`,
		`the value of max_output_tokens should be <= (\d+)`,
		`max_tokens.*must be.*<= (\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(errorMessage)
		if len(matches) > 1 {
			var result int
			_, err := fmt.Sscanf(matches[1], "%d", &result)
			if err == nil {
				return result, true
			}
		}
	}
	return 0, false
}

// generateLongPrompt は指定されたトークン数に合わせて長いプロンプトを生成する
func generateLongPrompt(targetTokens int) string {
	prompt := "非常に詳細な説明を生成してください。できるだけ長く、詳細な文章で回答してください。"
	prompt += "各トピックについて深く掘り下げ、例を挙げ、背後にある原理を説明してください。"

	// 目標トークン数に達するまで繰り返す
	for len(prompt) < targetTokens*3 {
		prompt += "さらに詳細を追加してください。具体的な例、歴史的背景、技術的な詳細、"
		prompt += "応用例、関連する概念、比較分析、将来展望などについても説明してください。"
	}

	return prompt
}

// MaxOutputResult は探索結果を表す
type MaxOutputResult struct {
	Model                   string
	MaxOutputTokens         int    // 推定された最大出力トークン数
	MethodConfidence        string // high/medium/low
	Trials                  int    // 試行回数
	Duration                time.Duration
	Error                   string
	InputTokensUsed         int    // 使用した入力トークン数
	Evidence                 string // "validation_error" or "max_output_incomplete" or "success"
	MaxSuccessfullyGenerated int    // 実際に生成できた最大トークン数
}

// String は結果を文字列として返す
func (r *MaxOutputResult) String() string {
	if r.Error != "" {
		return fmt.Sprintf("Error probing: %s", r.Error)
	}

	return fmt.Sprintf(
		"Model: %s\n"+
			"Max Output Tokens: %d\n"+
			"Method Confidence: %s\n"+
			"Trials: %d\n"+
			"Duration: %v\n"+
			"Input Tokens Used: %d\n"+
			"Evidence: %s\n"+
			"Max Successfully Generated: %d\n",
		r.Model,
		r.MaxOutputTokens,
		r.MethodConfidence,
		r.Trials,
		r.Duration.Round(time.Second),
		r.InputTokensUsed,
		r.Evidence,
		r.MaxSuccessfullyGenerated,
	)
}