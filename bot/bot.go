package bot

import (
	"context"
	"fmt"

	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/largo"
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
		// 特定の発言にMatchしたら発動するコマンド
		Commands []Command
		// コマンドが無くても発動するコマンド
		Default Command
		// 不明なコマンドを受け取った場合の挙動
		NotFound Command
	}
)

func (b *Bot) Handle(ctx context.Context, team slack.OAuthV2Response, event slackevents.AppMentionEvent) {
	client := service.NewSlackClient(team.AccessToken)
	err := b.handle(ctx, client, event)
	if err != nil {
		fmt.Printf("[ERROR] bot.Handle: %v\n%+v", err, event)
	}
}

func (b *Bot) handle(ctx context.Context, client *service.SlackClient, event slackevents.AppMentionEvent) (err error) {
	if tokens := largo.Tokenize(event.Text)[1:]; len(tokens) != 0 && tokens[0] == "help" {
		return b.help(ctx, client, event)
	}
	for _, cmd := range b.Commands {
		if cmd.Match(event) {
			err = cmd.Execute(ctx, client, event)
			return b.errwrap(err, cmd)
		}
	}
	if b.Default.Match(event) {
		err = b.Default.Execute(ctx, client, event)
		return b.errwrap(err, b.Default)
	}
	if b.NotFound != nil {
		err = b.NotFound.Execute(ctx, client, event)
		return
	}
	return
}

func (b *Bot) errwrap(err error, cmd interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%T: %v", cmd, err.Error())
}
