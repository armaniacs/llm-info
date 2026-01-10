package ui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/armaniacs/llm-info/internal/model"
)

// FilterCriteria はフィルタ条件を表す
type FilterCriteria struct {
	NamePattern    string   // モデル名のパターン（正規表現）
	MinTokens      int      // 最小トークン数
	MaxTokens      int      // 最大トークン数
	Modes          []string // 許可するモード
	MinInputCost   float64  // 最小入力コスト
	MaxInputCost   float64  // 最大入力コスト
	ExcludePattern string   // 除外するパターン
}

// Filter はフィルタ条件に基づいてモデルをフィルタリングする
func Filter(models []model.Model, criteria *FilterCriteria) []model.Model {
	if criteria == nil {
		return models
	}

	var filtered []model.Model
	for _, model := range models {
		if matchesCriteria(model, criteria) {
			filtered = append(filtered, model)
		}
	}

	return filtered
}

// matchesCriteria はモデルがフィルタ条件に一致するかチェックする
func matchesCriteria(model model.Model, criteria *FilterCriteria) bool {
	if criteria == nil {
		return true
	}

	// 名前パターンのチェック
	if criteria.NamePattern != "" {
		if matched, _ := regexp.MatchString(criteria.NamePattern, model.Name); !matched {
			return false
		}
	}

	// 除外パターンのチェック
	if criteria.ExcludePattern != "" {
		if matched, _ := regexp.MatchString(criteria.ExcludePattern, model.Name); matched {
			return false
		}
	}

	// トークン数の範囲チェック
	if criteria.MinTokens > 0 && model.MaxTokens < criteria.MinTokens {
		return false
	}
	if criteria.MaxTokens > 0 && model.MaxTokens > criteria.MaxTokens {
		return false
	}

	// モードのチェック
	if len(criteria.Modes) > 0 {
		found := false
		for _, mode := range criteria.Modes {
			if strings.EqualFold(model.Mode, mode) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 入力コストの範囲チェック
	if criteria.MinInputCost > 0 && model.InputCost < criteria.MinInputCost {
		return false
	}
	if criteria.MaxInputCost > 0 && model.InputCost > criteria.MaxInputCost {
		return false
	}

	return true
}

// ParseFilterString はフィルタ文字列を解析してFilterCriteriaを返す
func ParseFilterString(filterStr string) (*FilterCriteria, error) {
	if filterStr == "" {
		return nil, nil
	}

	criteria := &FilterCriteria{}

	// フィルタ文字列のパース（例: "name:gpt,tokens>1000,cost<0.001"）
	parts := strings.Split(filterStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if err := parseFilterPart(part, criteria); err != nil {
			return nil, err
		}
	}

	return criteria, nil
}

// parseFilterPart は個別のフィルタ条件を解析する
func parseFilterPart(part string, criteria *FilterCriteria) error {
	// 名前フィルタ（例: "name:gpt"）
	if strings.HasPrefix(part, "name:") {
		criteria.NamePattern = strings.TrimPrefix(part, "name:")
		return nil
	}

	// 除外フィルタ（例: "exclude:beta"）
	if strings.HasPrefix(part, "exclude:") {
		criteria.ExcludePattern = strings.TrimPrefix(part, "exclude:")
		return nil
	}

	// トークン数フィルタ（例: "tokens>1000", "tokens<100000"）
	if strings.Contains(part, "tokens") {
		return parseTokenFilter(part, criteria)
	}

	// コストフィルタ（例: "cost>0.001", "cost<0.01"）
	if strings.Contains(part, "cost") {
		return parseCostFilter(part, criteria)
	}

	// モードフィルタ（例: "mode:chat"）
	if strings.HasPrefix(part, "mode:") {
		mode := strings.TrimPrefix(part, "mode:")
		criteria.Modes = append(criteria.Modes, mode)
		return nil
	}

	// 単純な文字列の場合は名前パターンとして扱う
	criteria.NamePattern = part
	return nil
}

// parseTokenFilter はトークン数フィルタを解析する
func parseTokenFilter(part string, criteria *FilterCriteria) error {
	if strings.Contains(part, ">") {
		parts := strings.Split(part, ">")
		if len(parts) != 2 || parts[0] != "tokens" {
			return fmt.Errorf("invalid token filter format: %s", part)
		}
		value, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid token value: %s", parts[1])
		}
		criteria.MinTokens = value
	} else if strings.Contains(part, "<") {
		parts := strings.Split(part, "<")
		if len(parts) != 2 || parts[0] != "tokens" {
			return fmt.Errorf("invalid token filter format: %s", part)
		}
		value, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid token value: %s", parts[1])
		}
		criteria.MaxTokens = value
	}

	return nil
}

// parseCostFilter はコストフィルタを解析する
func parseCostFilter(part string, criteria *FilterCriteria) error {
	if strings.Contains(part, ">") {
		parts := strings.Split(part, ">")
		if len(parts) != 2 || parts[0] != "cost" {
			return fmt.Errorf("invalid cost filter format: %s", part)
		}
		value, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return fmt.Errorf("invalid cost value: %s", parts[1])
		}
		criteria.MinInputCost = value
	} else if strings.Contains(part, "<") {
		parts := strings.Split(part, "<")
		if len(parts) != 2 || parts[0] != "cost" {
			return fmt.Errorf("invalid cost filter format: %s", part)
		}
		value, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return fmt.Errorf("invalid cost value: %s", parts[1])
		}
		criteria.MaxInputCost = value
	}

	return nil
}
