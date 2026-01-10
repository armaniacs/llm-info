package api

import (
	"testing"
)

func TestGetEndpoints(t *testing.T) {
	baseURL := "https://api.example.com"
	endpoints := GetEndpoints(baseURL)

	if len(endpoints) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(endpoints))
	}

	// LiteLLMエンドポイントの確認
	if endpoints[0].Type != LiteLLM {
		t.Errorf("Expected first endpoint to be LiteLLM, got %v", endpoints[0].Type)
	}
	expectedLiteLLMPath := "https://api.example.com/model/info"
	if endpoints[0].Path != expectedLiteLLMPath {
		t.Errorf("Expected LiteLLM path to be %s, got %s", expectedLiteLLMPath, endpoints[0].Path)
	}

	// OpenAI標準エンドポイントの確認
	if endpoints[1].Type != OpenAIStandard {
		t.Errorf("Expected second endpoint to be OpenAIStandard, got %v", endpoints[1].Type)
	}
	expectedStandardPath := "https://api.example.com/v1/models"
	if endpoints[1].Path != expectedStandardPath {
		t.Errorf("Expected OpenAI Standard path to be %s, got %s", expectedStandardPath, endpoints[1].Path)
	}
}

func TestEndpointTypeString(t *testing.T) {
	tests := []struct {
		endpointType EndpointType
		expected     string
	}{
		{LiteLLM, "LiteLLM"},
		{OpenAIStandard, "OpenAI Standard"},
		{EndpointType(999), "Unknown"},
	}

	for _, test := range tests {
		result := test.endpointType.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}
