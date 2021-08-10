package commands

import (
	"context"
	"testing"

	. "github.com/otiai10/mint"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func TestAmeshCommand_Match(t *testing.T) {
	str := &mockStorage{}
	cmd := AmeshCommand{Storage: str}
	m := cmd.Match(slackevents.AppMentionEvent{Text: "@amesh"})
	Expect(t, m).ToBe(true)
}

func TestAmeshCommand_Execute(t *testing.T) {
	ctx := context.Background()
	str := &mockStorage{}
	cmd := AmeshCommand{Storage: str}
	ev := slackevents.AppMentionEvent{}
	scl := &mockSlackClient{}

	ev.Text = "@amesh"

	err := cmd.Execute(ctx, scl, ev)
	Expect(t, err).ToBe(nil)

	When(t, "animated option given", func(t *testing.T) {
		scl := &mockSlackClient{}
		ev.Text = "@amesh -a"
		err := cmd.Execute(ctx, scl, ev)
		Expect(t, err).ToBe(nil)
		Expect(t, len(scl.messages)).ToBe(2)
		msg := scl.messages[0]
		Expect(t, msg.Blocks[0].BlockType()).ToBe(slack.MBTContext)
	})

	When(t, "help requested", func(t *testing.T) {
		scl := &mockSlackClient{}
		ev.Text = "@amesh -h"
		err := cmd.Execute(ctx, scl, ev)
		Expect(t, err).ToBe(nil)
	})
}

func TestAmeshCommand_Help(t *testing.T) {
	str := &mockStorage{}
	cmd := AmeshCommand{Storage: str}
	msg := cmd.Help()
	Expect(t, msg).Not().ToBe("")
}
