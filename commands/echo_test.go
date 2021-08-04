package commands

import (
	"context"
	"testing"

	. "github.com/otiai10/mint"
	"github.com/slack-go/slack/slackevents"
)

func TestEchoCommand_Match(t *testing.T) {
	cmd := EchoCommand{}
	m := cmd.Match(slackevents.AppMentionEvent{Text: "@amesh echo foo baa baz"})
	Expect(t, m).ToBe(true)
}

func TestEchoCommand_Execute(t *testing.T) {
	ctx := context.Background()
	cmd := EchoCommand{}
	scl := &mockSlackClient{}
	err := cmd.Execute(ctx, scl, slackevents.AppMentionEvent{Text: "@amesh echo foobaa"})
	Expect(t, err).ToBe(nil)
	Expect(t, scl.messages[0].Text).ToBe("foobaa")
}

func TestEchoCommand_Help(t *testing.T) {
	cmd := EchoCommand{}
	msg := cmd.Help()
	Expect(t, msg).ToBe("オウム返しコマンド\n```@amesh echo {text1} {text2...}```")
}
