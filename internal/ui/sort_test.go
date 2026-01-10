package ui

import (
	"testing"

	"github.com/armaniacs/llm-info/internal/model"
)

func TestSort(t *testing.T) {
	// テスト用のモデルデータ
	models := []model.Model{
		{Name: "gpt-4", MaxTokens: 8192, Mode: "chat", InputCost: 0.00003},
		{Name: "gpt-3.5-turbo", MaxTokens: 4096, Mode: "chat", InputCost: 0.000002},
		{Name: "claude-3-opus", MaxTokens: 200000, Mode: "chat", InputCost: 0.000015},
		{Name: "text-davinci-003", MaxTokens: 4096, Mode: "completion", InputCost: 0.00002},
		{Name: "gemini-1.5-pro", MaxTokens: 1000000, Mode: "chat", InputCost: 0.0},
	}

	tests := []struct {
		name     string
		criteria *SortCriteria
		expected []string // 期待される順序のモデル名
	}{
		{
			name:     "nil criteria sorts by name ascending",
			criteria: nil,
			expected: []string{"claude-3-opus", "gemini-1.5-pro", "gpt-3.5-turbo", "gpt-4", "text-davinci-003"},
		},
		{
			name: "sort by name ascending",
			criteria: &SortCriteria{
				Field: SortByName,
				Order: Ascending,
			},
			expected: []string{"claude-3-opus", "gemini-1.5-pro", "gpt-3.5-turbo", "gpt-4", "text-davinci-003"},
		},
		{
			name: "sort by name descending",
			criteria: &SortCriteria{
				Field: SortByName,
				Order: Descending,
			},
			expected: []string{"text-davinci-003", "gpt-4", "gpt-3.5-turbo", "gemini-1.5-pro", "claude-3-opus"},
		},
		{
			name: "sort by max tokens ascending",
			criteria: &SortCriteria{
				Field: SortByMaxTokens,
				Order: Ascending,
			},
			expected: []string{"gpt-3.5-turbo", "text-davinci-003", "gpt-4", "claude-3-opus", "gemini-1.5-pro"},
		},
		{
			name: "sort by max tokens descending",
			criteria: &SortCriteria{
				Field: SortByMaxTokens,
				Order: Descending,
			},
			expected: []string{"gemini-1.5-pro", "claude-3-opus", "gpt-4", "text-davinci-003", "gpt-3.5-turbo"},
		},
		{
			name: "sort by input cost ascending",
			criteria: &SortCriteria{
				Field: SortByInputCost,
				Order: Ascending,
			},
			expected: []string{"gemini-1.5-pro", "gpt-3.5-turbo", "claude-3-opus", "text-davinci-003", "gpt-4"},
		},
		{
			name: "sort by input cost descending",
			criteria: &SortCriteria{
				Field: SortByInputCost,
				Order: Descending,
			},
			expected: []string{"gpt-4", "text-davinci-003", "claude-3-opus", "gpt-3.5-turbo", "gemini-1.5-pro"},
		},
		{
			name: "sort by mode ascending",
			criteria: &SortCriteria{
				Field: SortByMode,
				Order: Ascending,
			},
			expected: []string{"gpt-4", "gpt-3.5-turbo", "claude-3-opus", "gemini-1.5-pro", "text-davinci-003"},
		},
		{
			name: "sort by mode descending",
			criteria: &SortCriteria{
				Field: SortByMode,
				Order: Descending,
			},
			expected: []string{"text-davinci-003", "gemini-1.5-pro", "claude-3-opus", "gpt-3.5-turbo", "gpt-4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用にコピーを作成
			testModels := make([]model.Model, len(models))
			copy(testModels, models)

			Sort(testModels, tt.criteria)

			if len(testModels) != len(tt.expected) {
				t.Errorf("Sort() length = %v, want %v", len(testModels), len(tt.expected))
				return
			}

			for i, expectedName := range tt.expected {
				if testModels[i].Name != expectedName {
					t.Errorf("Sort()[%d] = %v, want %v", i, testModels[i].Name, expectedName)
				}
			}
		})
	}
}

