package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// VerboseFormatter provides real-time verbose logging for probe operations
type VerboseFormatter struct {
	mu             sync.Mutex
	lastLineLength int
	startTime      time.Time
	trialCount     int
}

// NewVerboseFormatter creates a new VerboseFormatter
func NewVerboseFormatter() *VerboseFormatter {
	return &VerboseFormatter{
		startTime: time.Now(),
	}
}

// LogInfo logs an informational message
func (vf *VerboseFormatter) LogInfo(message string) {
	vf.mu.Lock()
	defer vf.mu.Unlock()
	vf.clearLastLine()
	msg := fmt.Sprintf("[INFO] %s", message)
	fmt.Println(msg)
	vf.lastLineLength = 0
}

// LogProgress logs a progress update (updates same line)
func (vf *VerboseFormatter) LogProgress(current, total int, tokens int) {
	vf.mu.Lock()
	defer vf.mu.Unlock()
	vf.clearLastLine()

	// Calculate ETA
	elapsed := time.Since(vf.startTime)
	var eta string
	if current > 0 {
		remaining := time.Duration(float64(elapsed) * float64(total-current) / float64(current))
		eta = fmt.Sprintf(" ETA: %v", remaining.Round(time.Second))
	} else {
		eta = ""
	}

	// Build progress message
	msg := fmt.Sprintf("[PROGRESS] Trial %d/%d: Testing with %d tokens...%s",
		current, total, tokens, eta)

	// Print with carriage return to update same line
	fmt.Print("\r" + msg)
	vf.lastLineLength = len(msg)
}

// LogSuccess logs a successful trial result
func (vf *VerboseFormatter) LogSuccess(trial int, tokens int, duration time.Duration) {
	vf.mu.Lock()
	defer vf.mu.Unlock()
	vf.clearLastLine()
	msg := fmt.Sprintf("[SUCCESS] Trial %d: Success (prompt_tokens=%d, duration=%v)",
		trial, tokens, duration.Round(time.Millisecond))
	fmt.Println(msg)
	vf.lastLineLength = 0
	vf.trialCount++
}

// LogFailure logs a failed trial
func (vf *VerboseFormatter) LogFailure(trial int, tokens int, reason string) {
	vf.mu.Lock()
	defer vf.mu.Unlock()
	vf.clearLastLine()
	msg := fmt.Sprintf("[PROGRESS] Trial %d: Failed (%d tokens) - %s",
		trial, tokens, reason)
	fmt.Println(msg)
	vf.lastLineLength = 0
	vf.trialCount++
}

// LogAPIRequest logs essential API request details
func (vf *VerboseFormatter) LogAPIRequest(method, url string, tokens int, temperature float64) {
	vf.mu.Lock()
	defer vf.mu.Unlock()
	vf.clearLastLine()
	msg := fmt.Sprintf("[INFO] API Request: %s %s (tokens=%d, temperature=%.1f)",
		method, url, tokens, temperature)
	fmt.Println(msg)
	vf.lastLineLength = 0
}

// LogAPIResponse logs essential API response details
func (vf *VerboseFormatter) LogAPIResponse(status int, promptTokens, completionTokens int, duration time.Duration) {
	vf.mu.Lock()
	defer vf.mu.Unlock()
	vf.clearLastLine()
	msg := fmt.Sprintf("[INFO] API Response: status=%d (prompt=%d, completion=%d, duration=%v)",
		status, promptTokens, completionTokens, duration.Round(time.Millisecond))
	fmt.Println(msg)
	vf.lastLineLength = 0
}

// LogSearchStrategy logs search algorithm decisions
func (vf *VerboseFormatter) LogSearchStrategy(strategy, reason string, details map[string]any) {
	vf.mu.Lock()
	defer vf.mu.Unlock()
	vf.clearLastLine()

	var detailStrs []string
	for k, v := range details {
		detailStrs = append(detailStrs, fmt.Sprintf("%s=%v", k, v))
	}

	msg := fmt.Sprintf("[INFO] %s: %s (%s)",
		strategy, reason, strings.Join(detailStrs, ", "))
	fmt.Println(msg)
	vf.lastLineLength = 0
}

// LogError logs an error with stack trace
func (vf *VerboseFormatter) LogError(err error, context string) {
	vf.mu.Lock()
	defer vf.mu.Unlock()
	vf.clearLastLine()
	msg := fmt.Sprintf("[ERROR] %s: %v", context, err)
	fmt.Println(msg)
	vf.lastLineLength = 0
}

// LogCompletion logs the final completion message
func (vf *VerboseFormatter) LogCompletion(strategy string, finalEstimate int, tolerance int) {
	vf.mu.Lock()
	defer vf.mu.Unlock()
	vf.clearLastLine()
	msg := fmt.Sprintf("[INFO] Converged! Final estimate: %d tokens (±%d) using %s",
		finalEstimate, tolerance, strategy)
	fmt.Println(msg)
	vf.lastLineLength = 0
}

// clearLastLine clears the last line if it was a progress update
func (vf *VerboseFormatter) clearLastLine() {
	if vf.lastLineLength > 0 {
		fmt.Print("\r" + strings.Repeat(" ", vf.lastLineLength) + "\r")
		vf.lastLineLength = 0
	}
}

// Ensure the last line is properly cleared when done
func (vf *VerboseFormatter) Finish() {
	vf.mu.Lock()
	defer vf.mu.Unlock()
	vf.clearLastLine()
}