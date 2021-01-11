package commands

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/otiai10/amesh-bot/slack"
	"github.com/otiai10/goapis/google"
)

// ImageCommand ...
type ImageCommand struct{}

// Match ...
func (cmd ImageCommand) Match(payload *slack.Payload) bool {
	if len(payload.Ext.Words) == 0 {
		return false
	}
	return payload.Ext.Words[0] == "img" || payload.Ext.Words[0] == "image"
}

// Handle ...
func (cmd ImageCommand) Handle(ctx context.Context, payload *slack.Payload) *slack.Message {
	client := google.Client{
		APIKey:               os.Getenv("GOOGLE_CUSTOMSEARCH_API_KEY"),
		CustomSearchEngineID: os.Getenv("GOOGLE_CUSTOMSEARCH_ENGINE_ID"),
	}

	verbose := false
	if payload.Ext.Words.Flag("-v") {
		payload.Ext.Words = payload.Ext.Words.Remove("-v", 0)
		verbose = true
	}

	words := payload.Ext.Words[1:]
	if len(words) == 0 {
		return wrapError(payload, ErrorGoogleNoQueryGiven)
	}

	rand.Seed(time.Now().Unix())
	query := strings.Join(words, "+")
	q := url.Values{}
	q.Add("q", query)
	q.Add("searchType", "image")
	q.Add("num", "5")
	q.Add("start", fmt.Sprintf("%d", 1+rand.Intn(10)))

	res, err := client.CustomSearch(q)
	if err != nil {
		return wrapError(payload, err)
	}
	if len(res.Items) == 0 {
		return wrapError(payload, ErrorGoogleNotFound)
	}

	// TODO: ランダムにひとつ選ぶ
	index := rand.Intn(len(res.Items))
	item := res.Items[index]

	block := slack.Block{
		Type:     "image",
		ImageURL: item.Link,
		AltText:  query,
		Title:    createImageTitle(verbose, q, len(res.Items), index, item),
	}

	return &slack.Message{Channel: payload.Event.Channel, Blocks: []slack.Block{block}}
}

// Help ...
func (cmd ImageCommand) Help(payload *slack.Payload) *slack.Message {
	return &slack.Message{
		Channel: payload.Event.Channel,
		Text:    "画像検索コマンド\n```@amesh [image|img] {検索キーワード}```",
	}
}

func createImageTitle(verbose bool, q url.Values, found, randIndex int, item google.CustomSearchItem) *slack.Element {
	if !verbose {
		return nil
	}
	// {{{ サニタイズ
	q.Del("key")
	q.Del("cx")
	q.Del("searchType")
	// }}}
	lines := []string{
		"query: " + q.Get("q"),
		"context: " + item.Image.ContextLink,
	}
	q.Del("q")

	lines = append(
		lines,
		fmt.Sprintf("offset: %s, count: %s, found: %d, rand: %d", q.Get("start"), q.Get("num"), found, randIndex),
	)
	return &slack.Element{
		Type: "plain_text",
		Text: strings.Join(lines, "\n"),
	}
}
