package ui

import (
	"testing"

	"github.com/armaniacs/llm-info/internal/model"
)

func TestFilter(t *testing.T) {
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
		criteria *FilterCriteria
		expected int
	}{
		{
			name:     "nil criteria returns all models",
			criteria: nil,
			expected: 5,
		},
		{
			name: "filter by name pattern",
			criteria: &FilterCriteria{
				NamePattern: "gpt",
			},
			expected: 2,
		},
		{
			name: "filter by exclude pattern",
			criteria: &FilterCriteria{
				ExcludePattern: "gpt",
			},
			expected: 3,
		},
		{
			name: "filter by min tokens",
			criteria: &FilterCriteria{
				MinTokens: 10000,
			},
			expected: 2,
		},
		{
			name: "filter by max tokens",
			criteria: &FilterCriteria{
				MaxTokens: 5000,
			},
			expected: 2,
		},
		{
			name: "filter by token range",
			criteria: &FilterCriteria{
				MinTokens: 4000,
				MaxTokens: 10000,
			},
			expected: 3,
		},
		{
			name: "filter by mode",
			criteria: &FilterCriteria{
				Modes: []string{"chat"},
			},
			expected: 4,
		},
		{
			name: "filter by min cost",
			criteria: &FilterCriteria{
				MinInputCost: 0.00001,
			},
			expected: 3,
		},
		{
			name: "filter by max cost",
			criteria: &FilterCriteria{
				MaxInputCost: 0.00001,
			},
			expected: 2,
		},
		{
			name: "complex filter",
			criteria: &FilterCriteria{
				NamePattern:  "gpt",
				MinTokens:    4000,
				Modes:        []string{"chat"},
				MaxInputCost: 0.0001,
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Filter(models, tt.criteria)
			if len(result) != tt.expected {
				t.Errorf("Filter() = %v, want %v", len(result), tt.expected)
			}
		})
	}
}

func TestParseFilterString(t *testing.T) {
	tests := []struct {
		name      string
		filterStr string
		want      *FilterCriteria
		wantErr   bool
	}{
		{
			name:      "empty string",
			filterStr: "",
			want:      nil,
			wantErr:   false,
		},
		{
			name:      "simple name filter",
			filterStr: "gpt",
			want: &FilterCriteria{
				NamePattern: "gpt",
			},
			wantErr: false,
		},
		{
			name:      "name filter with prefix",
			filterStr: "name:gpt",
			want: &FilterCriteria{
				NamePattern: "gpt",
			},
			wantErr: false,
		},
		{
			name:      "exclude filter",
			filterStr: "exclude:beta",
			want: &FilterCriteria{
				ExcludePattern: "beta",
			},
			wantErr: false,
		},
		{
			name:      "tokens greater than filter",
			filterStr: "tokens>1000",
			want: &FilterCriteria{
				MinTokens: 1000,
			},
			wantErr: false,
		},
		{
			name:      "tokens less than filter",
			filterStr: "tokens<100000",
			want: &FilterCriteria{
				MaxTokens: 100000,
			},
			wantErr: false,
		},
		{
			name:      "cost greater than filter",
			filterStr: "cost>0.001",
			want: &FilterCriteria{
				MinInputCost: 0.001,
			},
			wantErr: false,
		},
		{
			name:      "cost less than filter",
			filterStr: "cost<0.01",
			want: &FilterCriteria{
				MaxInputCost: 0.01,
			},
			wantErr: false,
		},
		{
			name:      "mode filter",
			filterStr: "mode:chat",
			want: &FilterCriteria{
				Modes: []string{"chat"},
			},
			wantErr: false,
		},
		{
			name:      "complex filter",
			filterStr: "name:gpt,tokens>1000,mode:chat,cost<0.001",
			want: &FilterCriteria{
				NamePattern:  "gpt",
				MinTokens:    1000,
				Modes:        []string{"chat"},
				MaxInputCost: 0.001,
			},
			wantErr: false,
		},
		{
			name:      "invalid token filter",
			filterStr: "tokens>abc",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "invalid cost filter",
			filterStr: "cost>xyz",
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFilterString(tt.filterStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFilterString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.want == nil && got != nil {
					t.Errorf("ParseFilterString() expected nil result, got %v", got)
					return
				}
				if tt.want != nil && got == nil {
					t.Errorf("ParseFilterString() expected non-nil result, got nil")
					return
				}
				if tt.want == nil {
					// nilが期待される場合は、これ以上のチェックは不要
					return
				}
				if got.NamePattern != tt.want.NamePattern {
					t.Errorf("ParseFilterString() NamePattern = %v, want %v", got.NamePattern, tt.want.NamePattern)
				}
				if got.ExcludePattern != tt.want.ExcludePattern {
					t.Errorf("ParseFilterString() ExcludePattern = %v, want %v", got.ExcludePattern, tt.want.ExcludePattern)
				}
				if got.MinTokens != tt.want.MinTokens {
					t.Errorf("ParseFilterString() MinTokens = %v, want %v", got.MinTokens, tt.want.MinTokens)
				}
				if got.MaxTokens != tt.want.MaxTokens {
					t.Errorf("ParseFilterString() MaxTokens = %v, want %v", got.MaxTokens, tt.want.MaxTokens)
				}
				if got.MinInputCost != tt.want.MinInputCost {
					t.Errorf("ParseFilterString() MinInputCost = %v, want %v", got.MinInputCost, tt.want.MinInputCost)
				}
				if got.MaxInputCost != tt.want.MaxInputCost {
					t.Errorf("ParseFilterString() MaxInputCost = %v, want %v", got.MaxInputCost, tt.want.MaxInputCost)
				}
				if len(got.Modes) != len(tt.want.Modes) {
					t.Errorf("ParseFilterString() Modes length = %v, want %v", len(got.Modes), len(tt.want.Modes))
				} else {
					for i, mode := range got.Modes {
						if mode != tt.want.Modes[i] {
							t.Errorf("ParseFilterString() Modes[%d] = %v, want %v", i, mode, tt.want.Modes[i])
						}
					}
				}
			}
		})
	}
}

