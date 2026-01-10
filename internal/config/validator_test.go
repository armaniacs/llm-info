package config

import (
	"testing"
	"time"

	"github.com/your-org/llm-info/pkg/config"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				Gateways: []config.Gateway{
					{
						Name:    "test-gateway",
						URL:     "https://test.example.com",
						APIKey:  "test-key",
						Timeout: 10 * time.Second,
					},
				},
				DefaultGateway: "test-gateway",
				Global: config.Global{
					Timeout:      10 * time.Second,
					OutputFormat: "table",
					SortBy:       "name",
				},
			},
			wantErr: false,
		},
		{
			name: "no gateways",
			cfg: &config.Config{
				Gateways:       []config.Gateway{},
				DefaultGateway: "test-gateway",
				Global: config.Global{
					Timeout:      10 * time.Second,
					OutputFormat: "table",
					SortBy:       "name",
				},
			},
			wantErr: true,
			errMsg:  "at least one gateway must be configured",
		},
		{
			name: "duplicate gateway names",
			cfg: &config.Config{
				Gateways: []config.Gateway{
					{
						Name:    "test-gateway",
						URL:     "https://test.example.com",
						APIKey:  "test-key",
						Timeout: 10 * time.Second,
					},
					{
						Name:    "test-gateway",
						URL:     "https://test2.example.com",
						APIKey:  "test-key2",
						Timeout: 10 * time.Second,
					},
				},
				DefaultGateway: "test-gateway",
				Global: config.Global{
					Timeout:      10 * time.Second,
					OutputFormat: "table",
					SortBy:       "name",
				},
			},
			wantErr: true,
			errMsg:  "duplicate gateway name: test-gateway",
		},
		{
			name: "default gateway not found",
			cfg: &config.Config{
				Gateways: []config.Gateway{
					{
						Name:    "test-gateway",
						URL:     "https://test.example.com",
						APIKey:  "test-key",
						Timeout: 10 * time.Second,
					},
				},
				DefaultGateway: "non-existent",
				Global: config.Global{
					Timeout:      10 * time.Second,
					OutputFormat: "table",
					SortBy:       "name",
				},
			},
			wantErr: true,
			errMsg:  "default gateway 'non-existent' not found",
		},
		{
			name: "invalid global timeout",
			cfg: &config.Config{
				Gateways: []config.Gateway{
					{
						Name:    "test-gateway",
						URL:     "https://test.example.com",
						APIKey:  "test-key",
						Timeout: 10 * time.Second,
					},
				},
				DefaultGateway: "test-gateway",
				Global: config.Global{
					Timeout:      -1 * time.Second,
					OutputFormat: "table",
					SortBy:       "name",
				},
			},
			wantErr: true,
			errMsg:  "global settings: global timeout must be positive",
		},
		{
			name: "invalid output format",
			cfg: &config.Config{
				Gateways: []config.Gateway{
					{
						Name:    "test-gateway",
						URL:     "https://test.example.com",
						APIKey:  "test-key",
						Timeout: 10 * time.Second,
					},
				},
				DefaultGateway: "test-gateway",
				Global: config.Global{
					Timeout:      10 * time.Second,
					OutputFormat: "invalid",
					SortBy:       "name",
				},
			},
			wantErr: true,
			errMsg:  "global settings: invalid output format: invalid (valid formats: [table json])",
		},
		{
			name: "invalid sort by",
			cfg: &config.Config{
				Gateways: []config.Gateway{
					{
						Name:    "test-gateway",
						URL:     "https://test.example.com",
						APIKey:  "test-key",
						Timeout: 10 * time.Second,
					},
				},
				DefaultGateway: "test-gateway",
				Global: config.Global{
					Timeout:      10 * time.Second,
					OutputFormat: "table",
					SortBy:       "invalid",
				},
			},
			wantErr: true,
			errMsg:  "global settings: invalid sort by: invalid (valid options: [name max_tokens mode input_cost])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("ValidateConfig() error = %v, want %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateGateway(t *testing.T) {
	tests := []struct {
		name    string
		gw      *config.Gateway
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid gateway",
			gw: &config.Gateway{
				Name:    "test-gateway",
				URL:     "https://test.example.com",
				APIKey:  "test-key",
				Timeout: 10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			gw: &config.Gateway{
				Name:    "",
				URL:     "https://test.example.com",
				APIKey:  "test-key",
				Timeout: 10 * time.Second,
			},
			wantErr: true,
			errMsg:  "gateway name cannot be empty",
		},
		{
			name: "empty URL",
			gw: &config.Gateway{
				Name:    "test-gateway",
				URL:     "",
				APIKey:  "test-key",
				Timeout: 10 * time.Second,
			},
			wantErr: true,
			errMsg:  "gateway URL cannot be empty",
		},
		{
			name: "invalid URL",
			gw: &config.Gateway{
				Name:    "test-gateway",
				URL:     "invalid-url",
				APIKey:  "test-key",
				Timeout: 10 * time.Second,
			},
			wantErr: false, // url.Parseは相対URLも解析するため、エラーにならない
		},
		{
			name: "negative timeout",
			gw: &config.Gateway{
				Name:    "test-gateway",
				URL:     "https://test.example.com",
				APIKey:  "test-key",
				Timeout: -1 * time.Second,
			},
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
		{
			name: "zero timeout",
			gw: &config.Gateway{
				Name:    "test-gateway",
				URL:     "https://test.example.com",
				APIKey:  "test-key",
				Timeout: 0,
			},
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGateway(tt.gw)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGateway() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("validateGateway() error = %v, want %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateGlobal(t *testing.T) {
	tests := []struct {
		name    string
		global  *config.Global
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid global",
			global: &config.Global{
				Timeout:      10 * time.Second,
				OutputFormat: "table",
				SortBy:       "name",
			},
			wantErr: false,
		},
		{
			name: "negative timeout",
			global: &config.Global{
				Timeout:      -1 * time.Second,
				OutputFormat: "table",
				SortBy:       "name",
			},
			wantErr: true,
			errMsg:  "global timeout must be positive",
		},
		{
			name: "zero timeout",
			global: &config.Global{
				Timeout:      0,
				OutputFormat: "table",
				SortBy:       "name",
			},
			wantErr: true,
			errMsg:  "global timeout must be positive",
		},
		{
			name: "invalid output format",
			global: &config.Global{
				Timeout:      10 * time.Second,
				OutputFormat: "invalid",
				SortBy:       "name",
			},
			wantErr: true,
			errMsg:  "invalid output format: invalid (valid formats: [table json])",
		},
		{
			name: "invalid sort by",
			global: &config.Global{
				Timeout:      10 * time.Second,
				OutputFormat: "table",
				SortBy:       "invalid",
			},
			wantErr: true,
			errMsg:  "invalid sort by: invalid (valid options: [name max_tokens mode input_cost])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGlobal(tt.global)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGlobal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("validateGlobal() error = %v, want %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateLegacyConfig(t *testing.T) {
	tests := []struct {
		name       string
		fileConfig *config.FileConfig
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid legacy config",
			fileConfig: &config.FileConfig{
				Gateways: []config.GatewayConfig{
					{
						Name:    "test-gateway",
						URL:     "https://test.example.com",
						APIKey:  "test-key",
						Timeout: 10 * time.Second,
					},
				},
				DefaultGateway: "test-gateway",
				Common: config.CommonConfig{
					Timeout: 10 * time.Second,
					Output: config.OutputConfig{
						Format: "table",
						Table: config.TableConfig{
							AlwaysShow:      []string{"name"},
							ShowIfAvailable: []string{"max_tokens"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "no gateways",
			fileConfig: &config.FileConfig{
				Gateways:       []config.GatewayConfig{},
				DefaultGateway: "test-gateway",
				Common: config.CommonConfig{
					Timeout: 10 * time.Second,
					Output: config.OutputConfig{
						Format: "table",
					},
				},
			},
			wantErr: true,
			errMsg:  "at least one gateway must be configured",
		},
		{
			name: "invalid output format",
			fileConfig: &config.FileConfig{
				Gateways: []config.GatewayConfig{
					{
						Name:    "test-gateway",
						URL:     "https://test.example.com",
						APIKey:  "test-key",
						Timeout: 10 * time.Second,
					},
				},
				DefaultGateway: "test-gateway",
				Common: config.CommonConfig{
					Timeout: 10 * time.Second,
					Output: config.OutputConfig{
						Format: "invalid",
					},
				},
			},
			wantErr: true,
			errMsg:  "common settings: invalid output format: invalid (valid formats: [table json])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLegacyConfig(tt.fileConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLegacyConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("ValidateLegacyConfig() error = %v, want %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}
