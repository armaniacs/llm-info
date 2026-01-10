# PBI 4: 高度な表示機能 - 実装計画

## PBI概要と目的

**タイトル**: モデル情報の高度な表示とフィルタリング機能

**目的**: 特定の条件でモデル情報をフィルタリングし、目的に合ったモデルを素早く見つけられるようにすること。

**ビジネス価値**:
- 検索性向上：目的のモデルの迅速な発見
- 比較機能：モデルの特性比較
- カスタマイズ：ユーザー要件に合わせた表示

## 現状の課題

1. 現在はすべてのモデルを一覧表示するのみ
2. モデルの検索やフィルタリング機能がない
3. ソート機能がない
4. JSON出力形式に対応していない
5. 表示項目のカスタマイズができない

## 実装計画

### 1. 機能要件の定義

#### フィルタリング機能
- モデル名による部分一致検索
- 最大トークン数による範囲指定
- モード（chat/completion）による絞り込み
- 入力コストによる範囲指定

#### ソート機能
- モデル名の昇順/降順
- 最大トークン数の昇順/降順
- 入力コストの昇順/降順

#### 出力形式
- テーブル形式（既存）
- JSON形式（新規）
- CSV形式（将来的な拡張）

#### 表示カスタマイズ
- カラムの表示/非表示切り替え
- カラムの順序変更
- テーブル幅の自動調整

### 2. コード構造

#### 新規ファイル作成
- `internal/ui/filter.go` - フィルタリング機能
- `internal/ui/sort.go` - ソート機能
- `internal/ui/columns.go` - カラム管理機能
- `internal/ui/csv.go` - CSV出力機能（将来的な拡張）

#### 既存ファイル修正
- `internal/ui/table.go` - テーブル表示機能の拡張
- `internal/ui/json.go` - JSON出力機能の拡張
- `cmd/llm-info/main.go` - コマンドライン引数の追加

### 3. 詳細実装

#### internal/ui/filter.go
```go
package ui

import (
    "regexp"
    "strconv"
    "strings"
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
func Filter(models []ModelInfo, criteria *FilterCriteria) []ModelInfo {
    if criteria == nil {
        return models
    }
    
    var filtered []ModelInfo
    for _, model := range models {
        if matchesCriteria(model, criteria) {
            filtered = append(filtered, model)
        }
    }
    
    return filtered
}

// matchesCriteria はモデルがフィルタ条件に一致するかチェックする
func matchesCriteria(model ModelInfo, criteria *FilterCriteria) bool {
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
```

#### internal/ui/sort.go
```go
package ui

import (
    "sort"
    "strings"
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
func Sort(models []ModelInfo, criteria *SortCriteria) {
    if criteria == nil {
        // デフォルトは名前の昇順
        criteria = &SortCriteria{Field: SortByName, Order: Ascending}
    }
    
    sort.Slice(models, func(i, j int) bool {
        return compare(models[i], models[j], criteria)
    })
}

// compare は2つのモデルを比較する
func compare(a, b ModelInfo, criteria *SortCriteria) bool {
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
```

#### internal/ui/columns.go
```go
package ui

import (
    "strings"
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
func (cm *ColumnManager) GetColumnValue(model ModelInfo, columnName string) (interface{}, error) {
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
```

#### internal/ui/table.go（拡張）
```go
package ui

import (
    "fmt"
    "os"
    "strings"
    
    "github.com/olekukonko/tablewriter"
)

// TableRenderer はテーブル表示機能を提供する
type TableRenderer struct {
    columnManager *ColumnManager
    table         *tablewriter.Table
}

// NewTableRenderer は新しいテーブルレンダラーを作成する
func NewTableRenderer() *TableRenderer {
    return &TableRenderer{
        columnManager: NewColumnManager(),
    }
}

// Render はモデル情報をテーブル形式で表示する
func (tr *TableRenderer) Render(models []ModelInfo, options *RenderOptions) error {
    if options != nil && options.Columns != "" {
        if err := tr.columnManager.ParseColumnsString(options.Columns); err != nil {
            return fmt.Errorf("failed to parse columns: %w", err)
        }
    }
    
    // テーブルの設定
    tr.table = tablewriter.NewWriter(os.Stdout)
    tr.table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
    tr.table.SetAlignment(tablewriter.ALIGN_LEFT)
    tr.table.SetBorder(false)
    tr.table.SetHeaderLine(false)
    tr.table.SetRowSeparator("")
    tr.table.SetColumnSeparator("  ")
    tr.table.SetTablePadding("  ")
    
    // カラムヘッダーの設定
    visibleColumns := tr.columnManager.GetVisibleColumns()
    var headers []string
    for _, col := range visibleColumns {
        headers = append(headers, col.Header)
    }
    tr.table.SetHeader(headers)
    
    // データ行の追加
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
        }
        tr.table.Append(row)
    }
    
    // テーブルの表示
    tr.table.Render()
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
```