func TestMatchesCriteria(t *testing.T) {
	model := model.Model{
		Name:      "gpt-4",
		MaxTokens: 8192,
		Mode:      "chat",
		InputCost: 0.00003,
	}

	tests := []struct {
		name     string
		criteria *FilterCriteria
		want     bool
	}{
		{
			name:     "nil criteria matches",
			criteria: nil,
			want:     true,
		},
		{
			name: "name pattern matches",
			criteria: &FilterCriteria{
				NamePattern: "gpt",
			},
			want: true,
		},
		{
			name: "name pattern does not match",
			criteria: &FilterCriteria{
				NamePattern: "claude",
			},
			want: false,
		},
		{
			name: "exclude pattern matches",
			criteria: &FilterCriteria{
				ExcludePattern: "gpt",
			},
			want: false,
		},
		{
			name: "exclude pattern does not match",
			criteria: &FilterCriteria{
				ExcludePattern: "claude",
			},
			want: true,
		},
		{
			name: "min tokens matches",
			criteria: &FilterCriteria{
				MinTokens: 4000,
			},
			want: true,
		},
		{
			name: "min tokens does not match",
			criteria: &FilterCriteria{
				MinTokens: 10000,
			},
			want: false,
		},
		{
			name: "max tokens matches",
			criteria: &FilterCriteria{
				MaxTokens: 10000,
			},
			want: true,
		},
		{
			name: "max tokens does not match",
			criteria: &FilterCriteria{
				MaxTokens: 4000,
			},
			want: false,
		},
		{
			name: "mode matches",
			criteria: &FilterCriteria{
				Modes: []string{"chat"},
			},
			want: true,
		},
		{
			name: "mode does not match",
			criteria: &FilterCriteria{
				Modes: []string{"completion"},
			},
			want: false,
		},
		{
			name: "min cost matches",
			criteria: &FilterCriteria{
				MinInputCost: 0.00001,
			},
			want: true,
		},
		{
			name: "min cost does not match",
			criteria: &FilterCriteria{
				MinInputCost: 0.0001,
			},
			want: false,
		},
		{
			name: "max cost matches",
			criteria: &FilterCriteria{
				MaxInputCost: 0.0001,
			},
			want: true,
		},
		{
			name: "max cost does not match",
			criteria: &FilterCriteria{
				MaxInputCost: 0.00001,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesCriteria(model, tt.criteria)
			if got != tt.want {
				t.Errorf("matchesCriteria() = %v, want %v", got, tt.want)
			}
		})
	}
}
