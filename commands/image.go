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

	if payload.Ext.Words.Flag("-h") {
		return cmd.Help(payload)
	}

	client := google.Client{
		APIKey:               os.Getenv("GOOGLE_CUSTOMSEARCH_API_KEY"),
		CustomSearchEngineID: os.Getenv("GOOGLE_CUSTOMSEARCH_ENGINE_ID"),
		Eager:                true, // 検索結果が0の場合、条件を変えて再検索をかける
	}

	verbose := false
	if payload.Ext.Words.Flag("-v") {
		payload.Ext.Words = payload.Ext.Words.Remove("-v", 0)
		verbose = true
	}

	safe := "active"
	if payload.Ext.Words.Flag("-unsafe") {
		payload.Ext.Words = payload.Ext.Words.Remove("-unsafe", 0)
		safe = "off"
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
	q.Add("num", "10")
	q.Add("start", fmt.Sprintf("%d", 1+rand.Intn(10)))
	q.Add("safe", safe)

	res, err := client.CustomSearch(q)
	if err != nil {
		text := strings.Join(cmd.searchMetaInfo(q, 0, 0), "\n")
		return &slack.Message{Channel: payload.Event.Channel, Text: fmt.Sprintf("%v\n> %s", err, text)}
	}
	if len(res.Items) == 0 {
		text := strings.Join(cmd.searchMetaInfo(q, 0, 0), "\n")
		return &slack.Message{Channel: payload.Event.Channel, Text: "Not Found\n> " + text}
	}

	index := rand.Intn(len(res.Items))
	item := res.Items[index]

	var title *slack.Element
	if verbose {
		lines := append(cmd.searchMetaInfo(q, len(res.Items), index), cmd.searchResultItemInfo(&item)...)
		title = &slack.Element{Type: "plain_text", Text: strings.Join(lines, "\n")}
	}

	block := slack.Block{
		Type:     "image",
		ImageURL: item.Link,
		AltText:  query,
		Title:    title,
	}

	return &slack.Message{Channel: payload.Event.Channel, Blocks: []slack.Block{block}}
}

// Help ...
func (cmd ImageCommand) Help(payload *slack.Payload) *slack.Message {
	return &slack.Message{
		Channel: payload.Event.Channel,
		Text:    "画像検索コマンド\n```@amesh [image|img] {検索キーワード} [-h|-v|-unsafe]```",
	}
}

// 基本的な検索情報をテキストにするやつ
func (cmd ImageCommand) searchMetaInfo(q url.Values, found, index int) (lines []string) {
	return []string{
		"query:\t" + q.Get("q"),
		fmt.Sprintf(
			"num: %s, start: %s, safe: %s, found: %d, random: %d",
			q.Get("num"), q.Get("start"), q.Get("safe"), found, index,
		),
	}
}

// アイテムが見つかったときの結果verboseをテキストにするやつ
func (cmd ImageCommand) searchResultItemInfo(item *google.CustomSearchItem) (lines []string) {
	if item == nil {
		return []string{}
	}
	return []string{
		fmt.Sprintf("context:\t%s", item.Image.ContextLink),
		fmt.Sprintf("title:\t%s", item.HTMLTitle),
	}
}
