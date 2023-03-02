package commands

import (
	"bytes"
	"context"
	"encoding/json"
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
	return strings.HasPrefix(event.Text, "<@") // Only replies to direct mentions.
}

func (cmd AICompletion) Execute(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) (err error) {
	msg := inreply(event, true)

	help := bytes.NewBuffer(nil)
	dump := false
	fset := largo.NewFlagSet("ai chat", largo.ContinueOnError)
	fset.BoolVar(&dump, "dump", false, "詳細情報を表示します").Alias("d")
	fset.Parse(largo.Tokenize(event.Text[1:]))
	fset.Output = help

	tokens := fset.Rest()
	ai := &openaigo.Client{APIKey: cmd.APIKey, BaseURL: cmd.BaseURL}
	res, err := ai.Chat(ctx, openaigo.ChatCompletionRequestBody{
		Model: "gpt-3.5-turbo",
		Messages: []openaigo.ChatMessage{
			{Role: "user", Content: strings.Join(tokens, " ")},
		},
		MaxTokens: 1024,
		User:      fmt.Sprintf("%s:%s", event.Channel, event.TimeStamp),
	})
	if err != nil {
		openaistatuspage := "https://status.openai.com/"
		msg.Text = fmt.Sprintf(":pleading_face: %v", openaistatuspage)
		_, foerr := client.PostMessage(ctx, msg)
		return fmt.Errorf("openai.Ask failed with: %v (and error on failover: %v)", err, foerr)
	}
	if len(res.Choices) == 0 {
		nferr := NotFound{}.Execute(ctx, client, event)
		return fmt.Errorf("openai.Ask returns zero choice (and NotFound Cmd error: %v)", nferr)
	}
	msg.Text = res.Choices[rand.Intn(len(res.Choices))].Message.Content

	if dump {
		buf := bytes.NewBuffer(nil)
		enc := json.NewEncoder(buf)
		enc.SetIndent("", "  ")
		if err := enc.Encode(res); err != nil {
			msg.Text += "\n[dump]\n```\n" + buf.String() + "\n```"
		}
	}

	_, err = client.PostMessage(ctx, msg)
	return err
}

func (cmd AICompletion) Help() string {
	return ""
}
