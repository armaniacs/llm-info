package ui

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/armaniacs/llm-info/internal/model"
)

// JSONRenderer はJSON出力機能を提供する
type JSONRenderer struct {
	PrettyPrint bool
}

// NewJSONRenderer は新しいJSONレンダラーを作成する
func NewJSONRenderer(prettyPrint bool) *JSONRenderer {
	return &JSONRenderer{
		PrettyPrint: prettyPrint,
	}
}

// Render はモデル情報をJSON形式で表示する
func (jr *JSONRenderer) Render(models []model.Model, options *RenderOptions) error {
	var output interface{}

	if options != nil && options.Filter != "" {
		// フィルタ条件をメタデータとして含める
		output = map[string]interface{}{
			"filter": options.Filter,
			"models": models,
		}
	} else {
		output = models
	}

	var encoder *json.Encoder
	if jr.PrettyPrint {
		encoder = json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
	} else {
		encoder = json.NewEncoder(os.Stdout)
	}

	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// RenderJSON はモデル情報をJSON形式で表示します
func RenderJSON(models []model.Model) error {
	if len(models) == 0 {
		fmt.Println("{}")
		return nil
	}

	// JSON出力用の構造体に変換
	jsonModels := make([]JSONModel, len(models))
	for i, model := range models {
		jsonModels[i] = JSONModel{
			Name:      model.Name,
			MaxTokens: model.MaxTokens,
			Mode:      model.Mode,
			InputCost: model.InputCost,
		}
	}

	// JSONにエンコード
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(jsonModels); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// JSONModel はJSON出力用のモデル構造体です
type JSONModel struct {
	Name      string  `json:"name"`
	MaxTokens int     `json:"max_tokens,omitempty"`
	Mode      string  `json:"mode,omitempty"`
	InputCost float64 `json:"input_cost,omitempty"`
}

// RenderCompactJSON はモデル情報をコンパクトなJSON形式で表示します
func RenderCompactJSON(models []model.Model) error {
	if len(models) == 0 {
		fmt.Println("[]")
		return nil
	}

	// JSON出力用の構造体に変換
	jsonModels := make([]JSONModel, len(models))
	for i, model := range models {
		jsonModels[i] = JSONModel{
			Name:      model.Name,
			MaxTokens: model.MaxTokens,
			Mode:      model.Mode,
			InputCost: model.InputCost,
		}
	}

	// JSONにエンコード（インデントなし）
	data, err := json.Marshal(jsonModels)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// RenderJSONWithMetadata はメタデータ付きでモデル情報をJSON形式で表示します
func RenderJSONWithMetadata(models []model.Model, metadata map[string]interface{}) error {
	output := map[string]interface{}{
		"models":   models,
		"metadata": metadata,
	}

	// JSONにエンコード
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// RenderJSONWithOptions はオプション付きでモデル情報をJSON形式で表示します
func RenderJSONWithOptions(models []model.Model, options *RenderOptions) error {
	renderer := NewJSONRenderer(true)
	return renderer.Render(models, options)
}
