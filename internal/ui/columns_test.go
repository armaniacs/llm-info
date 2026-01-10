package ui

import (
	"testing"

	"github.com/your-org/llm-info/internal/model"
)

func TestNewColumnManager(t *testing.T) {
	cm := NewColumnManager()
	if cm == nil {
		t.Fatal("NewColumnManager() returned nil")
	}

	if len(cm.columns) != 4 {
		t.Errorf("NewColumnManager() created %d columns, want 4", len(cm.columns))
	}

	// デフォルトですべてのカラムが表示されていることを確認
	for _, col := range cm.columns {
		if !col.Visible {
			t.Errorf("NewColumnManager() column %s is not visible by default", col.Name)
		}
	}
}

func TestGetVisibleColumns(t *testing.T) {
	cm := NewColumnManager()

	// すべてのカラムが表示されていることを確認
	visible := cm.GetVisibleColumns()
	if len(visible) != 4 {
		t.Errorf("GetVisibleColumns() returned %d columns, want 4", len(visible))
	}

	// 優先順位でソートされていることを確認
	for i := 1; i < len(visible); i++ {
		if visible[i-1].Priority > visible[i].Priority {
			t.Errorf("GetVisibleColumns() columns not sorted by priority: %d > %d", visible[i-1].Priority, visible[i].Priority)
		}
	}

	// 一部のカラムを非表示にして確認
	cm.SetColumnVisibility("max_tokens", false)
	visible = cm.GetVisibleColumns()
	if len(visible) != 3 {
		t.Errorf("GetVisibleColumns() returned %d columns after hiding one, want 3", len(visible))
	}

	// 非表示にしたカラムが含まれていないことを確認
	for _, col := range visible {
		if col.Name == "max_tokens" {
			t.Error("GetVisibleColumns() included hidden column 'max_tokens'")
		}
	}
}

func TestSetColumnVisibility(t *testing.T) {
	cm := NewColumnManager()

	// 存在するカラムを非表示にする
	err := cm.SetColumnVisibility("name", false)
	if err != nil {
		t.Errorf("SetColumnVisibility() error = %v", err)
	}

	// カラムが非表示になっていることを確認
	for _, col := range cm.columns {
		if col.Name == "name" && col.Visible {
			t.Error("SetColumnVisibility() did not hide column 'name'")
		}
	}

	// カラムを再表示する
	err = cm.SetColumnVisibility("name", true)
	if err != nil {
		t.Errorf("SetColumnVisibility() error = %v", err)
	}

	// カラムが表示されていることを確認
	for _, col := range cm.columns {
		if col.Name == "name" && !col.Visible {
			t.Error("SetColumnVisibility() did not show column 'name'")
		}
	}

	// 存在しないカラムを操作しようとする
	err = cm.SetColumnVisibility("nonexistent", false)
	if err == nil {
		t.Error("SetColumnVisibility() should return error for nonexistent column")
	}
}

func TestParseColumnsString(t *testing.T) {
	cm := NewColumnManager()

	tests := []struct {
		name       string
		columnsStr string
		wantErr    bool
		expected   []string
	}{
		{
			name:       "empty string",
			columnsStr: "",
			wantErr:    false,
			expected:   []string{"name", "max_tokens", "mode", "input_cost"}, // デフォルトのまま
		},
		{
			name:       "single column",
			columnsStr: "name",
			wantErr:    false,
			expected:   []string{"name"},
		},
		{
			name:       "multiple columns",
			columnsStr: "name,max_tokens",
			wantErr:    false,
			expected:   []string{"name", "max_tokens"},
		},
		{
			name:       "columns with spaces",
			columnsStr: "name, max_tokens , mode",
			wantErr:    false,
			expected:   []string{"name", "max_tokens", "mode"},
		},
		{
			name:       "all columns",
			columnsStr: "name,max_tokens,mode,input_cost",
			wantErr:    false,
			expected:   []string{"name", "max_tokens", "mode", "input_cost"},
		},
		{
			name:       "nonexistent column",
			columnsStr: "name,nonexistent",
			wantErr:    true,
			expected:   nil,
		},
		{
			name:       "empty parts",
			columnsStr: "name,,mode",
			wantErr:    false,
			expected:   []string{"name", "mode"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト前にリセット
			cm.ResetToDefaults()

			err := cm.ParseColumnsString(tt.columnsStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseColumnsString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				visible := cm.GetVisibleColumns()
				if len(visible) != len(tt.expected) {
					t.Errorf("ParseColumnsString() visible columns count = %v, want %v", len(visible), len(tt.expected))
					return
				}

				for i, expectedName := range tt.expected {
					if i >= len(visible) || visible[i].Name != expectedName {
						t.Errorf("ParseColumnsString() visible column[%d] = %v, want %v", i, visible[i].Name, expectedName)
					}
				}
			}
		})
	}
}

