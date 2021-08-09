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
	scl := &mockSlackClient{}
	err := cmd.Execute(ctx, scl, slackevents.AppMentionEvent{Text: "@amesh -a"})
	Expect(t, err).ToBe(nil)
	Expect(t, len(scl.messages)).ToBe(2)
	msg := scl.messages[0]
	Expect(t, msg.Blocks[0].BlockType()).ToBe(slack.MBTContext)
	Expect(t, msg.Blocks[1].BlockType()).ToBe(slack.MBTImage)
	blck := msg.Blocks[0].(*slack.ImageBlock)
	Expect(t, blck.ImageURL).Match(".gif")

	err = cmd.Execute(ctx, scl, slackevents.AppMentionEvent{Text: "@amesh"})
	Expect(t, err).ToBe(nil)
	Expect(t, scl.messages[1].Blocks[0].(*slack.ImageBlock).ImageURL).Match(".png")
}

func TestAmeshCommand_Help(t *testing.T) {
	str := &mockStorage{}
	cmd := AmeshCommand{Storage: str}
	msg := cmd.Help()
	Expect(t, msg).Not().ToBe("")
}
