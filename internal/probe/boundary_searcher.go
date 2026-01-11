package probe

import (
	"fmt"
	"math"
	"regexp"
	"time"
)

// BoundarySearchResult は探索結果を表す
type BoundarySearchResult struct {
	Value           int
	Success         bool
	ErrorMessage    string
	Source          string // "validation_error" or "max_output_incomplete"
	Trials          int
	EstimatedTokens int
}

// BoundarySearcher は境界値を効率的に探索する
type BoundarySearcher struct {
	maxTrials    int
	initialValue int
}

// NewBoundarySearcher は新しい BoundarySearcher を作成する
func NewBoundarySearcher() *BoundarySearcher {
	return &BoundarySearcher{
		maxTrials:    40,
		initialValue: 4096,
	}
}

// Search は下界と上界を指定して境界値を探す
func (bs *BoundarySearcher) Search(lower, upper int, runner func(int) (*BoundarySearchResult, error)) (*BoundarySearchResult, error) {
	lowerBound := lower
	upperBound := upper
	trials := 0

	// 二分探索の実行
	for upperBound-lowerBound > 128 && trials < bs.maxTrials {
		mid := (lowerBound + upperBound) / 2
		result, err := runner(mid)

		if err != nil {
			return nil, err
		}

		trials++

		// 成功/失敗に応じて境界を更新
		if result.Success {
			lowerBound = mid
		} else {
			upperBound = mid
		}

		// API呼び出し間の待機（レート制限対策）
		time.Sleep(1 * time.Second)
	}

	// 最終的な下界が成功した場合
	successResult, err := runner(lowerBound)
	if err != nil {
		return nil, err
	}

	return &BoundarySearchResult{
		Value:           lowerBound,
		Success:         successResult.Success,
		ErrorMessage:    successResult.ErrorMessage,
		Source:          successResult.Source,
		Trials:          trials + 1,
		EstimatedTokens: lowerBound,
	}, nil
}

// ExponentialSearch は指数探索で上限を見つける
func (bs *BoundarySearcher) ExponentialSearch(runner func(int) (*BoundarySearchResult, error)) (*BoundarySearchResult, error) {
	value := bs.initialValue
	trials := 0
	var lastSuccessValue int

	// 成功するまで2倍ずつ増やしていく
	for trials < bs.maxTrials {
		result, err := runner(value)

		if err != nil {
			return &BoundarySearchResult{
				Value:           value,
				Success:         false,
				ErrorMessage:    err.Error(),
				Source:          "error",
				Trials:          trials + 1,
				EstimatedTokens: 0,
			}, nil
		}

		trials++

		if result.Success {
			lastSuccessValue = value
			// 成功した場合、さらに次の値で試して失敗した場合の境界を特定
			nextValue := value * 2
			if nextResult, nextErr := runner(nextValue); nextErr == nil && !nextResult.Success {
				return &BoundarySearchResult{
					Value:           value,
					Success:         true,
					ErrorMessage:    "",
					Source:          "success",
					Trials:          trials + 1,
					EstimatedTokens: value,
				}, nil
			}
			// 次の値でも成功した場合は探索を続ける
			value = nextValue
			continue
		}

		// 失敗した場合、その前の成功値が境界となる
		if lastSuccessValue > 0 {
			return &BoundarySearchResult{
				Value:           lastSuccessValue,
				Success:         true,
				ErrorMessage:    "",
				Source:          "success",
				Trials:          trials + 1,
				EstimatedTokens: lastSuccessValue,
			}, nil
		}

		// 失敗した場合、値を増やして探索を続ける
		value *= 2
	}

	// 最大試行回数に達した場合
	return &BoundarySearchResult{
		Value:           value,
		Success:         false,
		ErrorMessage:    "max_trials_reached",
		Source:          "error",
		Trials:          trials,
		EstimatedTokens: value,
	}, nil
}

// ExtractTokenLimitFromError はエラーメッセージからトークン制限を抽出する
func (bs *BoundarySearcher) ExtractTokenLimitFromError(errorMessage string) (int, bool) {
	// 一般的な最大コンテキスト長のパターン
	patterns := []string{
		`maximum context length is (\d+) tokens`,
		`your request resulted in (\d+) tokens`,
		`this model's maximum context length is (\d+) tokens`,
		`prompt tokens must be less than (\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(errorMessage)
		if len(matches) > 1 {
			// 最初のキャプチャグループを数値として返す
			var result int
			_, err := fmt.Sscanf(matches[1], "%d", &result)
			if err == nil {
				return result, true
			}
		}
	}
	return 0, false
}

// CalculateConfidence は探索結果の信頼度を計算する
func (bs *BoundarySearcher) CalculateConfidence(trials int, errorSource string, exactValue int) string {
	if errorSource == "success" && trials > 5 {
		return "high"
	}
	if errorSource == "validation_error" {
		return "high"
	}
	if exactValue > 0 && (float64(trials)/math.Log2(float64(exactValue)) > 0.8) {
		return "high"
	}
	return "medium"
}