package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/otiai10/largo"

	"github.com/otiai10/amesh-bot/service"
	"github.com/slack-go/slack/slackevents"
)

// AmeshCommand ...
type AmeshCommand struct{}

// Match ...
func (cmd AmeshCommand) Match(ev slackevents.AppMentionEvent) bool {
	// 第1要素はメンションUIDの部分だと思うので
	tokens := largo.Tokenize(ev.Text)[1:]
	if len(tokens) == 0 {
		return true
	}
	return false
}

func (cmd AmeshCommand) Execute(ctx context.Context, client *service.SlackClient, event slackevents.AppMentionEvent) error {
	// tokens := largo.Tokenize(event.Text)
	words := strings.Fields(event.Text)[1:]
	// fmt.Printf("[DEBUG:001:tokens] %v\n", strings.Join(tokens, "<>"))
	fmt.Println("[DEBUG:003:bytes] %v\n", []byte(event.Text))
	res, err := client.PostMessage(ctx, struct {
		Channel string `json:"channel"`
		Text    string `json:"text"`
	}{Channel: event.Channel, Text: strings.Join(words, " ")})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", res)
	return nil
}

func (cmd AmeshCommand) Help() string {
	return ""
}
