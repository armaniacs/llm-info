package config

import (
	"time"

	"github.com/armaniacs/llm-info/internal/logging"
)

// ProbeConfig contains configuration specific to probe commands
type ProbeConfig struct {
	Log    LogConfig    `yaml:"log" json:"log"`
	Result ResultConfig `yaml:"result" json:"result"`
}

// LogConfig contains configuration for probe logging
type LogConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled"`
	Dir             string        `yaml:"dir" json:"dir"`
	Format          string        `yaml:"format" json:"format"`
	IncludeHistory  bool          `yaml:"include_history" json:"include_history"`
	Compress        bool          `yaml:"compress" json:"compress"`
	Retention       time.Duration `yaml:"retention" json:"retention"`
}

// ResultConfig contains configuration for result storage
type ResultConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`
	Dir       string `yaml:"dir" json:"dir"`
	Overwrite bool   `yaml:"overwrite" json:"overwrite"`
}

// GetDefaultProbeConfig returns default probe configuration
func GetDefaultProbeConfig() ProbeConfig {
	logConfig := logging.GetDefaultLogConfig()

	return ProbeConfig{
		Log: LogConfig{
			Enabled:         logConfig.Enabled,
			Dir:             logConfig.Dir,
			Format:          logConfig.Format,
			IncludeHistory:  logConfig.IncludeHistory,
			Compress:        logConfig.Compress,
			Retention:       logConfig.Retention,
		},
		Result: ResultConfig{
			Enabled:   true,
			Dir:       "/Users/y-araki/.config/llm-info/estimates",
			Overwrite: false,
		},
	}
}

// ConvertToProbeLogConfig converts LogConfig to ProbeLogConfig
func (c *LogConfig) ConvertToProbeLogConfig() logging.ProbeLogConfig {
	return logging.ProbeLogConfig{
		Enabled:         c.Enabled,
		Dir:             c.Dir,
		Format:          c.Format,
		IncludeHistory:  c.IncludeHistory,
		Compress:        c.Compress,
		Retention:       c.Retention,
	}
}