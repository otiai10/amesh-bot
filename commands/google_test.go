package commands

import (
	"context"
	"testing"

	. "github.com/otiai10/mint"
	"github.com/slack-go/slack/slackevents"
)

func TestGoogleCommand_Match(t *testing.T) {
	cmd := GoogleCommand{}
	m := cmd.Match(slackevents.AppMentionEvent{Text: "@amesh ggl pikachu"})
	Expect(t, m).ToBe(true)
	m = cmd.Match(slackevents.AppMentionEvent{Text: "@amesh"})
	Expect(t, m).ToBe(false)
	m = cmd.Match(slackevents.AppMentionEvent{Text: "@amesh foobaa"})
	Expect(t, m).ToBe(false)
}

func TestGoogleCommand_Help(t *testing.T) {
	cmd := GoogleCommand{}
	help := cmd.Help()
	Expect(t, help).ToBe("グーグル検索コマンド\n```@amesh ggl|google {query}```")
}

func TestGoogleCommand_Execute(t *testing.T) {
	sc := &mockSlackClient{}
	search := &mockGoogleClient{}
	search.ResponseBody = `{
		"items": [
			{"title":"hoge", "link":"qwerty"}
		]
	}`
	ctx := context.Background()
	cmd := GoogleCommand{Search: search}
	event := slackevents.AppMentionEvent{Text: "@amesh ggl pikachu"}
	err := cmd.Execute(ctx, sc, event)
	Expect(t, err).ToBe(nil)

	msg := sc.messages[0]
	Expect(t, msg.Text).ToBe("> pikachu\nqwerty\n")

	When(t, "no item found", func(t *testing.T) {
		sc := &mockSlackClient{}
		search.ResponseBody = `{
			"items": []
		}`
		cmd.Search = search
		err := cmd.Execute(ctx, sc, event)
		Expect(t, err).ToBe(nil)
		msg := sc.messages[0]
		Expect(t, msg.Text).Match("Not found for query:")
	})
}
