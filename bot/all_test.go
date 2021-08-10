package bot

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/logging"
	"github.com/otiai10/amesh-bot/service"
	m "github.com/otiai10/mint"
	"github.com/slack-go/slack/slackevents"
)

type mockLogger struct{}

func (ml *mockLogger) Log(e logging.Entry) {
	// pass
}

type mockSlack struct{}

func (ms *mockSlack) PostMessage(ctx context.Context, msg interface{}) (*service.PostMessageResponse, error) {
	return nil, nil
}

// func (ms *mockSlack) DeleteMessage(ctx context.Context, msg interface{}) error {
// 	return nil
// }

func (ms *mockSlack) UpdateMessage(ctx context.Context, msg interface{}) error {
	return nil
}

type dummycommand struct {
	err error
}

func (dc *dummycommand) Match(ev slackevents.AppMentionEvent) bool {
	return true
}

func (dc *dummycommand) Help() string {
	return ""
}

func (dc *dummycommand) Execute(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) error {
	return dc.err
}

func TestBot_Handle(t *testing.T) {
	oauth := service.OAuthResponse{}
	event := slackevents.AppMentionEvent{Text: "@amesh"}
	bot := Bot{Logger: &mockLogger{}}
	ctx := context.Background()
	bot.Handle(ctx, oauth, event)

	m.When(t, "command returns error", func(t *testing.T) {
		bot.Commands = append(bot.Commands, &dummycommand{err: fmt.Errorf("test_test")})
		bot.Handle(ctx, oauth, event)

		event.Text = "@amesh help"
		bot.Handle(ctx, oauth, event)
	})

	m.When(t, "default set", func(t *testing.T) {
		bot.Commands = []Command{}
		bot.Default = &dummycommand{}
		event.Text = "@amesh hoge"
		bot.Handle(ctx, oauth, event)
	})

	m.When(t, "notfound set", func(t *testing.T) {
		bot.Commands = []Command{}
		bot.Default = nil
		bot.NotFound = &dummycommand{err: fmt.Errorf("error on notfound")}
		event.Text = "@amesh hoge"
		bot.Handle(ctx, oauth, event)
	})
}

func TestBot_Help(t *testing.T) {
	bot := Bot{Logger: &mockLogger{}}
	ctx := context.Background()
	bot.Help(ctx, &mockSlack{}, slackevents.AppMentionEvent{})
}
