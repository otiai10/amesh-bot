package bot

import (
	"context"
	"testing"

	"cloud.google.com/go/logging"
	"github.com/otiai10/amesh-bot/service"
	m "github.com/otiai10/mint"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type mockLog struct{}

func (ml *mockLog) Logger(name string, opt ...logging.LoggerOption) *logging.Logger {
	return new(logging.Logger)
}

type mockSlack struct{}

func (ms *mockSlack) PostMessage(ctx context.Context, msg interface{}) (*slack.SlackResponse, error) {
	return nil, nil
}

func TestBot_Handle(t *testing.T) {
	bot := Bot{Log: &mockLog{}}
	ctx := context.Background()
	bot.Handle(ctx, service.OAuthResponse{}, slackevents.AppMentionEvent{Text: "@amesh"})
	m.Expect(t, true).ToBe(true)
}

func TestBot_Help(t *testing.T) {
	bot := Bot{Log: &mockLog{}}
	ctx := context.Background()
	bot.Help(ctx, &mockSlack{}, slackevents.AppMentionEvent{})
}
