package bot

import (
	"context"

	"github.com/otiai10/amesh-bot/service"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type (
	Command interface {
		Match(event slackevents.AppMentionEvent) bool
		Execute(ctx context.Context, client *service.SlackClient, event slackevents.AppMentionEvent) error
		Help() string
	}
)

type (
	Bot struct {
		Commands []Command
	}
)

func (b *Bot) Handle(ctx context.Context, team slack.OAuthV2Response, event slackevents.AppMentionEvent) {
	client := service.NewSlackClient(team.AccessToken)
	// TODO: Handle top-level help
	for _, cmd := range b.Commands {
		if cmd.Match(event) {
			// TODO: error handling
			cmd.Execute(ctx, client, event)
		}
	}
}
