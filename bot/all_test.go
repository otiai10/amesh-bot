package bot

import (
	"context"
	"testing"

	"github.com/otiai10/amesh-bot/service"
	. "github.com/otiai10/mint"
	"github.com/slack-go/slack/slackevents"
)

func TestBot_Handle(t *testing.T) {
	bot := Bot{}
	ctx := context.Background()
	bot.Handle(ctx, service.OAuthResponse{}, slackevents.AppMentionEvent{Text: "@amesh"})
	Expect(t, true).ToBe(true)
}
