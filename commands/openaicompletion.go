package commands

import (
	"context"
	"math/rand"

	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/goapis/openai"
	"github.com/otiai10/largo"
	"github.com/slack-go/slack/slackevents"
)

type AICompletion struct {
	APIKey  string
	BaseURL string
}

// Match ...
func (cmd AICompletion) Match(event slackevents.AppMentionEvent) bool {
	return true
}

func (cmd AICompletion) Execute(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) (err error) {
	msg := inreply(event)
	tokens := largo.Tokenize(event.Text)[1:]
	ai := openai.Client{APIKey: cmd.APIKey, BaseURL: cmd.BaseURL}
	res, err := ai.Ask(openai.Davinci, tokens)
	if err != nil || len(res.Choices) == 0 {
		return NotFound{}.Execute(ctx, client, event)
	}
	msg.Text = res.Choices[rand.Intn(len(res.Choices))].Text
	_, err = client.PostMessage(ctx, msg)
	return err
}

func (cmd AICompletion) Help() string {
	return ""
}
