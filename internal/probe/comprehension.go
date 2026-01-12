package probe

import (
	"strings"
)

// ComprehensionResult はcomprehensionテストの結果
type ComprehensionResult struct {
	Correct  bool
	Answer   string
	Expected string
}

// CheckComprehension はモデルの応答に期待される回答が含まれるかチェックする
func CheckComprehension(response, expected string) ComprehensionResult {
	// 大文字小文字を区別しないで比較
	responseLower := strings.ToLower(response)
	expectedLower := strings.ToLower(expected)

	return ComprehensionResult{
		Correct:  strings.Contains(responseLower, expectedLower),
		Answer:   response,
		Expected: expected,
	}
}