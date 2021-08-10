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
func (cmd ImageCommand) Execute(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) (err error) {

	safe := "active"
	fset := largo.NewFlagSet("img", largo.ContinueOnError)
	fset.Parse(largo.Tokenize(event.Text)[2:])
	words := fset.Rest()

	rand.Seed(time.Now().Unix())
	query := strings.Join(words, "+")
	q := url.Values{}
	q.Add("q", query)
	q.Add("num", "10")
	q.Add("start", fmt.Sprintf("%d", 1+rand.Intn(10)))
	q.Add("safe", safe)
	q.Add("searchType", "image")

	res, err := cmd.Search.CustomSearch(q)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	result := new(google.CustomSearchResponse)
	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		return err
	}

	msg := service.SlackMsg{Channel: event.Channel}

	if len(result.Items) == 0 {
		msg.Blocks = append(msg.Blocks, cmd.notfoundMessageBlock(q))
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
	return "画像検索コマンド\n```@amesh img|image {query}```"
}

func (cmd ImageCommand) notfoundMessageBlock(q url.Values) slack.Block {
	q.Del("cx")
	q.Del("key")
	text := fmt.Sprintf(
		":neutral_face: 画像が見つかりませんでした: q=%s, num=%s, start=%s, safe=%s",
		q.Get("q"), q.Get("num"), q.Get("start"), q.Get("safe"),
	)
	return slack.NewContextBlock("", slack.NewTextBlockObject(slack.MarkdownType, text, false, true))
}
