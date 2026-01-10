package api

// ModelInfoResponse はAPIレスポンスの構造体です
type ModelInfoResponse struct {
	Models []ModelInfo `json:"models"`
}

// ModelInfo は個別のモデル情報です
type ModelInfo struct {
	ID        string  `json:"id"`
	MaxTokens int     `json:"max_tokens"`
	Mode      string  `json:"mode"`
	InputCost float64 `json:"input_cost"`
}
