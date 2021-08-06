package commands

import (
	"context"
	"fmt"
	"testing"

	"github.com/slack-go/slack/slackevents"

	. "github.com/otiai10/mint"
)

func TestLGTMCommand_Match(t *testing.T) {
	cmd := LGTMCommand{}
	m := cmd.Match(slackevents.AppMentionEvent{Text: "@amesh lgtm"})
	Expect(t, m).ToBe(true)
	m = cmd.Match(slackevents.AppMentionEvent{Text: "@amesh"})
	Expect(t, m).ToBe(false)
}

func TestLGTMCommand_Help(t *testing.T) {
	cmd := LGTMCommand{}
	help := cmd.Help()
	Expect(t, help).ToBe("lgtmコマンド\n```@amesh lgtm```")
}

func TestLGTMCommand_Execute(t *testing.T) {
	cmd := LGTMCommand{Service: &mockLGTM{
		imgurl: "https://lgtm.lol/p/100",
	}}
	ctx := context.Background()
	sc := &mockSlackClient{}
	err := cmd.Execute(ctx, sc, slackevents.AppMentionEvent{Text: "@amesh lgtm"})
	Expect(t, err).ToBe(nil)

	When(t, "service returns error", func(t *testing.T) {
		cmd.Service = &mockLGTM{err: fmt.Errorf("foo baa")}
		sc := &mockSlackClient{}
		err := cmd.Execute(ctx, sc, slackevents.AppMentionEvent{Text: "@amesh lgtm"})
		Expect(t, err).ToBe(nil)
	})
}
