package commands

import (
	"context"

	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/largo"
	"github.com/slack-go/slack/slackevents"
)

type TyphoonCommand struct{}

func (cmd TyphoonCommand) Match(event slackevents.AppMentionEvent) bool {
	tokens := largo.Tokenize(event.Text)[1:]
	if len(tokens) == 0 {
		return false
	}
	return tokens[0] == "typhoon" || tokens[0] == "台風"
}

func (cmd TyphoonCommand) Help() string {
	return "台風情報コマンド\n```@amesh typhoon|台風```"
}

func (cmd TyphoonCommand) Execute(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) error {
	_, err := client.PostMessage(ctx, service.SlackMsg{Text: "https://tenki.jp/bousai/typhoon/japan-near/", Channel: event.Channel})
	return err
}
