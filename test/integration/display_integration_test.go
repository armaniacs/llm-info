package integration

import (
	"testing"

	"github.com/armaniacs/llm-info/internal/model"
	"github.com/armaniacs/llm-info/internal/ui"
)

func TestDisplayIntegration(t *testing.T) {
	// テスト用のモデルデータ
	models := []model.Model{
		{Name: "gpt-4", MaxTokens: 8192, Mode: "chat", InputCost: 0.00003},
		{Name: "gpt-3.5-turbo", MaxTokens: 4096, Mode: "chat", InputCost: 0.000002},
		{Name: "claude-3-opus", MaxTokens: 200000, Mode: "chat", InputCost: 0.000015},
		{Name: "text-davinci-003", MaxTokens: 4096, Mode: "completion", InputCost: 0.00002},
		{Name: "gemini-1.5-pro", MaxTokens: 1000000, Mode: "chat", InputCost: 0.0},
	}

	t.Run("filter and sort integration", func(t *testing.T) {
		// フィルタリング
		filterCriteria, err := ui.ParseFilterString("name:gpt,mode:chat")
		if err != nil {
			t.Fatalf("ParseFilterString() error = %v", err)
		}

		filteredModels := ui.Filter(models, filterCriteria)
		if len(filteredModels) != 2 {
			t.Errorf("Filter() returned %d models, want 2", len(filteredModels))
		}

		// ソート
		sortCriteria, err := ui.ParseSortString("-max_tokens")
		if err != nil {
			t.Fatalf("ParseSortString() error = %v", err)
		}

		ui.Sort(filteredModels, sortCriteria)

		// ソート順序の確認
		if filteredModels[0].Name != "gpt-4" || filteredModels[1].Name != "gpt-3.5-turbo" {
			t.Errorf("Sort() order incorrect: got %v, want [gpt-4, gpt-3.5-turbo]",
				getModelNames(filteredModels))
		}
	})

	t.Run("table rendering with custom columns", func(t *testing.T) {
		// カスタムカラムでテーブル表示
		options := &ui.RenderOptions{
			Columns: "name,max_tokens",
			Filter:  "mode:chat",
			Sort:    "name",
		}

		// フィルタリング
		filterCriteria, err := ui.ParseFilterString(options.Filter)
		if err != nil {
			t.Fatalf("ParseFilterString() error = %v", err)
		}

		filteredModels := ui.Filter(models, filterCriteria)

		// ソート
		sortCriteria, err := ui.ParseSortString(options.Sort)
		if err != nil {
			t.Fatalf("ParseSortString() error = %v", err)
		}

		ui.Sort(filteredModels, sortCriteria)

		// テーブルレンダラーのテスト
		renderer := ui.NewTableRenderer()
		err = renderer.Render(filteredModels, options)
		if err != nil {
			t.Errorf("TableRenderer.Render() error = %v", err)
		}
	})

	t.Run("JSON rendering with metadata", func(t *testing.T) {
		// JSON出力のテスト
		options := &ui.RenderOptions{
			Filter: "tokens>10000",
		}

		// フィルタリング
		filterCriteria, err := ui.ParseFilterString(options.Filter)
		if err != nil {
			t.Fatalf("ParseFilterString() error = %v", err)
		}

		filteredModels := ui.Filter(models, filterCriteria)

		// JSONレンダラーのテスト
		renderer := ui.NewJSONRenderer(true)
		err = renderer.Render(filteredModels, options)
		if err != nil {
			t.Errorf("JSONRenderer.Render() error = %v", err)
		}
	})

	t.Run("complex filter and sort combination", func(t *testing.T) {
		// 複雑なフィルタとソートの組み合わせ
		options := &ui.RenderOptions{
			Filter:  "tokens>4000,cost<0.0001,mode:chat",
			Sort:    "max_tokens",
			Columns: "name,max_tokens,input_cost",
		}

		// フィルタリング
		filterCriteria, err := ui.ParseFilterString(options.Filter)
		if err != nil {
			t.Fatalf("ParseFilterString() error = %v", err)
		}

		filteredModels := ui.Filter(models, filterCriteria)
		expectedCount := 4 // gpt-4, gpt-3.5-turbo, claude-3-opus, gemini-1.5-pro (gpt-3.5-turbo has cost 0.000002 which is < 0.0001)
		if len(filteredModels) != expectedCount {
			t.Errorf("Filter() returned %d models, want %d", len(filteredModels), expectedCount)
		}

		// ソート
		sortCriteria, err := ui.ParseSortString(options.Sort)
		if err != nil {
			t.Fatalf("ParseSortString() error = %v", err)
		}

		ui.Sort(filteredModels, sortCriteria)

		// テーブル表示
		renderer := ui.NewTableRenderer()
		err = renderer.Render(filteredModels, options)
		if err != nil {
			t.Errorf("TableRenderer.Render() error = %v", err)
		}
	})

	t.Run("empty results handling", func(t *testing.T) {
		// 存在しないモデルでフィルタリング
		filterCriteria, err := ui.ParseFilterString("name:nonexistent")
		if err != nil {
			t.Fatalf("ParseFilterString() error = %v", err)
		}

		filteredModels := ui.Filter(models, filterCriteria)
		if len(filteredModels) != 0 {
			t.Errorf("Filter() returned %d models, want 0", len(filteredModels))
		}

		// 空の結果でのテーブル表示
		renderer := ui.NewTableRenderer()
		err = renderer.Render(filteredModels, nil)
		if err != nil {
			t.Errorf("TableRenderer.Render() with empty models error = %v", err)
		}

		// 空の結果でのJSON表示
		jsonRenderer := ui.NewJSONRenderer(true)
		err = jsonRenderer.Render(filteredModels, nil)
		if err != nil {
			t.Errorf("JSONRenderer.Render() with empty models error = %v", err)
		}
	})
}

