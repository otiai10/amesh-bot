package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/otiai10/amesh-bot/bot"
	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/goapis/google"
	"github.com/otiai10/largo"
	"github.com/slack-go/slack/slackevents"
)

type GoogleCommand struct {
	Search CustomSearchClient
}

// Match ...
func (cmd GoogleCommand) Match(event slackevents.AppMentionEvent) bool {
	tokens := largo.Tokenize(event.Text)[1:]
	if len(tokens) == 0 {
		return false
	}
	return tokens[0] == "ggl" || tokens[0] == "google"
}

// Handle ...
func (cmd GoogleCommand) Execute(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) *bot.CommandError {

	safe := "active"
	fset := largo.NewFlagSet("google", largo.ContinueOnError)
	fset.Parse(largo.Tokenize(event.Text)[2:])
	words := fset.Rest()

	rand.Seed(time.Now().Unix())
	query := strings.Join(words, "+")
	q := url.Values{}
	q.Add("q", query)
	q.Add("num", "10")
	q.Add("start", fmt.Sprintf("%d", 1+rand.Intn(10)))
	q.Add("safe", safe)

	res, err := cmd.Search.CustomSearch(q)
	if err != nil {
		return commandError(err)
	}
	defer res.Body.Close()

	result := new(google.CustomSearchResponse)
	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		return commandError(err)
	}

	msg := inreply(event)

	if len(result.Items) == 0 {
		q.Del("cx")
		q.Del("key")
		msg.Text = fmt.Sprintf("Not found for query: %v", q)
		if _, err := client.PostMessage(ctx, msg); err != nil {
			return commandError(err)
		}
		return nil
	}

	index := rand.Intn(len(result.Items))
	item := result.Items[index]

	msg.Text = fmt.Sprintf("> %s\n%s\n", query, item.Link)

	if _, err := client.PostMessage(ctx, msg); err != nil {
		return commandError(err)
	}
	return nil
}

// Help ...
func (cmd GoogleCommand) Help() string {
	return "グーグル検索コマンド\n```@amesh ggl|google {query}```"
}
