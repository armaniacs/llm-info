package probe

import (
	"fmt"
	"strings"
)

// TestDataGenerator は探索用のテストデータを生成する
type TestDataGenerator struct {
	sampleTexts []string
}

// NewTestDataGenerator は新しい TestDataGenerator を作成する
func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{
		sampleTexts: []string{
			"吾輩は猫である。名前はまだ無い。",
			"どこで生れたかとんと見当がつかぬ。",
			"何でも薄暗いじめじめした所でニャーニャー泣いていた事だけは記憶している。",
			"吾輩はここで始めて人間というものを見た。",
			"名前はまだ無いが、分厚い hashMap にデーターつの入力のバッファーに記憶されていた。",
			"次に茶のところへ行った。茶の湯は沸々と泡を立てている。",
			"少し待っていると、ご主人さんが出てきた。",
			"「お可愛いものですね。 definitiveiyar rag a\"",
			"その家飼い猫として迎えられ、",
			"「名前はまだないが、君もね」と言われた。",
			"猫はプードルのような声でニャーと鳴いた。",
			"「吾輩は猫である」。",
			"かく言っているような顔をした。",
			"その猫から見取れるのは、猫の母親の眼の色と、",
			"諸君の眼の色で、吾輩がただ人間の言葉を話しているという事であった。",
		},
	}
}

// GenerateData は指定されたトークン数に合うテストデータを生成する
func (g *TestDataGenerator) GenerateData(targetTokens int) (string, []string) {
	// まず基本的なコンポーネント
	preamble := "以下の内容を記憶してください。"
	needle := "【重要情報】ラッキーカラーは青色です"
	question := "ラッキーカラーは何色でしたか？"

	// 本文を構築
	var bodyBuilder strings.Builder

	// サンプルテキストを繰り返して目標トに近づける
	for len(bodyBuilder.String()) < targetTokens*3/4 { // 日本語なので1文字≈1トークンと仮定
		for _, text := range g.sampleTexts {
			if len(bodyBuilder.String())+len(text)+len(needle)+len(question) > targetTokens*3/4 {
				break
			}
			bodyBuilder.WriteString(text)
			bodyBuilder.WriteString(" ")
		}
	}

	// 最終的なテキストを組み立て
	result := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
		preamble,
		bodyBuilder.String(),
		needle,
		question,
	)

	return result, []string{preamble, bodyBuilder.String(), needle, question}
}

// GenerateWithNeedlePosition はneedleの位置を指定してデータを生成する
func (g *TestDataGenerator) GenerateWithNeedlePosition(targetTokens int, needlePosition NeedlePosition) (string, []string) {
	preamble := "以下の内容を記憶してください。"
	needle := "【重要情報】ラッキーカラーは青色です"
	question := "ラッキーカラーは何色でしたか？"

	bodyBuilder := strings.Builder{}

	// 基本的な本文
	for i := 0; i < len(g.sampleTexts); i++ {
		bodyBuilder.WriteString(g.sampleTexts[i])
		if bodyBuilder.Len() > targetTokens*3/4 {
			break
		}
		bodyBuilder.WriteString(" ")
	}

	var fullText string

	switch needlePosition {
	case End:
		fullText = fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
			preamble,
			bodyBuilder.String(),
			needle,
			question,
		)
	case Middle:
		mid := bodyBuilder.Len() / 2
		begin := bodyBuilder.String()[:mid]
		end := bodyBuilder.String()[mid:]
		fullText = fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s",
			preamble,
			begin,
			needle,
			end,
			question,
		)
	case Percent80:
		targetLen := targetTokens * 4 / 5
		if bodyBuilder.Len() < targetLen {
			// 本文が不足している場合は繰り返す
			for bodyBuilder.Len() < targetLen {
				for _, text := range g.sampleTexts {
					bodyBuilder.WriteString(text)
					if bodyBuilder.Len() >= targetLen {
						break
					}
					bodyBuilder.WriteString(" ")
				}
			}
		}

		pos := int(float64(bodyBuilder.Len()) * 0.8)
		begin := bodyBuilder.String()[:pos]
		end := bodyBuilder.String()[pos:]
		fullText = fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s",
			preamble,
			begin,
			needle,
			end,
			question,
		)
	}

	return fullText, []string{preamble, bodyBuilder.String(), needle, question}
}

// NeedlePosition はneedle（重要情報）の埋め込み位置
type NeedlePosition string

const (
	End      NeedlePosition = "end"
	Middle   NeedlePosition = "middle"
	Percent80 NeedlePosition = "80pct"
)