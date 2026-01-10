package model

import (
	"sort"
	"strings"

	"github.com/armaniacs/llm-info/internal/api"
)

// Model はアプリケーション内のモデルデータです
type Model struct {
	Name      string
	MaxTokens int
	Mode      string
	InputCost float64
}

// FromAPIResponse はAPIレスポンスをアプリケーションモデルに変換します
func FromAPIResponse(apiModels []api.ModelInfo) []Model {
	if apiModels == nil {
		return nil
	}

	models := make([]Model, len(apiModels))
	for i, apiModel := range apiModels {
		models[i] = Model{
			Name:      apiModel.ID,
			MaxTokens: apiModel.MaxTokens,
			Mode:      apiModel.Mode,
			InputCost: apiModel.InputCost,
		}
	}
	return models
}

// FilterByName はモデル名でフィルタリングします
func FilterByName(models []Model, filter string) []Model {
	if filter == "" {
		return models
	}

	var filtered []Model
	for _, model := range models {
		if strings.Contains(strings.ToLower(model.Name), strings.ToLower(filter)) {
			filtered = append(filtered, model)
		}
	}
	return filtered
}

// SortBy は指定されたフィールドでモデルをソートします
func SortBy(models []Model, field string) []Model {
	if field == "" {
		return models
	}

	switch field {
	case "name":
		sort.Slice(models, func(i, j int) bool {
			return models[i].Name < models[j].Name
		})
	case "max_tokens":
		sort.Slice(models, func(i, j int) bool {
			return models[i].MaxTokens > models[j].MaxTokens // 降順
		})
	case "mode":
		sort.Slice(models, func(i, j int) bool {
			return models[i].Mode < models[j].Mode
		})
	case "input_cost":
		sort.Slice(models, func(i, j int) bool {
			return models[i].InputCost < models[j].InputCost
		})
	}

	return models
}
