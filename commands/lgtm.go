package commands

import (
	"context"
	"fmt"

	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/largo"
	"github.com/slack-go/slack/slackevents"
)

type (
	LGTM interface {
		Random() (string, error)
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
	return "lgtmコマンド\n```@amesh lgtm```"
}

func (cmd LGTMCommand) Execute(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) error {
	imgurl, err := cmd.Service.Random()
	msg := inreply(event)
	if err != nil {
		msg.Text = fmt.Sprintf("LGTM: %v", err.Error())
	} else {
		msg.Text = imgurl
	}
	_, err = client.PostMessage(ctx, msg)
	return err
}
