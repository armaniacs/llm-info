package api

import "fmt"

// EndpointType はAPIエンドポイントの種類を表す
type EndpointType int

const (
	// LiteLLM はLiteLLM互換のエンドポイントを表す
	LiteLLM EndpointType = iota
	// OpenAIStandard はOpenAI標準互換のエンドポイントを表す
	OpenAIStandard
)

// Endpoint はAPIエンドポイント情報を表す
type Endpoint struct {
	Type EndpointType
	Path string
}

// GetEndpoints は利用可能なエンドポイント一覧を返す
func GetEndpoints(baseURL string) []Endpoint {
	return []Endpoint{
		{Type: LiteLLM, Path: fmt.Sprintf("%s/model/info", baseURL)},
		{Type: OpenAIStandard, Path: fmt.Sprintf("%s/v1/models", baseURL)},
	}
}

// String はEndpointTypeの文字列表現を返す
func (et EndpointType) String() string {
	switch et {
	case LiteLLM:
		return "LiteLLM"
	case OpenAIStandard:
		return "OpenAI Standard"
	default:
		return "Unknown"
	}
}
