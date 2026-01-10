package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/armaniacs/llm-info/internal/model"
)

// Column はテーブルカラムを表す
type Column struct {
	Name     string
	Header   string
	Visible  bool
	Width    int
	Format   string
	Priority int
}

// ColumnManager はカラム管理機能を提供する
type ColumnManager struct {
	columns []Column
}

// NewColumnManager は新しいカラムマネージャーを作成する
func NewColumnManager() *ColumnManager {
	return &ColumnManager{
		columns: []Column{
			{
				Name:     "name",
				Header:   "MODEL NAME",
				Visible:  true,
				Width:    30,
				Format:   "%s",
				Priority: 1,
			},
			{
				Name:     "max_tokens",
				Header:   "MAX TOKENS",
				Visible:  true,
				Width:    12,
				Format:   "%d",
				Priority: 2,
			},
			{
				Name:     "mode",
				Header:   "MODE",
				Visible:  true,
				Width:    8,
				Format:   "%s",
				Priority: 3,
			},
			{
				Name:     "input_cost",
				Header:   "INPUT COST",
				Visible:  true,
				Width:    12,
				Format:   "%.6f",
				Priority: 4,
			},
		},
	}
}

// GetVisibleColumns は表示可能なカラムを返す
func (cm *ColumnManager) GetVisibleColumns() []Column {
	var visible []Column
	for _, col := range cm.columns {
		if col.Visible {
			visible = append(visible, col)
		}
	}

	// 優先順位でソート
	sort.Slice(visible, func(i, j int) bool {
		return visible[i].Priority < visible[j].Priority
	})

	return visible
}

// SetColumnVisibility はカラムの表示/非表示を設定する
func (cm *ColumnManager) SetColumnVisibility(columnName string, visible bool) error {
	for i, col := range cm.columns {
		if col.Name == columnName {
			cm.columns[i].Visible = visible
			return nil
		}
	}
	return fmt.Errorf("column not found: %s", columnName)
}

// ParseColumnsString はカラム文字列を解析してカラム設定を更新する
func (cm *ColumnManager) ParseColumnsString(columnsStr string) error {
	if columnsStr == "" {
		return nil
	}

	// すべてのカラムを非表示にする
	for i := range cm.columns {
		cm.columns[i].Visible = false
	}

	// 指定されたカラムを表示する
	columns := strings.Split(columnsStr, ",")
	for _, colName := range columns {
		colName = strings.TrimSpace(colName)
		if colName == "" {
			continue
		}

		if err := cm.SetColumnVisibility(colName, true); err != nil {
			return err
		}
	}

	return nil
}

// GetColumnValue はモデルからカラム値を取得する
func (cm *ColumnManager) GetColumnValue(model model.Model, columnName string) (interface{}, error) {
	switch columnName {
	case "name":
		return model.Name, nil
	case "max_tokens":
		return model.MaxTokens, nil
	case "mode":
		return model.Mode, nil
	case "input_cost":
		return model.InputCost, nil
	default:
		return nil, fmt.Errorf("unknown column: %s", columnName)
	}
}

// GetColumnNames は利用可能なカラム名のリストを返す
func (cm *ColumnManager) GetColumnNames() []string {
	var names []string
	for _, col := range cm.columns {
		names = append(names, col.Name)
	}
	return names
}

// ResetToDefaults はカラム設定をデフォルトにリセットする
func (cm *ColumnManager) ResetToDefaults() {
	for i := range cm.columns {
		switch cm.columns[i].Name {
		case "name", "max_tokens", "mode", "input_cost":
			cm.columns[i].Visible = true
		default:
			cm.columns[i].Visible = false
		}
	}
}
