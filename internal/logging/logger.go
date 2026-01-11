package logging

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ProbeLogger defines the interface for probe logging
type ProbeLogger interface {
	LogTrial(model, gateway string, probeType string, trial TrialLogEntry) error
	LogResult(model, gateway string, probeType string, result interface{}) error
	Close() error
}

// TrialLogEntry represents a single trial in the probing process
type TrialLogEntry struct {
	Index        int           `json:"index"`
	TokenCount   int           `json:"token_count"`
	Success      bool          `json:"success"`
	Message      string        `json:"message,omitempty"`
	Duration     time.Duration `json:"duration"`
	Timestamp    time.Time     `json:"timestamp"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// ProbeLogEntry represents a complete probe log entry
type ProbeLogEntry struct {
	Timestamp   time.Time         `json:"timestamp"`
	Model       string           `json:"model"`
	Gateway     string           `json:"gateway"`
	ProbeType   string           `json:"probe_type"` // "context" or "max_output"
	Trial       TrialLogEntry    `json:"trial"`
	Metadata    ProbeLogMetadata `json:"metadata"`
}

// ProbeLogMetadata contains metadata about the probe session
type ProbeLogMetadata struct {
	Version      string `json:"version"`
	ExecutionID  string `json:"execution_id"`
	SessionIndex int    `json:"session_index"`
}

// ProbeLogConfig contains configuration for probe logging
type ProbeLogConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled"`
	Dir             string        `yaml:"dir" json:"dir"`
	Format          string        `yaml:"format" json:"format"`
	IncludeHistory  bool          `yaml:"include_history" json:"include_history"`
	Compress        bool          `yaml:"compress" json:"compress"`
	Retention       time.Duration `yaml:"retention" json:"retention"`
}

// JSONProbeLogger implements ProbeLogger with JSON Lines format
type JSONProbeLogger struct {
	config      ProbeLogConfig
	logFile     *os.File
	executionID string
	sessionIdx  int
}

// NewProbeLogger creates a new probe logger based on configuration
func NewProbeLogger(config ProbeLogConfig) (ProbeLogger, error) {
	if !config.Enabled {
		return &NoOpLogger{}, nil
	}

	// Expand ~ to home directory
	if len(config.Dir) > 0 && config.Dir[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		config.Dir = filepath.Join(home, config.Dir[1:])
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(config.Dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logger := &JSONProbeLogger{
		config:      config,
		executionID: generateUUID(),
		sessionIdx:  0,
	}

	return logger, nil
}

// LogTrial logs a single trial attempt
func (l *JSONProbeLogger) LogTrial(model, gateway string, probeType string, trial TrialLogEntry) error {
	// Generate log file name
	date := time.Now().Format("2006-01-02")
	logFileName := fmt.Sprintf("%s-%s-%s.log", date, sanitizeModelName(model), probeType)
	logFilePath := filepath.Join(l.config.Dir, logFileName)

	// Open log file in append mode
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Create log entry
	entry := ProbeLogEntry{
		Timestamp: time.Now(),
		Model:     model,
		Gateway:   gateway,
		ProbeType: probeType,
		Trial:     trial,
		Metadata: ProbeLogMetadata{
			Version:      "2.1.0",
			ExecutionID:  l.executionID,
			SessionIndex: l.sessionIdx,
		},
	}

	// Encode as JSON
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Write to file with newline
	if _, err := file.Write(append(entryJSON, '\n')); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	return nil
}

// LogResult logs the final probe result
func (l *JSONProbeLogger) LogResult(model, gateway string, probeType string, result interface{}) error {
	// Create a special result entry
	resultEntry := TrialLogEntry{
		Index:     -1, // Special index for results
		Success:   true,
		Message:   "Probe completed",
		Timestamp: time.Now(),
	}

	// Add result data as JSON string
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	resultEntry.Message = string(resultJSON)

	return l.LogTrial(model, gateway, probeType, resultEntry)
}

// Close finalizes the logger
func (l *JSONProbeLogger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// NoOpLogger provides a no-operation logger for when logging is disabled
type NoOpLogger struct{}

func (l *NoOpLogger) LogTrial(model, gateway string, probeType string, trial TrialLogEntry) error {
	return nil
}

func (l *NoOpLogger) LogResult(model, gateway string, probeType string, result interface{}) error {
	return nil
}

func (l *NoOpLogger) Close() error {
	return nil
}

// sanitizeModelName sanitizes model name for file system
func sanitizeModelName(model string) string {
	// Replace slashes and other problematic characters
	sanitized := model
	replacements := map[string]string{
		"/": "-",
		"\\": "-",
		":": "-",
		"*": "-",
		"?": "-",
		"\"": "",
		"<": "-",
		">": "-",
		"|": "-",
	}

	for old, new := range replacements {
		sanitized = replaceAll(sanitized, old, new)
	}

	return sanitized
}

// replaceAll is a simple string replacement helper
func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old) - 1
		} else {
			result += string(s[i])
		}
	}
	return result
}

// generateUUID generates a simple UUID v4 equivalent
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// GetDefaultLogConfig returns default logging configuration
func GetDefaultLogConfig() ProbeLogConfig {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/tmp"
	}

	return ProbeLogConfig{
		Enabled:         true,
		Dir:             filepath.Join(home, ".config", "llm-info", "log"),
		Format:          "json",
		IncludeHistory:  true,
		Compress:        false,
		Retention:       30 * 24 * time.Hour, // 30 days
	}
}