func TestColumnManagerIntegration(t *testing.T) {
	cm := ui.NewColumnManager()

	t.Run("column visibility and rendering", func(t *testing.T) {
		// カラムの表示/非表示
		err := cm.SetColumnVisibility("input_cost", false)
		if err != nil {
			t.Errorf("SetColumnVisibility() error = %v", err)
		}

		visible := cm.GetVisibleColumns()
		for _, col := range visible {
			if col.Name == "input_cost" {
				t.Error("input_cost column should be hidden")
			}
		}

		// カラム文字列のパース
		err = cm.ParseColumnsString("name,mode")
		if err != nil {
			t.Errorf("ParseColumnsString() error = %v", err)
		}

		visible = cm.GetVisibleColumns()
		if len(visible) != 2 {
			t.Errorf("GetVisibleColumns() returned %d columns, want 2", len(visible))
		}

		expectedColumns := []string{"name", "mode"}
		for i, expected := range expectedColumns {
			if visible[i].Name != expected {
				t.Errorf("GetVisibleColumns()[%d] = %v, want %v", i, visible[i].Name, expected)
			}
		}
	})

	t.Run("column value extraction", func(t *testing.T) {
		model := model.Model{
			Name:      "test-model",
			MaxTokens: 4096,
			Mode:      "chat",
			InputCost: 0.001,
		}

		// すべてのカラム値の取得
		columns := cm.GetColumnNames()
		for _, colName := range columns {
			value, err := cm.GetColumnValue(model, colName)
			if err != nil {
				t.Errorf("GetColumnValue() error for %s = %v", colName, err)
			}

			if value == nil {
				t.Errorf("GetColumnValue() returned nil for %s", colName)
			}
		}
	})
}

func TestErrorHandlingIntegration(t *testing.T) {
	t.Run("invalid filter string", func(t *testing.T) {
		_, err := ui.ParseFilterString("tokens>abc")
		if err == nil {
			t.Error("ParseFilterString() should return error for invalid token value")
		}
	})

	t.Run("invalid sort string", func(t *testing.T) {
		_, err := ui.ParseSortString("invalid_field")
		if err == nil {
			t.Error("ParseSortString() should return error for invalid field")
		}
	})

	t.Run("invalid column name", func(t *testing.T) {
		cm := ui.NewColumnManager()
		err := cm.ParseColumnsString("name,invalid_column")
		if err == nil {
			t.Error("ParseColumnsString() should return error for invalid column")
		}
	})
}

// ヘルパー関数
func getModelNames(models []model.Model) []string {
	names := make([]string, len(models))
	for i, model := range models {
		names[i] = model.Name
	}
	return names
}
