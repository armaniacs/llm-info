package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/armaniacs/llm-info/internal/model"
)

// TableRenderer はテーブル表示機能を提供する
type TableRenderer struct {
	columnManager *ColumnManager
}

// NewTableRenderer は新しいテーブルレンダラーを作成する
func NewTableRenderer() *TableRenderer {
	return &TableRenderer{
		columnManager: NewColumnManager(),
	}
}

// Render はモデル情報をテーブル形式で表示する
func (tr *TableRenderer) Render(models []model.Model, options *RenderOptions) error {
	if len(models) == 0 {
		fmt.Println("No models found.")
		return nil
	}

	if options != nil && options.Columns != "" {
		if err := tr.columnManager.ParseColumnsString(options.Columns); err != nil {
			return fmt.Errorf("failed to parse columns: %w", err)
		}
	}

	// 表示カラムの取得
	visibleColumns := tr.columnManager.GetVisibleColumns()

	// ヘッダーの構成
	var headers []string
	for _, col := range visibleColumns {
		headers = append(headers, col.Header)
	}

	// 列幅を計算
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	// データ行の準備
	var rows [][]string
	for _, model := range models {
		var row []string
		for _, col := range visibleColumns {
			value, err := tr.columnManager.GetColumnValue(model, col.Name)
			if err != nil {
				return err
			}

			var formattedValue string
			switch v := value.(type) {
			case string:
				formattedValue = v
			case int:
				formattedValue = fmt.Sprintf(col.Format, v)
			case float64:
				formattedValue = fmt.Sprintf(col.Format, v)
			default:
				formattedValue = fmt.Sprintf("%v", v)
			}

			row = append(row, formattedValue)

			// 列幅の更新
			if len(formattedValue) > colWidths[len(row)-1] {
				colWidths[len(row)-1] = len(formattedValue)
			}
		}
		rows = append(rows, row)
	}

	// テーブルの表示
	printTable(headers, rows, colWidths)
	return nil
}

// SetColumnVisibility はカラムの表示/非表示を設定する
func (tr *TableRenderer) SetColumnVisibility(columnName string, visible bool) error {
	return tr.columnManager.SetColumnVisibility(columnName, visible)
}

// RenderOptions は表示オプションを表す
type RenderOptions struct {
	Columns string // 表示するカラム（カンマ区切り）
	Filter  string // フィルタ条件
	Sort    string // ソート条件
}

// RenderTable はモデル情報をテーブル形式で表示します（互換性のための関数）
func RenderTable(models []model.Model) {
	if len(models) == 0 {
		fmt.Println("No models found.")
		return
	}

	// 利用可能なデータをチェック
	var hasMaxTokens, hasMode, hasInputCost bool
	for _, model := range models {
		if model.MaxTokens > 0 {
			hasMaxTokens = true
		}
		if model.Mode != "" {
			hasMode = true
		}
		if model.InputCost > 0 {
			hasInputCost = true
		}
	}

	// ヘッダーを動的に構成
	headers := []string{"Model Name"}
	if hasMaxTokens {
		headers = append(headers, "Max Tokens")
	}
	if hasMode {
		headers = append(headers, "Mode")
	}
	if hasInputCost {
		headers = append(headers, "Input Cost")
	}

	// 列幅を計算
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	// データ行の幅を計算
	var rows [][]string
	for _, model := range models {
		var row []string
		row = append(row, model.Name)

		if hasMaxTokens {
			tokens := strconv.Itoa(model.MaxTokens)
			row = append(row, tokens)
			if len(tokens) > colWidths[1] {
				colWidths[1] = len(tokens)
			}
		}
		if hasMode {
			row = append(row, model.Mode)
			modeIndex := 1
			if hasMaxTokens {
				modeIndex = 2
			}
			if len(model.Mode) > colWidths[modeIndex] {
				colWidths[modeIndex] = len(model.Mode)
			}
		}
		if hasInputCost {
			cost := fmt.Sprintf("%.6f", model.InputCost)
			row = append(row, cost)
			costIndex := 1
			if hasMaxTokens {
				costIndex = 2
			}
			if hasMode {
				costIndex = 3
			}
			if len(cost) > colWidths[costIndex] {
				colWidths[costIndex] = len(cost)
			}
		}

		rows = append(rows, row)
	}

	// テーブルを表示
	printTable(headers, rows, colWidths)
}

// RenderTableWithOptions はオプション付きでモデル情報をテーブル形式で表示します
func RenderTableWithOptions(models []model.Model, options *RenderOptions) error {
	renderer := NewTableRenderer()
	return renderer.Render(models, options)
}

// printTable はテーブルを表示します
func printTable(headers []string, rows [][]string, colWidths []int) {
	// ヘッダー行を表示
	printRow(headers, colWidths)

	// 区切り線を表示
	separators := make([]string, len(colWidths))
	for i, width := range colWidths {
		separators[i] = strings.Repeat("-", width)
	}
	printRow(separators, colWidths)

	// データ行を表示
	for _, row := range rows {
		printRow(row, colWidths)
	}
}

// printRow は行を表示します
func printRow(row []string, colWidths []int) {
	for i, cell := range row {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Printf("%-*s", colWidths[i], cell)
	}
	fmt.Println()
}
