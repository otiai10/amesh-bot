package commands

import (
	"context"
	"fmt"

	"github.com/otiai10/amesh-bot/bot"
	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/largo"
	"github.com/slack-go/slack/slackevents"
)

type NotFound struct {
}

// Match ...
func (cmd NotFound) Match(event slackevents.AppMentionEvent) bool {
	return true
}

func (cmd NotFound) Execute(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) *bot.CommandError {
	msg := inreply(event)
	tokens := largo.Tokenize(event.Text)[1:]
	msg.Text = fmt.Sprintf("ちょっと何言ってるかわからない :face_with_rolling_eyes:\n> %v\n以下のコマンドを試してみてください.\n```@amesh help```", tokens)
	if _, err := client.PostMessage(ctx, msg); err != nil {
		return commandError(err)
	}
	return nil
}

func (cmd NotFound) Help() string {
	return ""
}
