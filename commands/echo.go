package commands

import (
	"context"
	"strings"

	"github.com/otiai10/amesh-bot/bot"
	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/largo"
	"github.com/slack-go/slack/slackevents"
)

type EchoCommand struct {
}

// Match ...
func (cmd EchoCommand) Match(event slackevents.AppMentionEvent) bool {
	tokens := largo.Tokenize(event.Text)
	return len(tokens) > 1 && tokens[1] == "echo"
}

func (cmd EchoCommand) Execute(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) *bot.CommandError {
	msg := inreply(event)
	tokens := largo.Tokenize(event.Text)[1:]
	msg.Text = strings.Join(tokens[1:], " ")
	// スレッドの中での発言なら、スレッドに返す
	msg.ThreadTimestamp = event.ThreadTimeStamp
	if _, err := client.PostMessage(ctx, msg); err != nil {
		return commandError(err)
	}
	return nil
}

func (cmd EchoCommand) Help() string {
	return "オウム返しコマンド\n```@amesh echo {text1} {text2...}```"
}