func TestGetColumnValue(t *testing.T) {
	cm := NewColumnManager()
	model := model.Model{
		Name:      "gpt-4",
		MaxTokens: 8192,
		Mode:      "chat",
		InputCost: 0.00003,
	}

	tests := []struct {
		name        string
		columnName  string
		expected    interface{}
		expectError bool
	}{
		{
			name:        "name column",
			columnName:  "name",
			expected:    "gpt-4",
			expectError: false,
		},
		{
			name:        "max_tokens column",
			columnName:  "max_tokens",
			expected:    8192,
			expectError: false,
		},
		{
			name:        "mode column",
			columnName:  "mode",
			expected:    "chat",
			expectError: false,
		},
		{
			name:        "input_cost column",
			columnName:  "input_cost",
			expected:    0.00003,
			expectError: false,
		},
		{
			name:        "nonexistent column",
			columnName:  "nonexistent",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := cm.GetColumnValue(model, tt.columnName)
			if (err != nil) != tt.expectError {
				t.Errorf("GetColumnValue() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				if value != tt.expected {
					t.Errorf("GetColumnValue() = %v, want %v", value, tt.expected)
				}
			}
		})
	}
}

func TestGetColumnNames(t *testing.T) {
	cm := NewColumnManager()
	names := cm.GetColumnNames()

	expected := []string{"name", "max_tokens", "mode", "input_cost"}
	if len(names) != len(expected) {
		t.Errorf("GetColumnNames() returned %d names, want %d", len(names), len(expected))
	}

	for i, expectedName := range expected {
		if i >= len(names) || names[i] != expectedName {
			t.Errorf("GetColumnNames()[%d] = %v, want %v", i, names[i], expectedName)
		}
	}
}

func TestResetToDefaults(t *testing.T) {
	cm := NewColumnManager()

	// 一部のカラムを非表示にする
	cm.SetColumnVisibility("name", false)
	cm.SetColumnVisibility("max_tokens", false)

	// リセットする
	cm.ResetToDefaults()

	// すべてのカラムが表示されていることを確認
	visible := cm.GetVisibleColumns()
	if len(visible) != 4 {
		t.Errorf("ResetToDefaults() visible columns count = %v, want 4", len(visible))
	}

	for _, col := range cm.columns {
		if !col.Visible {
			t.Errorf("ResetToDefaults() column %s is not visible", col.Name)
		}
	}
}

func TestColumnProperties(t *testing.T) {
	cm := NewColumnManager()

	// 各カラムのプロパティを確認
	expectedColumns := []struct {
		name     string
		header   string
		width    int
		format   string
		priority int
	}{
		{"name", "MODEL NAME", 30, "%s", 1},
		{"max_tokens", "MAX TOKENS", 12, "%d", 2},
		{"mode", "MODE", 8, "%s", 3},
		{"input_cost", "INPUT COST", 12, "%.6f", 4},
	}

	for _, expected := range expectedColumns {
		var found *Column
		for _, col := range cm.columns {
			if col.Name == expected.name {
				found = &col
				break
			}
		}

		if found == nil {
			t.Errorf("Column %s not found", expected.name)
			continue
		}

		if found.Header != expected.header {
			t.Errorf("Column %s header = %v, want %v", expected.name, found.Header, expected.header)
		}
		if found.Width != expected.width {
			t.Errorf("Column %s width = %v, want %v", expected.name, found.Width, expected.width)
		}
		if found.Format != expected.format {
			t.Errorf("Column %s format = %v, want %v", expected.name, found.Format, expected.format)
		}
		if found.Priority != expected.priority {
			t.Errorf("Column %s priority = %v, want %v", expected.name, found.Priority, expected.priority)
		}
	}
}
