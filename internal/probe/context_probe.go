package probe

import (
	"fmt"
	"strings"
	"time"

	"github.com/armaniacs/llm-info/internal/api"
	"github.com/armaniacs/llm-info/pkg/config"
)

// ContextWindowProbe はコンテキストウィンドウを探索する
type ContextWindowProbe struct {
	client                *api.ProbeClient
	generator             *TestDataGenerator
	searcher              *BoundarySearcher
	lastComprehensionResult []bool  // test-all-positions用の一時的な保存領域
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
			ErrorMessage:    upperLimit.ErrorMessage,
			Success:          false,
		}, nil
	}

	// 値がエラーメッセージから抽出された場合
	if tokenLimit, found := p.searcher.ExtractTokenLimitFromError(upperLimit.ErrorMessage); found {
		return &ContextWindowResult{
			MaxContextTokens: tokenLimit,
			MethodConfidence: p.searcher.CalculateConfidence(upperLimit.Trials, upperLimit.Source, tokenLimit),
			Trials:           upperLimit.Trials,
			Duration:        time.Since(startTime),
			ErrorMessage:    upperLimit.ErrorMessage,
			Source:          "validation_error",
			Success:         true,
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

// ProbeWithNeedle はneedle位置を指定してcontext windowを推定する
func (p *ContextWindowProbe) ProbeWithNeedle(model string, position NeedlePosition, needleKeyword, needleAnswer string, verbose bool) (*ContextWindowResult, error) {
	startTime := time.Now()

	// デフォルト値を設定
	if needleKeyword == "" {
		needleKeyword = "【重要情報】ラッキーカラーは青色です"
	}
	if needleAnswer == "" {
		needleAnswer = "青色"
	}

	// 第1段階: 指数探索で上限を特定
	upperLimit, err := p.searcher.ExponentialSearch(func(tokens int) (*BoundarySearchResult, error) {
		return p.testWithNeedlePosition(model, tokens, position, needleKeyword, needleAnswer, verbose)
	})

	if err != nil {
		return nil, fmt.Errorf("exponential search phase failed: %w", err)
	}

	// 上限が見つからなかった場合
	if !upperLimit.Success {
		return &ContextWindowResult{
			Model:               model,
			MaxContextTokens:    upperLimit.Value,
			MethodConfidence:    "low",
			Trials:              upperLimit.Trials,
			Duration:            time.Since(startTime),
			ErrorMessage:        upperLimit.ErrorMessage,
			Success:             false,
			NeedlePosition:      position,
			NeedleKeyword:       needleKeyword,
			NeedleAnswer:        needleAnswer,
		}, nil
	}

	// 値がエラーメッセージから抽出された場合
	if tokenLimit, found := p.searcher.ExtractTokenLimitFromError(upperLimit.ErrorMessage); found {
		return &ContextWindowResult{
			Model:               model,
			MaxContextTokens:    tokenLimit,
			MethodConfidence:    p.searcher.CalculateConfidence(upperLimit.Trials, upperLimit.Source, tokenLimit),
			Trials:              upperLimit.Trials,
			Duration:            time.Since(startTime),
			ErrorMessage:        upperLimit.ErrorMessage,
			Source:              "validation_error",
			Success:             true,
			NeedlePosition:      position,
			NeedleKeyword:       needleKeyword,
			NeedleAnswer:        needleAnswer,
		}, nil
	}

	// 第2段階: 二分探索で境界を絞る
	boundaryResult, err := p.searcher.Search(upperLimit.Value-1024, upperLimit.Value+1024, func(tokens int) (*BoundarySearchResult, error) {
		return p.testWithNeedlePosition(model, tokens, position, needleKeyword, needleAnswer, verbose)
	})

	if err != nil {
		return nil, fmt.Errorf("binary search phase failed: %w", err)
	}

	// 結果の整形
	result := &ContextWindowResult{
		Model:               model,
		MaxContextTokens:    boundaryResult.Value,
		MethodConfidence:    p.searcher.CalculateConfidence(boundaryResult.Trials, boundaryResult.Source, boundaryResult.Value),
		Trials:              upperLimit.Trials + boundaryResult.Trials + 1,
		Duration:            time.Since(startTime),
		Success:             true,
		NeedlePosition:      position,
		NeedleKeyword:       needleKeyword,
		NeedleAnswer:        needleAnswer,
	}

	return result, nil
}

// ProbeAllNeedlePositions は全てのneedle位置をテストする
func (p *ContextWindowProbe) ProbeAllNeedlePositions(model string, needleKeyword, needleAnswer string, verbose bool) (*ContextWindowResult, error) {
	startTime := time.Now()

	// デフォルト値を設定
	if needleKeyword == "" {
		needleKeyword = "【重要情報】ラッキーカラーは青色です"
	}
	if needleAnswer == "" {
		needleAnswer = "青色"
	}

	positions := []NeedlePosition{End, Middle, Percent80}
	var needleTests []NeedleTestResult
	maxTokens := 0
	var allErrors []string

	// 各位置でテストを実行
	for _, pos := range positions {
		if verbose {
			fmt.Printf("Testing needle at position: %s\n", pos)
		}

		upperLimit, err := p.searcher.ExponentialSearch(func(tokens int) (*BoundarySearchResult, error) {
			return p.testWithNeedlePosition(model, tokens, pos, needleKeyword, needleAnswer, verbose)
		})

		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("%s: %v", pos, err))
			continue
		}

		var comprehension bool
		var tokenCount int

		if upperLimit.Success {
			// 最後成功した試行でcomprehensionをチェック
			if len(p.lastComprehensionResult) > 0 {
				comprehension = p.lastComprehensionResult[len(p.lastComprehensionResult)-1]
			}
		}

		// トークン数を取得
		if tokenLimit, found := p.searcher.ExtractTokenLimitFromError(upperLimit.ErrorMessage); found {
			tokenCount = tokenLimit
		} else if upperLimit.Success {
			tokenCount = upperLimit.Value
		}

		needleTests = append(needleTests, NeedleTestResult{
			Position:      pos,
			Comprehension: comprehension,
			TokenCount:   tokenCount,
			Error:        upperLimit.ErrorMessage,
		})

		if tokenCount > maxTokens {
			maxTokens = tokenCount
		}
	}

	// 結果を構築
	var errorMsg string
	if len(allErrors) > 0 {
		errorMsg = strings.Join(allErrors, "; ")
	}

	averageComprehension := 0
	if len(needleTests) > 0 {
		comprehensionCount := 0
		for _, test := range needleTests {
			if test.Comprehension {
				comprehensionCount++
			}
		}
		averageComprehension = comprehensionCount * 100 / len(needleTests)
	}

	return &ContextWindowResult{
		Model:               model,
		MaxContextTokens:    maxTokens,
		MethodConfidence:    "medium",
		Trials:              len(needleTests) * 10, // 概算値
		Duration:            time.Since(startTime),
		Success:             maxTokens > 0,
		ErrorMessage:        errorMsg,
		NeedlePosition:      Percent80, // デフォルト表示
		NeedleKeyword:       needleKeyword,
		NeedleAnswer:        needleAnswer,
		NeedleComprehension: averageComprehension > 0,
		NeedleTests:         needleTests,
	}, nil
}

