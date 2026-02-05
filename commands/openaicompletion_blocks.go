package commands

import (
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
	"github.com/slack-go/slack"
)

const openaiBlockKitInstructions = `あなたはSlack Bot「amesh」です。ユーザーの質問に対して、Slack Block Kit形式のJSON構造で回答を返してください。

## 出力ルール
- fallback_text: 通知プレビュー用の短いプレーンテキスト要約（150文字以内）
- blocks: Slack Block Kitブロックの配列

## 使用可能なブロックタイプ
1. section: メインコンテンツ用。text.textは3000文字以内。長い回答はsectionブロックを分けて3000文字制限を守ること。
2. divider: 区切り線。トピックが変わる時に使用。
3. context: 補足情報やメタデータ用。elementsは最大10個。

## テキストフォーマット（Slack mrkdwn記法）
- 太字: *太字*（標準Markdownの**太字**ではない）
- イタリック: _イタリック_
- 取り消し線: ~取り消し線~
- インラインコード: ` + "`コード`" + `
- コードブロック: ` + "```コードブロック```" + `
- リンク: <https://example.com|表示テキスト>（標準Markdownの[text](url)ではない）
- リスト: 改行と「• 」や「- 」で表現
- 引用: > で始める

## 注意事項
- 回答は日本語で行う（ユーザーが英語で質問した場合は英語で回答）
- 簡潔で読みやすい回答を心がける`

// slackBlocksSchema は OpenAI Structured Outputs で使用するJSON Schemaの定義。
// strict: true に対応するため、全propertyキーをrequiredに含め、
// オプショナルなフィールドはnullable（"type": ["...", "null"]）とする。
var slackBlocksSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"fallback_text": map[string]any{
			"type":        "string",
			"description": "通知プレビュー用の短いプレーンテキスト要約（150文字以内）",
		},
		"blocks": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"type": map[string]any{
						"type": "string",
						"enum": []any{"section", "divider", "context"},
					},
					"text": map[string]any{
						"type": []any{"object", "null"},
						"properties": map[string]any{
							"type": map[string]any{"type": "string", "enum": []any{"mrkdwn"}},
							"text": map[string]any{"type": "string"},
						},
						"required":             []any{"type", "text"},
						"additionalProperties": false,
					},
					"elements": map[string]any{
						"type": []any{"array", "null"},
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"type": map[string]any{"type": "string", "enum": []any{"mrkdwn"}},
								"text": map[string]any{"type": "string"},
							},
							"required":             []any{"type", "text"},
							"additionalProperties": false,
						},
					},
				},
				"required":             []any{"type", "text", "elements"},
				"additionalProperties": false,
			},
		},
	},
	"required":             []any{"fallback_text", "blocks"},
	"additionalProperties": false,
}

// slackBlockKitResponse はLLMから返されるJSON構造の中間表現。
type slackBlockKitResponse struct {
	FallbackText string           `json:"fallback_text"`
	Blocks       []slackBlockJSON `json:"blocks"`
}

type slackBlockJSON struct {
	Type     string                `json:"type"`
	Text     *slackTextObjectJSON  `json:"text"`
	Elements []slackTextObjectJSON `json:"elements"`
}

type slackTextObjectJSON struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// blockKitTextConfig は OpenAI Responses API に渡す Text 設定を返す。
func blockKitTextConfig() responses.ResponseTextConfigParam {
	return responses.ResponseTextConfigParam{
		Format: responses.ResponseFormatTextConfigUnionParam{
			OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
				Name:   "slack_block_kit",
				Schema: slackBlocksSchema,
				Strict: openai.Bool(true),
			},
		},
	}
}

// parseBlockKitResponse はLLMのJSON出力をパースし、Slack Block Kitのブロック配列に変換する。
// パース失敗時はerrorを返し、呼び出し元でプレーンテキストへfallbackする。
func parseBlockKitResponse(output string) (fallbackText string, blocks []slack.Block, err error) {
	var resp slackBlockKitResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		return "", nil, fmt.Errorf("json unmarshal: %w", err)
	}
	if len(resp.Blocks) == 0 {
		return "", nil, fmt.Errorf("empty blocks array")
	}
	// Slack制限: 最大50ブロック
	if len(resp.Blocks) > 50 {
		resp.Blocks = resp.Blocks[:50]
	}
	for _, b := range resp.Blocks {
		switch b.Type {
		case "section":
			if b.Text != nil && b.Text.Text != "" {
				text := b.Text.Text
				// Slack制限: SectionBlock textは最大3000文字
				if len([]rune(text)) > 3000 {
					text = string([]rune(text)[:3000])
				}
				textObj := slack.NewTextBlockObject(slack.MarkdownType, text, false, false)
				blocks = append(blocks, slack.NewSectionBlock(textObj, nil, nil))
			}
		case "divider":
			blocks = append(blocks, slack.NewDividerBlock())
		case "context":
			elements := b.Elements
			// Slack制限: ContextBlockは最大10要素
			if len(elements) > 10 {
				elements = elements[:10]
			}
			var mixed []slack.MixedElement
			for _, e := range elements {
				mixed = append(mixed, slack.NewTextBlockObject(slack.MarkdownType, e.Text, false, false))
			}
			if len(mixed) > 0 {
				blocks = append(blocks, slack.NewContextBlock("", mixed...))
			}
		}
	}
	if len(blocks) == 0 {
		return "", nil, fmt.Errorf("no valid blocks produced")
	}
	return resp.FallbackText, blocks, nil
}

// truncateForNotification は通知プレビュー用にテキストを150文字（rune単位）に切り詰める。
func truncateForNotification(s string) string {
	runes := []rune(s)
	if len(runes) <= 150 {
		return s
	}
	return string(runes[:147]) + "..."
}
