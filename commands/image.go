package commands

import (
	"bytes"
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

	help := bytes.NewBuffer(nil)
	unsafe := false
	verbose := false
	filter := false // TODO: 今はBoolVarだが、他のfilterにも対応した
	flevel := 60
	fset := largo.NewFlagSet("img", largo.ContinueOnError)
	fset.Description = "画像検索コマンド"
	fset.BoolVar(&verbose, "verbose", false, "検索のverboseログを表示します").Alias("v")
	fset.BoolVar(&unsafe, "unsafe", false, "セーフサーチを無効にした検索をします").Alias("U")
	fset.BoolVar(&filter, "filter", false, "画像をフィルタ処理して表示します（今はモザイクだけ対応）").Alias("F")
	fset.IntVar(&flevel, "level", 60, "画像フィルタの強さ").Alias("L")
	fset.Output = help
	fset.Parse(largo.Tokenize(event.Text)[2:])
	words := fset.Rest()

	msg := inreply(event)
	if fset.HelpRequested() {
		msg.Text = "```" + help.String() + "```"
		_, err := client.PostMessage(ctx, msg)
		return err
	}

	rand.Seed(time.Now().Unix())
	query := strings.Join(words, "+")
	q := url.Values{}
	q.Add("q", query)
	q.Add("num", "10")
	q.Add("start", fmt.Sprintf("%d", 1+rand.Intn(10)))
	if unsafe {
		q.Add("safe", "off")
	} else {
		q.Add("safe", "active")
	}
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

	if len(result.Items) == 0 {
		msg.Blocks = append(msg.Blocks, cmd.notfoundMessageBlock(q))
		_, err = client.PostMessage(ctx, msg)
		return err
	}

	index := rand.Intn(len(result.Items))
	item := result.Items[index]

	link := item.Link
	title := item.Title

	if req, ok := ctx.Value("webhook_request").(*http.Request); ok && filter {
		u, _ := url.Parse("https://" + req.Host + "/image")
		u.RawQuery = url.Values{"url": {link}, "level": {fmt.Sprintf("%d", flevel)}}.Encode()
		// AppEngine上のProxyサーバのエンドポイントを向かせる
		// あとの流れは、controllers.Imageを参照
		link = u.String()
	}

	block := slack.NewImageBlock(link, title, "", slack.NewTextBlockObject(
		slack.PlainTextType, item.Title, false, false,
	))
	msg.Blocks = append(msg.Blocks, block)

	if verbose {
		msg.Blocks = append(msg.Blocks, slack.NewContextBlock("",
			slack.NewTextBlockObject(
				slack.MarkdownType,
				item.Image.ContextLink+"\n"+cmd.formatQueryMetadata(q),
				false, false,
			),
		))
	}

	sent, err := client.PostMessage(ctx, msg)
	// FIXME: slack-imgs.comのproxy errorが出るとすればここだと思う

	if err != nil {
		return err
	}

	// filterリクエストの場合は、自分の投稿に、unfilterなリンクを返す
	if filter {
		unfurl := false
		_, err = client.PostMessage(ctx, service.SlackMsg{
			Channel:         event.Channel,
			ThreadTimestamp: sent.Timestamp,
			Text:            ":warning: " + item.Link,
			UnfurlMedia:     &unfurl,
		})
	}
	return err
}

// Help ...
func (cmd ImageCommand) Help() string {
	return "画像検索コマンド\n```@amesh img|image {query}```"
}

func (cmd ImageCommand) notfoundMessageBlock(q url.Values) slack.Block {
	q.Del("cx")
	q.Del("key")
	text := ":neutral_face: 画像が見つかりませんでした: " + cmd.formatQueryMetadata(q)
	return slack.NewContextBlock("", slack.NewTextBlockObject(slack.MarkdownType, text, false, true))
}

func (cmd ImageCommand) formatQueryMetadata(q url.Values) string {
	return fmt.Sprintf(
		"q=%s, num=%s, start=%s, safe=%s",
		q.Get("q"), q.Get("num"), q.Get("start"), q.Get("safe"),
	)
}
