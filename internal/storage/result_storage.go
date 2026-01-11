package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ResultStorage defines the interface for storing probe results
type ResultStorage interface {
	SaveContextResult(provider, model string, result interface{}) error
	SaveMaxOutputResult(provider, model string, result interface{}) error
	LoadContextResult(provider, model string) (interface{}, error)
	LoadMaxOutputResult(provider, model string) (interface{}, error)
}

// SavedResult represents the structure of saved probe results
type SavedResult struct {
	ContextWindow  interface{} `json:"context_window,omitempty"`
	MaxOutput      interface{} `json:"max_output,omitempty"`
	EstimatedAt    time.Time  `json:"estimated_at"`
	LLMInfoVersion string     `json:"llm_info_version"`
}

// JSONResultStorage implements ResultStorage with JSON files
type JSONResultStorage struct {
	baseDir string
}

// NewResultStorage creates a new result storage instance
func NewResultStorage(baseDir string) (ResultStorage, error) {
	// Expand ~ to home directory
	if len(baseDir) > 0 && baseDir[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(home, baseDir[1:])
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create result directory: %w", err)
	}

	return &JSONResultStorage{
		baseDir: baseDir,
	}, nil
}

// SaveContextResult saves a context window probe result
func (s *JSONResultStorage) SaveContextResult(provider, model string, result interface{}) error {
	fileName := fmt.Sprintf("%s-%s.json", sanitizeProviderName(provider), sanitizeModelName(model))
	filePath := filepath.Join(s.baseDir, fileName)

	// Try to load existing file
	var existing SavedResult
	if data, err := os.ReadFile(filePath); err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			// If unmarshal fails, start fresh
			existing = SavedResult{}
		}
	}

	// Update with new result
	existing.ContextWindow = result
	existing.EstimatedAt = time.Now()
	existing.LLMInfoVersion = "2.1.0"

	// Save to file
	return s.saveToFile(filePath, existing)
}

// SaveMaxOutputResult saves a max output probe result
func (s *JSONResultStorage) SaveMaxOutputResult(provider, model string, result interface{}) error {
	fileName := fmt.Sprintf("%s-%s.json", sanitizeProviderName(provider), sanitizeModelName(model))
	filePath := filepath.Join(s.baseDir, fileName)

	// Try to load existing file
	var existing SavedResult
	if data, err := os.ReadFile(filePath); err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			// If unmarshal fails, start fresh
			existing = SavedResult{}
		}
	}

	// Update with new result
	existing.MaxOutput = result
	existing.EstimatedAt = time.Now()
	existing.LLMInfoVersion = "2.1.0"

	// Save to file
	return s.saveToFile(filePath, existing)
}

// LoadContextResult loads a context window probe result
func (s *JSONResultStorage) LoadContextResult(provider, model string) (interface{}, error) {
	fileName := fmt.Sprintf("%s-%s.json", sanitizeProviderName(provider), sanitizeModelName(model))
	filePath := filepath.Join(s.baseDir, fileName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read result file: %w", err)
	}

	var result SavedResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	if result.ContextWindow == nil {
		return nil, fmt.Errorf("no context window result found")
	}

	return result.ContextWindow, nil
}

// LoadMaxOutputResult loads a max output probe result
func (s *JSONResultStorage) LoadMaxOutputResult(provider, model string) (interface{}, error) {
	fileName := fmt.Sprintf("%s-%s.json", sanitizeProviderName(provider), sanitizeModelName(model))
	filePath := filepath.Join(s.baseDir, fileName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read result file: %w", err)
	}

	var result SavedResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	if result.MaxOutput == nil {
		return nil, fmt.Errorf("no max output result found")
	}

	return result.MaxOutput, nil
}

// saveToFile helper function to save data as JSON
func (s *JSONResultStorage) saveToFile(filePath string, data SavedResult) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	// Write to temporary file first
	tempFile := filePath + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Rename to final file (atomic operation)
	if err := os.Rename(tempFile, filePath); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// sanitizeProviderName sanitizes provider name for file system
func sanitizeProviderName(provider string) string {
	// Convert to lowercase and replace spaces
	sanitized := provider
	replacements := map[string]string{
		" ": "_",
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

// sanitizeModelName sanitizes model name for file system (same as in logging package)
func sanitizeModelName(model string) string {
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

// GetDefaultResultDir returns the default directory for storing results
func GetDefaultResultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/tmp"
	}
	return filepath.Join(home, ".config", "llm-info", "estimates")
}