// testWithNeedlePosition はneedle位置を指定してテストを実行する
func (p *ContextWindowProbe) testWithNeedlePosition(model string, tokens int, position NeedlePosition, needleKeyword, needleAnswer string, verbose bool) (*BoundarySearchResult, error) {
	// テストデータを生成
	content, _ := p.generator.GenerateWithNeedlePosition(tokens, position)

	// needleキーワードと回答を置換
	content = strings.Replace(content, "【重要情報】ラッキーカラーは青色です", needleKeyword, -1)
	content = strings.Replace(content, "ラッキーカラーは何色でしたか？", strings.Split(needleKeyword, "は")[1]+"は何色でしたか？", -1)

	// APIクライアントを作成（既存のprobeクライアントを利用）
	cfg := p.client.GetConfig()

	// 注: Configをコピーしてtimeoutのみ調整
	adjustedCfg := config.NewAppConfig()
	adjustedCfg.BaseURL = cfg.BaseURL
	adjustedCfg.APIKey = cfg.APIKey
	adjustedCfg.Timeout = cfg.Timeout

	client := api.NewProbeClient(adjustedCfg)

	// APIリクエストを送信
	response, err := client.ProbeModelWithContent(model, content)
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
			}, nil
		}

		return &BoundarySearchResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("API request failed: %v", err),
		}, nil
	}

	// 成功した場合
	if response.Usage != nil && response.Usage.PromptTokens > 0 {
		// Comprehensionをチェック
		var comprehension bool
		if len(response.Choices) > 0 && response.Choices[0].Message.Content != "" {
			result := CheckComprehension(response.Choices[0].Message.Content, needleAnswer)
			comprehension = result.Correct

			// 結果を保存（test-all-positionsで使用）
			p.lastComprehensionResult = append(p.lastComprehensionResult, comprehension)
		}

		return &BoundarySearchResult{
			Value:         response.Usage.PromptTokens,
			Success:       true,
			ErrorMessage: "",
		}, nil
	}

	// Usageフィールドがない場合、エラーとして返す
	return &BoundarySearchResult{
		Success:      false,
		ErrorMessage: "No usage information in response",
	}, nil
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

// TrialInfo は試行情報
type TrialInfo struct {
	TokenCount int
	Success    bool
	Message    string
}

// ContextWindowResult は探索結果を表す
type ContextWindowResult struct {
	Model             string
	MaxContextTokens  int    // *実際の*最大コンテキストトークン数
	MethodConfidence  string // high/medium/low
	Trials            int    // 試行した試行回数
	Duration          time.Duration
	MaxInputAtSuccess int    // 最後に成功した入力トークン数
	Success           bool   // 成功フラグ
	ErrorMessage      string // エラー情報（あれば）
	Source            string // 情報ソース
	TrialHistory      []TrialInfo // 試行履歴

	// Needle test fields
	NeedlePosition      NeedlePosition // Needleの位置
	NeedleKeyword       string         // 使用されたneedleキーワード
	NeedleAnswer        string         // 期待される回答
	NeedleComprehension bool           // Needleを理解できたか
	NeedleTests         []NeedleTestResult // test-all-positionsの場合
}

// NeedleTestResult は各位置でのneedleテスト結果
type NeedleTestResult struct {
	Position       NeedlePosition // 位置
	Comprehension  bool           // 理解できたか
	TokenCount     int            // この位置での最大トークン数
	Error          string         // エラー情報（あれば）
}

// String は結果を文字列として返す
func (r *ContextWindowResult) String() string {
	if r.ErrorMessage != "" {
		return fmt.Sprintf("Error probing: %s", r.ErrorMessage)
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