func TestParseSortString(t *testing.T) {
	tests := []struct {
		name    string
		sortStr string
		want    *SortCriteria
		wantErr bool
	}{
		{
			name:    "empty string",
			sortStr: "",
			want: &SortCriteria{
				Field: SortByName,
				Order: Ascending,
			},
			wantErr: false,
		},
		{
			name:    "name ascending",
			sortStr: "name",
			want: &SortCriteria{
				Field: SortByName,
				Order: Ascending,
			},
			wantErr: false,
		},
		{
			name:    "name descending",
			sortStr: "-name",
			want: &SortCriteria{
				Field: SortByName,
				Order: Descending,
			},
			wantErr: false,
		},
		{
			name:    "model alias",
			sortStr: "model",
			want: &SortCriteria{
				Field: SortByName,
				Order: Ascending,
			},
			wantErr: false,
		},
		{
			name:    "tokens ascending",
			sortStr: "tokens",
			want: &SortCriteria{
				Field: SortByMaxTokens,
				Order: Ascending,
			},
			wantErr: false,
		},
		{
			name:    "tokens descending",
			sortStr: "-tokens",
			want: &SortCriteria{
				Field: SortByMaxTokens,
				Order: Descending,
			},
			wantErr: false,
		},
		{
			name:    "max_tokens alias",
			sortStr: "max_tokens",
			want: &SortCriteria{
				Field: SortByMaxTokens,
				Order: Ascending,
			},
			wantErr: false,
		},
		{
			name:    "cost ascending",
			sortStr: "cost",
			want: &SortCriteria{
				Field: SortByInputCost,
				Order: Ascending,
			},
			wantErr: false,
		},
		{
			name:    "cost descending",
			sortStr: "-cost",
			want: &SortCriteria{
				Field: SortByInputCost,
				Order: Descending,
			},
			wantErr: false,
		},
		{
			name:    "input_cost alias",
			sortStr: "input_cost",
			want: &SortCriteria{
				Field: SortByInputCost,
				Order: Ascending,
			},
			wantErr: false,
		},
		{
			name:    "mode ascending",
			sortStr: "mode",
			want: &SortCriteria{
				Field: SortByMode,
				Order: Ascending,
			},
			wantErr: false,
		},
		{
			name:    "mode descending",
			sortStr: "-mode",
			want: &SortCriteria{
				Field: SortByMode,
				Order: Descending,
			},
			wantErr: false,
		},
		{
			name:    "invalid field",
			sortStr: "invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "case insensitive",
			sortStr: "NAME",
			want: &SortCriteria{
				Field: SortByName,
				Order: Ascending,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSortString(tt.sortStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSortString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Field != tt.want.Field {
					t.Errorf("ParseSortString() Field = %v, want %v", got.Field, tt.want.Field)
				}
				if got.Order != tt.want.Order {
					t.Errorf("ParseSortString() Order = %v, want %v", got.Order, tt.want.Order)
				}
			}
		})
	}
}

func TestCompare(t *testing.T) {
	modelA := model.Model{
		Name:      "a-model",
		MaxTokens: 1000,
		Mode:      "chat",
		InputCost: 0.001,
	}
	modelB := model.Model{
		Name:      "b-model",
		MaxTokens: 2000,
		Mode:      "completion",
		InputCost: 0.002,
	}

	tests := []struct {
		name     string
		a        model.Model
		b        model.Model
		criteria *SortCriteria
		want     bool
	}{
		{
			name: "name ascending",
			a:    modelA,
			b:    modelB,
			criteria: &SortCriteria{
				Field: SortByName,
				Order: Ascending,
			},
			want: true,
		},
		{
			name: "name descending",
			a:    modelA,
			b:    modelB,
			criteria: &SortCriteria{
				Field: SortByName,
				Order: Descending,
			},
			want: false,
		},
		{
			name: "max tokens ascending",
			a:    modelA,
			b:    modelB,
			criteria: &SortCriteria{
				Field: SortByMaxTokens,
				Order: Ascending,
			},
			want: true,
		},
		{
			name: "max tokens descending",
			a:    modelA,
			b:    modelB,
			criteria: &SortCriteria{
				Field: SortByMaxTokens,
				Order: Descending,
			},
			want: false,
		},
		{
			name: "input cost ascending",
			a:    modelA,
			b:    modelB,
			criteria: &SortCriteria{
				Field: SortByInputCost,
				Order: Ascending,
			},
			want: true,
		},
		{
			name: "input cost descending",
			a:    modelA,
			b:    modelB,
			criteria: &SortCriteria{
				Field: SortByInputCost,
				Order: Descending,
			},
			want: false,
		},
		{
			name: "mode ascending",
			a:    modelA,
			b:    modelB,
			criteria: &SortCriteria{
				Field: SortByMode,
				Order: Ascending,
			},
			want: true,
		},
		{
			name: "mode descending",
			a:    modelA,
			b:    modelB,
			criteria: &SortCriteria{
				Field: SortByMode,
				Order: Descending,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compare(tt.a, tt.b, tt.criteria)
			if got != tt.want {
				t.Errorf("compare() = %v, want %v", got, tt.want)
			}
		})
	}
}
