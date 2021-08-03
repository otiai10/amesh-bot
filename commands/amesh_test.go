package commands

import (
	"context"
	"fmt"
	"testing"

	"github.com/otiai10/amesh-bot/service"
	. "github.com/otiai10/mint"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type mockStorage struct{}

func (mocks *mockStorage) Exists(ctx context.Context, bucket, name string) (bool, error) {
	return true, nil
}

func (mocks *mockStorage) URL(bucket, name string) string {
	return fmt.Sprintf("%s/%s", bucket, name)
}

func (mocks *mockStorage) Upload(ctx context.Context, bucket, name string, contents []byte) error {
	return nil
}

type mockSlackClient struct {
	messages []interface{}
}

func (sc *mockSlackClient) PostMessage(ctx context.Context, msg interface{}) (*slack.SlackResponse, error) {
	sc.messages = append(sc.messages, msg)
	return nil, nil
}

func TestAmeshCommand_Match(t *testing.T) {
	str := &mockStorage{}
	cmd := AmeshCommand{Storage: str}
	m := cmd.Match(slackevents.AppMentionEvent{Text: "@amesh"})
	Expect(t, m).ToBe(true)
}

func TestAmeshCommand_Execute(t *testing.T) {
	Expect(t, true).ToBe(true)

	ctx := context.Background()
	str := &mockStorage{}
	cmd := AmeshCommand{Storage: str}

	scl := &mockSlackClient{}
	err := cmd.Execute(ctx, scl, slackevents.AppMentionEvent{Text: "@amesh -a"})
	Expect(t, err).ToBe(nil)
	Expect(t, len(scl.messages)).ToBe(1)
	msg := scl.messages[0].(service.SlackMsg)
	Expect(t, msg.Blocks[0].BlockType()).ToBe(slack.MBTImage)
	blck := msg.Blocks[0].(*slack.ImageBlock)
	Expect(t, blck.ImageURL).Match(".gif")
}
