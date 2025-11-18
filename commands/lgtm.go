package commands

import (
	"bytes"
	"context"

	"github.com/otiai10/amesh-bot/bot"
	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/largo"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type (
	LGTM interface {
		Random() (string, string, error)
	}
)

type LGTMCommand struct {
	Service LGTM
}

func (cmd LGTMCommand) Match(event slackevents.AppMentionEvent) bool {
	tokens := largo.Tokenize(event.Text)[1:]
	return len(tokens) > 0 && tokens[0] == "lgtm"
}

func (cmd LGTMCommand) Help() string {
	return "lgtmコマンド\n```@amesh lgtm [-markdown|-md]```"
}

func (cmd LGTMCommand) Execute(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) *bot.CommandError {
	imgurl, mrkdwn, err := cmd.Service.Random()
	if err != nil {
		return commandErrorWithMessage(err, ":warning: LGTM画像の取得に失敗しました。")
	}
	msg := inreply(event)

	help := bytes.NewBuffer(nil)
	fset := largo.NewFlagSet("", largo.ContinueOnError)
	fset.Output = help
	var includeMarkdown bool
	fset.BoolVar(&includeMarkdown, "markdown", false, "Markdownを表示").Alias("md")

	fset.Parse(largo.Tokenize(event.Text)[2:])

	if fset.HelpRequested() {
		msg.Text = cmd.Help()
		if _, err := client.PostMessage(ctx, msg); err != nil {
			return commandError(err)
		}
		return nil
	}

	msg.Blocks = append(msg.Blocks, slack.NewImageBlock(imgurl, "LGTM", "", nil))
	if includeMarkdown {
		msg.Blocks = append(msg.Blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject(slack.MarkdownType, "```"+mrkdwn+"```", false, false), nil, nil),
		)
	}

	if _, err := client.PostMessage(ctx, msg); err != nil {
		return commandError(err)
	}
	return nil
}
