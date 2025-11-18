package bot

import (
	"context"

	"cloud.google.com/go/logging"
	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/largo"
	"github.com/slack-go/slack/slackevents"
)

type (
	Command interface {
		Match(event slackevents.AppMentionEvent) bool
		Execute(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) *CommandError
		Help() string
	}
	Logger interface {
		Log(logging.Entry)
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

		Logger Logger
	}
)

func (b *Bot) Handle(ctx context.Context, team service.OAuthResponse, event slackevents.AppMentionEvent) {
	client := service.NewSlackClient(team.AccessToken)

	// if os.Getenv("DEV_SLACK_APP_ID") != "" {
	// 	b.Logger.Log(logging.Entry{Severity: logging.Debug, Payload: event})
	// }

	if cmderr := b.handle(ctx, client, event); cmderr != nil {
		b.sendErrorMessage(ctx, client, event, cmderr)
		b.Logger.Log(logging.Entry{Severity: logging.Error, Payload: cmderr, Labels: cmderr.labels()})
	}
}

func (b *Bot) handle(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) *CommandError {
	if tokens := largo.Tokenize(event.Text)[1:]; len(tokens) != 0 && tokens[0] == "help" {
		return wrapWithContext(b.Help(ctx, client, event), "builtin:help", event)
	}
	for _, cmd := range b.Commands {
		if cmd.Match(event) {
			return wrapWithContext(cmd.Execute(ctx, client, event), cmd, event)
		}
	}
	if b.Default != nil && b.Default.Match(event) {
		return wrapWithContext(b.Default.Execute(ctx, client, event), b.Default, event)
	}
	if b.NotFound != nil && b.NotFound.Match(event) {
		return wrapWithContext(b.NotFound.Execute(ctx, client, event), "builtin:notfound", event)
	}
	return nil
}

func (b *Bot) sendErrorMessage(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent, cmderr *CommandError) {
	if cmderr == nil || cmderr.Message == "" {
		return
	}
	msg := service.SlackMsg{Channel: event.Channel, Text: cmderr.Message}
	if cmderr.ThreadTimestamp != "" {
		msg.ThreadTimestamp = cmderr.ThreadTimestamp
	} else if event.ThreadTimeStamp != "" {
		msg.ThreadTimestamp = event.ThreadTimeStamp
	}
	if _, err := client.PostMessage(ctx, msg); err != nil {
		b.Logger.Log(logging.Entry{Severity: logging.Critical, Payload: map[string]string{
			"error":   err.Error(),
			"context": "failed to reply error message",
		}})
	}
}
