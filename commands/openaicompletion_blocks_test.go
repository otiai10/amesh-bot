package commands

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/slack-go/slack"
)

func TestParseBlockKitResponse(t *testing.T) {
	cases := []struct {
		name         string
		input        string
		wantBlocks   int
		wantFallback string
		wantErr      bool
	}{
		{
			name:         "sectionブロック単体",
			input:        `{"fallback_text":"こんにちは","blocks":[{"type":"section","text":{"type":"mrkdwn","text":"*こんにちは*、世界！"}}]}`,
			wantBlocks:   1,
			wantFallback: "こんにちは",
		},
		{
			name: "複数sectionブロック",
			input: mustJSON(slackBlockKitResponse{
				FallbackText: "要約テスト",
				Blocks: []slackBlockJSON{
					{Type: "section", Text: &slackTextObjectJSON{Type: "mrkdwn", Text: "段落1"}},
					{Type: "section", Text: &slackTextObjectJSON{Type: "mrkdwn", Text: "段落2"}},
					{Type: "section", Text: &slackTextObjectJSON{Type: "mrkdwn", Text: "段落3"}},
				},
			}),
			wantBlocks:   3,
			wantFallback: "要約テスト",
		},
		{
			name:    "空blocks配列",
			input:   `{"fallback_text":"x","blocks":[]}`,
			wantErr: true,
		},
		{
			name:    "不正JSON",
			input:   `{not valid json`,
			wantErr: true,
		},
		{
			name: "未知のブロックタイプは無視される",
			input: `{"fallback_text":"test","blocks":[
				{"type":"unknown","text":null}
			]}`,
			wantErr: true,
		},
		{
			name: "section textが3000文字超で切り詰め",
			input: mustJSON(slackBlockKitResponse{
				FallbackText: "長文",
				Blocks: []slackBlockJSON{
					{Type: "section", Text: &slackTextObjectJSON{Type: "mrkdwn", Text: strings.Repeat("あ", 3500)}},
				},
			}),
			wantBlocks:   1,
			wantFallback: "長文",
		},
		{
			name: "50ブロック超で切り詰め",
			input: mustJSON(slackBlockKitResponse{
				FallbackText: "many",
				Blocks: func() []slackBlockJSON {
					blocks := make([]slackBlockJSON, 55)
					for i := range blocks {
						blocks[i] = slackBlockJSON{Type: "section", Text: &slackTextObjectJSON{Type: "mrkdwn", Text: "x"}}
					}
					return blocks
				}(),
			}),
			wantBlocks:   50,
			wantFallback: "many",
		},
		{
			name: "textがnilのsectionは無視される",
			input: mustJSON(slackBlockKitResponse{
				FallbackText: "nil-text",
				Blocks: []slackBlockJSON{
					{Type: "section", Text: nil},
					{Type: "section", Text: &slackTextObjectJSON{Type: "mrkdwn", Text: "有効"}},
				},
			}),
			wantBlocks:   1,
			wantFallback: "nil-text",
		},
		{
			name: "text.textが空文字のsectionは無視される",
			input: mustJSON(slackBlockKitResponse{
				FallbackText: "empty-text",
				Blocks: []slackBlockJSON{
					{Type: "section", Text: &slackTextObjectJSON{Type: "mrkdwn", Text: ""}},
					{Type: "section", Text: &slackTextObjectJSON{Type: "mrkdwn", Text: "有効"}},
				},
			}),
			wantBlocks:   1,
			wantFallback: "empty-text",
		},
		{
			name: "dividerやcontextは無視される",
			input: `{"fallback_text":"ignore","blocks":[
				{"type":"divider","text":null},
				{"type":"section","text":{"type":"mrkdwn","text":"有効"}},
				{"type":"context","text":null}
			]}`,
			wantBlocks:   1,
			wantFallback: "ignore",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fallback, blocks, err := parseBlockKitResponse(c.input)
			if c.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(blocks) != c.wantBlocks {
				t.Errorf("blocks count: got %d, want %d", len(blocks), c.wantBlocks)
			}
			if fallback != c.wantFallback {
				t.Errorf("fallback: got %q, want %q", fallback, c.wantFallback)
			}
		})
	}
}

func TestParseBlockKitResponse_SectionTruncation(t *testing.T) {
	input := mustJSON(slackBlockKitResponse{
		FallbackText: "truncated",
		Blocks: []slackBlockJSON{
			{Type: "section", Text: &slackTextObjectJSON{Type: "mrkdwn", Text: strings.Repeat("あ", 3500)}},
		},
	})
	_, blocks, err := parseBlockKitResponse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	section, ok := blocks[0].(*slack.SectionBlock)
	if !ok {
		t.Fatalf("expected SectionBlock, got %T", blocks[0])
	}
	if len([]rune(section.Text.Text)) != 3000 {
		t.Errorf("section text rune count: got %d, want 3000", len([]rune(section.Text.Text)))
	}
}

func TestTruncateForNotification(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "短い文字列はそのまま",
			input: "短いテキスト",
			want:  "短いテキスト",
		},
		{
			name:  "150文字ちょうどはそのまま",
			input: strings.Repeat("a", 150),
			want:  strings.Repeat("a", 150),
		},
		{
			name:  "151文字以上は147文字+...",
			input: strings.Repeat("a", 200),
			want:  strings.Repeat("a", 147) + "...",
		},
		{
			name:  "日本語の長文",
			input: strings.Repeat("あ", 200),
			want:  strings.Repeat("あ", 147) + "...",
		},
		{
			name:  "空文字列",
			input: "",
			want:  "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := truncateForNotification(c.input)
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
