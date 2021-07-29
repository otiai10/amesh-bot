package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/goapis/google"
	"github.com/otiai10/largo"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type CustomSearchClient interface {
	CustomSearch(url.Values) (*http.Response, error)
}

// ImageCommand ...
type ImageCommand struct {
	Search CustomSearchClient
}

// Match ...
func (cmd ImageCommand) Match(event slackevents.AppMentionEvent) bool {
	tokens := largo.Tokenize(event.Text)[1:]
	if len(tokens) == 0 {
		return false
	}
	return tokens[0] == "img" || tokens[0] == "image"
}

// Handle ...
func (cmd ImageCommand) Execute(ctx context.Context, client *service.SlackClient, event slackevents.AppMentionEvent) (err error) {

	safe := "active"
	fset := largo.NewFlagSet("img", largo.ContinueOnError)
	fset.Parse(largo.Tokenize(event.Text)[2:])
	words := fset.Rest()

	rand.Seed(time.Now().Unix())
	query := strings.Join(words, "+")
	q := url.Values{}
	q.Add("q", query)
	q.Add("searchType", "image")
	q.Add("num", "10")
	q.Add("start", fmt.Sprintf("%d", 1+rand.Intn(10)))
	q.Add("safe", safe)

	res, err := cmd.Search.CustomSearch(q)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	result := new(google.CustomSearchResponse)
	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		return err
	}

	msg := struct {
		Channel string        `json:"channel"`
		Text    string        `json:"text,omitempty"`
		Blocks  []slack.Block `json:"blocks,omitempty"`
	}{Channel: event.Channel}

	if len(result.Items) == 0 {
		q.Del("cx")
		q.Del("key")
		msg.Text = fmt.Sprintf("Not found for query: %v", q)
		_, err = client.PostMessage(ctx, msg)
		return err
	}

	index := rand.Intn(len(result.Items))
	item := result.Items[index]

	block := slack.NewImageBlock(item.Link, item.Title, "", nil)
	msg.Blocks = append(msg.Blocks, block)

	_, err = client.PostMessage(ctx, msg)
	return err
}

// Help ...
func (cmd ImageCommand) Help() string {
	return ""
}
