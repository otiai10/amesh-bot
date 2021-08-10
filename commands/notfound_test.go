package commands

import (
	"context"
	"testing"

	. "github.com/otiai10/mint"
	"github.com/slack-go/slack/slackevents"
)

func TestNotFound_Match(t *testing.T) {
	cmd := NotFound{}
	m := cmd.Match(slackevents.AppMentionEvent{})
	Expect(t, m).ToBe(true)
}

func TestNotFound_Help(t *testing.T) {
	cmd := NotFound{}
	help := cmd.Help()
	Expect(t, help).ToBe("")
}

func TestNotFound_Execute(t *testing.T) {
	ev := slackevents.AppMentionEvent{
		Text: "@amesh",
	}
	sc := &mockSlackClient{}
	ctx := context.Background()
	cmd := NotFound{}
	err := cmd.Execute(ctx, sc, ev)
	Expect(t, err).ToBe(nil)
}
