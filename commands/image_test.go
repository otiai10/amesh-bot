package commands

import (
	"context"
	"testing"

	. "github.com/otiai10/mint"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func TestImageCommand_Match(t *testing.T) {
	cmd := ImageCommand{}
	m := cmd.Match(slackevents.AppMentionEvent{Text: "@amesh img foo"})
	Expect(t, m).ToBe(true)
	m = cmd.Match(slackevents.AppMentionEvent{Text: "@amesh"})
	Expect(t, m).ToBe(false)
	m = cmd.Match(slackevents.AppMentionEvent{Text: "@amesh foobaa"})
	Expect(t, m).ToBe(false)
}

func TestImageCommand_Help(t *testing.T) {
	cmd := ImageCommand{}
	help := cmd.Help()
	Expect(t, help).ToBe("画像検索コマンド\n```@amesh img|image {query}```")
}

func TestImageCommand_Execute(t *testing.T) {
	sc := &mockSlackClient{}
	search := &mockGoogleClient{}
	search.ResponseBody = `{
		"items": [
			{"title":"hoge", "link":"qwerty"}
		]
	}`
	ctx := context.Background()
	cmd := ImageCommand{Search: search}
	event := slackevents.AppMentionEvent{Text: "@amesh img pikachu"}
	err := cmd.Execute(ctx, sc, event)
	Expect(t, err).ToBe(nil)

	msg := sc.messages[0]
	Expect(t, msg.Blocks[0].(*slack.ImageBlock).ImageURL).ToBe("qwerty")

	When(t, "no item found", func(t *testing.T) {
		sc := &mockSlackClient{}
		search.ResponseBody = `{
			"items": []
		}`
		cmd.Search = search
		err := cmd.Execute(ctx, sc, event)
		Expect(t, err).ToBe(nil)
	})

	When(t, "help requested", func(t *testing.T) {
		sc := &mockSlackClient{}
		event.Text = "@amesh img -h"
		err := cmd.Execute(ctx, sc, event)
		Expect(t, err).ToBe(nil)
		Expect(t, sc.messages[0].Text).Match("NAME\n  img")
	})

	When(t, "unsafe given", func(t *testing.T) {
		sc := &mockSlackClient{}
		event.Text = "@amesh img foobaa -unsafe"
		err := cmd.Execute(ctx, sc, event)
		Expect(t, err).ToBe(nil)
	})

	When(t, "verbose given", func(t *testing.T) {
		sc := &mockSlackClient{}
		search.ResponseBody = `{
			"items": [
				{"title":"hoge", "link":"qwerty"}
			]
		}`
		event.Text = "@amesh img foobaa -v"
		err := cmd.Execute(ctx, sc, event)
		Expect(t, err).ToBe(nil)
	})
}
