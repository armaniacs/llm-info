package cost

import (
	"fmt"
	"strings"

	"github.com/armaniacs/llm-info/pkg/config"
)

// Calculator provides cost calculation functionality
type Calculator struct {
	configData       *config.CostConfig
	models           map[string]config.Pricing
	costCache        map[string]float64
	warningThreshold float64
}

// NewCalculator creates a new cost calculator
func NewCalculator(costConfig *config.CostConfig, modelName string) *Calculator {
	p := make(map[string]config.Pricing)
	threshold := 0.05 // Default threshold

	if costConfig != nil {
		p = costConfig.Pricing
		if costConfig.WarningThreshold > 0 {
			threshold = costConfig.WarningThreshold
		}
	}

	return &Calculator{
		configData:       costConfig,
		models:           p,
		costCache:        make(map[string]float64),
		warningThreshold: threshold,
	}
}

// SetWarningThreshold updates the warning threshold
func (c *Calculator) SetWarningThreshold(threshold float64) {
	c.warningThreshold = threshold
	if c.configData != nil {
		c.configData.WarningThreshold = threshold
	}
}

// CalculateTrialCost calculates the cost for a single trial
func (c *Calculator) CalculateTrialCost(inputTokens, outputTokens int, modelName string) float64 {
	pricing, exists := c.models[modelName]
	if !exists {
		// Use default pricing for unknown models
		pricing = config.Pricing{
			InputPricePer1K:  0.00015, // gpt-3.5-turbo rate
			OutputPricePer1K: 0.0006,
		}
	}

	inputCost := float64(inputTokens) * pricing.InputPricePer1K / 1000
	outputCost := float64(outputTokens) * pricing.OutputPricePer1K / 1000

	return inputCost + outputCost
}

// EstimateProbeCost estimates the total cost for a probe operation
func (c *Calculator) EstimateProbeCost(modelName string, probeType string) (float64, int, int) {
	// Get pricing for the model
	pricing, exists := c.models[modelName]
	if !exists {
		// Use conservative estimates
		pricing = config.Pricing{
			InputPricePer1K:  0.00015,
			OutputPricePer1K: 0.0006,
		}
	}

	// Estimate trials based on probe type
	var estimatedTrials int
	var estimatedInputTokens int
	var estimatedOutputTokens int

	if strings.Contains(strings.ToLower(probeType), "context") {
		// Context window probe: More input tokens, less output
		estimatedTrials = 12
		estimatedInputTokens = 4096 + 8192 + 16384 + 32768 + 65536 // Sum of exponential search
		estimatedOutputTokens = estimatedTrials * 16
	} else {
		// Max output probe: Less input tokens, more output
		estimatedTrials = 12
		estimatedInputTokens = estimatedTrials * 1000
		estimatedOutputTokens = 256 + 512 + 1024 + 2048 + 4096 + 8192 // Sum of exponential search
	}

	inputCost := float64(estimatedInputTokens) * pricing.InputPricePer1K / 1000
	outputCost := float64(estimatedOutputTokens) * pricing.OutputPricePer1K / 1000

	return inputCost + outputCost, estimatedInputTokens, estimatedOutputTokens
}

// FormatCost formats a cost value for display
func (c *Calculator) FormatCost(cost float64, precision int) string {
	return fmt.Sprintf("$%.3f", cost)
}

// FormatTokenCount formats token counts with thousand separators
func (c *Calculator) FormatTokenCount(count int) string {
	s := fmt.Sprintf("%d", count)
	var result []rune

	for i, r := range reverse(s) {
		if i > 0 && i%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, r)
	}

	return string(reverse(string(result)))
}

// reverse is a helper to reverse a string
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