#### internal/ui/json.go（拡張）
```go
package ui

import (
    "encoding/json"
    "fmt"
    "os"
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
func (jr *JSONRenderer) Render(models []ModelInfo, options *RenderOptions) error {
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
```

### 4. テスト戦略

#### 単体テスト
- `internal/ui/filter_test.go` - フィルタリング機能のテスト
- `internal/ui/sort_test.go` - ソート機能のテスト
- `internal/ui/columns_test.go` - カラム管理機能のテスト

#### 統合テスト
- `test/integration/display_integration_test.go` - 表示機能の統合テスト

#### E2Eテスト
- `test/e2e/advanced_display_test.go` - 高度な表示機能のE2Eテスト

### 5. 必要なファイルの新規作成・修正

#### 新規作成ファイル
1. `internal/ui/filter.go`
2. `internal/ui/sort.go`
3. `internal/ui/columns.go`
4. `internal/ui/filter_test.go`
5. `internal/ui/sort_test.go`
6. `internal/ui/columns_test.go`
7. `test/integration/display_integration_test.go`
8. `test/e2e/advanced_display_test.go`

#### 修正ファイル
1. `internal/ui/table.go`
2. `internal/ui/json.go`
3. `cmd/llm-info/main.go`
4. `internal/ui/table_test.go`
5. `internal/ui/json_test.go`

### 6. 受け入れ基準チェックリスト

- [ ] モデル名によるフィルタリング機能
- [ ] 各項目でのソート機能
- [ ] JSON出力形式のサポート
- [ ] カラムの表示/非表示切り替え
- [ ] 複合フィルタ条件のサポート
- [ ] ソート順序の指定（昇順/降順）
- [ ] 単体テストカバレッジ80%以上
- [ ] 統合テストの実装
- [ ] E2Eテストの実装

### 7. 実装手順

1. **フィルタリング機能の実装**
   - `internal/ui/filter.go`の作成
   - フィルタ条件の解析と適用ロジック

2. **ソート機能の実装**
   - `internal/ui/sort.go`の作成
   - ソート条件の解析と適用ロジック

3. **カラム管理機能の実装**
   - `internal/ui/columns.go`の作成
   - カラムの表示/非表示制御

4. **テーブル表示機能の拡張**
   - `internal/ui/table.go`の修正
   - カラム管理機能の統合

5. **JSON出力機能の拡張**
   - `internal/ui/json.go`の修正
   - メタデータの追加

6. **コマンドラインインターフェースの修正**
   - `cmd/llm-info/main.go`の修正
   - 新しいオプションの追加

7. **テストの実装**
   - 単体テストの作成
   - 統合テストの作成
   - E2Eテストの作成

8. **ドキュメント更新**
   - フィルタリング構文のドキュメント化
   - 使用例の追加

### 8. リスクと対策

#### リスク
1. 複雑なフィルタ構文のパースエラー
2. 大量のモデルデータのパフォーマンス問題
3. テーブル表示のレイアウト崩れ

#### 対策
1. 堅牢なパーサーと詳細なエラーメッセージ
2. 効率的なフィルタリングアルゴリズム
3. レスポンシブなテーブルデザイン

### 9. 成功指標

- フィルタリング機能の正確性100%
- ソート機能の正確性100%
- JSON出力のフォーマット適合性100%
- テストカバレッジ80%以上

### 10. 使用例

#### 基本フィルタリング
```bash
# モデル名でフィルタリング
llm-info --filter "gpt"

# トークン数でフィルタリング
llm-info --filter "tokens>1000"

# 複合条件でフィルタリング
llm-info --filter "name:gpt,tokens>1000,mode:chat"
```

#### ソート
```bash
# トークン数で昇順ソート
llm-info --sort "tokens"

# トークン数で降順ソート
llm-info --sort "-tokens"

# コストでソート
llm-info --sort "cost"
```

#### カラム制御
```bash
# 特定のカラムのみ表示
llm-info --columns "name,max_tokens"

# JSON出力
llm-info --format json

# フィルタリングとJSON出力の組み合わせ
llm-info --filter "gpt" --format json