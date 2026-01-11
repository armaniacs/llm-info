package probe

import (
	"fmt"
	"time"

	"github.com/armaniacs/llm-info/internal/api"
	"github.com/armaniacs/llm-info/pkg/config"
)

// ContextWindowProbe はコンテキストウィンドウを探索する
type ContextWindowProbe struct {
	client    *api.ProbeClient
	generator *TestDataGenerator
	searcher  *BoundarySearcher
}

// NewContextWindowProbe は新しいContextWindowProbeを作成する
func NewContextWindowProbe(client *api.ProbeClient) *ContextWindowProbe {
	return &ContextWindowProbe{
		client:    client,
		generator: NewTestDataGenerator(),
		searcher:  NewBoundarySearcher(),
	}
}

// Probe は指定されたモデルのcontext windowを推定する
func (p *ContextWindowProbe) Probe(model string, verbose bool) (*ContextWindowResult, error) {
	startTime := time.Now()

	// 第1段階: 指数探索で上限を特定
	upperLimit, err := p.searcher.ExponentialSearch(func(tokens int) (*BoundarySearchResult, error) {
		return p.testWithTokenCount(model, tokens, verbose)
	})

	if err != nil {
		return nil, fmt.Errorf("exponential search phase failed: %w", err)
	}

	// 上限が見つからなかった場合
	if !upperLimit.Success {
		return &ContextWindowResult{
			MaxContextTokens: upperLimit.Value,
			MethodConfidence: "low",
			Trials:           upperLimit.Trials,
			Duration:        time.Since(startTime),
			Error:            upperLimit.ErrorMessage,
		}, nil
	}

	// 値がエラーメッセージから抽出された場合
	if tokenLimit, found := p.searcher.ExtractTokenLimitFromError(upperLimit.ErrorMessage); found {
		return &ContextWindowResult{
			MaxContextTokens: tokenLimit,
			MethodConfidence: p.searcher.CalculateConfidence(upperLimit.Trials, upperLimit.Source, tokenLimit),
			Trials:           upperLimit.Trials,
			Duration:        time.Since(startTime),
			Error:            upperLimit.ErrorMessage,
			Source:            "validation_error",
		}, nil
	}

	// 第2段階: 二分探索で境界を絞る
	boundaryResult, err := p.searcher.Search(upperLimit.Value-1024, upperLimit.Value+1024, func(tokens int) (*BoundarySearchResult, error) {
		return p.testWithTokenCount(model, tokens, verbose)
	})

	if err != nil {
		return nil, fmt.Errorf("binary search phase failed: %w", err)
	}

	// 結果の整形
	result := &ContextWindowResult{
		Model:              model,
		MaxContextTokens:    boundaryResult.Value,
		MethodConfidence:    p.searcher.CalculateConfidence(boundaryResult.Trials, boundaryResult.Source, boundaryResult.Value),
		Trials:              upperLimit.Trials + boundaryResult.Trials + 1,
		Duration:            time.Since(startTime),
	}

	return result, nil
}

// testWithTokenCount は指定されたトークン数でテストを実行する
func (p *ContextWindowProbe) testWithTokenCount(model string, tokens int, verbose bool) (*BoundarySearchResult, error) {
	// テストデータを生成
	_, _ = p.generator.GenerateData(tokens)

	// APIクライアントを作成（既存のprobeクライアントを利用）
	cfg := p.client.GetConfig()

	// 注: Configをコピーしてtimeoutのみ調整
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

		// トークン制限エラーかチェック
		if tokenLimit, found := p.searcher.ExtractTokenLimitFromError(errorMessage); found {
			return &BoundarySearchResult{
				Value:        tokenLimit,
				Success:      false,
				ErrorMessage: errorMessage,
				Source:        "validation_error",
				Trials:        1,
				EstimatedTokens: 0,
			}, nil
		}

		return &BoundarySearchResult{
			Value:        0,
			Success:      false,
			ErrorMessage: err.Error(),
			Source:        "error",
			Trials:        1,
			EstimatedTokens: 0,
		}, nil
	}

	// レスポンスをチェック
	if response.Error != nil {
		return &BoundarySearchResult{
			Value:        0,
			Success:      false,
			ErrorMessage: response.Error.Message,
			Source:        "api_error",
			Trials:        1,
			EstimatedTokens: 0,
		}, nil
	}

	// 成功した場合 - Usageフィールドのnilチェック
	if response.Usage == nil {
		return &BoundarySearchResult{
			Value:           0,
			Success:         false,
			ErrorMessage:    "Response missing usage information",
			Source:          "api_error",
			Trials:          1,
			EstimatedTokens: 0,
		}, nil
	}

	return &BoundarySearchResult{
		Value:           response.Usage.PromptTokens,
		Success:         true,
		ErrorMessage:    "",
		Source:          "success",
		Trials:          1,
		EstimatedTokens: response.Usage.TotalTokens,
	}, nil
}

// ContextWindowResult は探索結果を表す
type ContextWindowResult struct {
	Model           string
	MaxContextTokens int    // *実際の*最大コンテキストトークン数
	MethodConfidence string    // high/medium/low
	Trials          int            // 試行した試行回数
	Duration        time.Duration   // 実行時間
	Error           string            // エラー情報（あれば）
	Source          string            // 情報ソース
}

// String は結果を文字列として返す
func (r *ContextWindowResult) String() string {
	if r.Error != "" {
		return fmt.Sprintf("Error probing: %s", r.Error)
	}

	return fmt.Sprintf(
		"Model: %s\n"+
			"Max Context Tokens: %d\n"+
			"Method Confidence: %s\n"+
			"Trials: %d\n"+
			"Duration: %v\n",
		r.Model,
		r.MaxContextTokens,
		r.MethodConfidence,
		r.Trials,
		r.Duration.Round(time.Second),
	)
}