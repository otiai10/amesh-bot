package commands

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/largo"
	"github.com/otiai10/openaigo"
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
	ai := &openaigo.Client{APIKey: cmd.APIKey, BaseURL: cmd.BaseURL}
	res, err := ai.Completion(ctx, openaigo.CompletionRequestBody{
		Prompt:    []string{strings.Join(tokens, "\n")},
		Model:     "text-davinci-003",
		MaxTokens: 1024,
	})
	if err != nil {
		nferr := NotFound{}.Execute(ctx, client, event)
		return fmt.Errorf("openai.Ask failed with: %v (and NotFound Cmd error: %v)", err, nferr)
	}
	if len(res.Choices) == 0 {
		nferr := NotFound{}.Execute(ctx, client, event)
		return fmt.Errorf("openai.Ask returns zero choice (and NotFound Cmd error: %v)", nferr)
	}
	msg.Text = res.Choices[rand.Intn(len(res.Choices))].Text
	_, err = client.PostMessage(ctx, msg)
	return err
}

func (cmd AICompletion) Help() string {
	return ""
}
