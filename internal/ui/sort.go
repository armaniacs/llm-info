package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/your-org/llm-info/internal/model"
)

// SortField はソートフィールドを表す
type SortField int

const (
	SortByName SortField = iota
	SortByMaxTokens
	SortByInputCost
	SortByMode
)

// SortOrder はソート順序を表す
type SortOrder int

const (
	Ascending SortOrder = iota
	Descending
)

// SortCriteria はソート条件を表す
type SortCriteria struct {
	Field SortField
	Order SortOrder
}

// Sort はソート条件に基づいてモデルをソートする
func Sort(models []model.Model, criteria *SortCriteria) {
	if criteria == nil {
		// デフォルトは名前の昇順
		criteria = &SortCriteria{Field: SortByName, Order: Ascending}
	}

	sort.Slice(models, func(i, j int) bool {
		return compare(models[i], models[j], criteria)
	})
}

// compare は2つのモデルを比較する
func compare(a, b model.Model, criteria *SortCriteria) bool {
	var result bool

	switch criteria.Field {
	case SortByName:
		result = strings.ToLower(a.Name) < strings.ToLower(b.Name)
	case SortByMaxTokens:
		result = a.MaxTokens < b.MaxTokens
	case SortByInputCost:
		result = a.InputCost < b.InputCost
	case SortByMode:
		result = strings.ToLower(a.Mode) < strings.ToLower(b.Mode)
	}

	if criteria.Order == Descending {
		result = !result
	}

	return result
}

// ParseSortString はソート文字列を解析してSortCriteriaを返す
func ParseSortString(sortStr string) (*SortCriteria, error) {
	if sortStr == "" {
		return &SortCriteria{Field: SortByName, Order: Ascending}, nil
	}

	// 降順の場合はプレフィックスをチェック
	order := Ascending
	if strings.HasPrefix(sortStr, "-") {
		order = Descending
		sortStr = strings.TrimPrefix(sortStr, "-")
	}

	// フィールドの判定
	var field SortField
	switch strings.ToLower(sortStr) {
	case "name", "model":
		field = SortByName
	case "tokens", "max_tokens":
		field = SortByMaxTokens
	case "cost", "input_cost":
		field = SortByInputCost
	case "mode":
		field = SortByMode
	default:
		return nil, fmt.Errorf("unknown sort field: %s", sortStr)
	}

	return &SortCriteria{Field: field, Order: order}, nil
}
