package commands

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	. "github.com/otiai10/mint"
)

func TestForecastCommand_Match(t *testing.T) {
	cmd := ForecastCommand{Timezone: timezone}
	m := cmd.Match(slackevents.AppMentionEvent{Text: "@amesh forecast"})
	Expect(t, m).ToBe(true)
	m = cmd.Match(slackevents.AppMentionEvent{Text: "@amesh"})
	Expect(t, m).ToBe(false)
	m = cmd.Match(slackevents.AppMentionEvent{Text: "@amesh foobaa"})
	Expect(t, m).ToBe(false)
}

func TestForecastCommand_Help(t *testing.T) {
	cmd := ForecastCommand{Timezone: timezone}
	help := cmd.Help()
	Expect(t, help).ToBe("天気予報コマンド\n```@amesh forecast|予報 {都市の名前=tokyo}```")
}

func TestForecastCommand_Execute(t *testing.T) {

	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		f, err := os.Open("../testdata" + req.URL.Path)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		}
		defer f.Close()
		io.Copy(w, f)
	})
	s := httptest.NewServer(m)

	sc := new(mockSlackClient)
	cmd := ForecastCommand{SourceURL: s.URL, Timezone: timezone}
	ctx := context.Background()
	ev := slackevents.AppMentionEvent{Text: "@amesh forecast"}
	err := cmd.Execute(ctx, sc, ev)
	Expect(t, err).ToBe(nil)

	msg := sc.messages[0]
	Expect(t, msg.Blocks[0].BlockType()).ToBe(slack.MBTSection)
	Expect(t, msg.Blocks[1].BlockType()).ToBe(slack.MBTContext)
	Expect(t, len(msg.Blocks)).ToBe(1 + 8) // 東京地方、っていうタイトル + 9日分

	day1 := msg.Blocks[1].(*slack.ContextBlock).ContextElements
	d1_e0 := day1.Elements[0].(*slack.TextBlockObject)
	Expect(t, d1_e0.Text).ToBe("08/15（日）")

	day2 := msg.Blocks[2].(*slack.ContextBlock).ContextElements
	d2_e0 := day2.Elements[0].(*slack.TextBlockObject)
	Expect(t, d2_e0.Text).ToBe("08/16（月）")

	day3 := msg.Blocks[3].(*slack.ContextBlock).ContextElements
	d3_e0 := day3.Elements[0].(*slack.TextBlockObject)
	Expect(t, d3_e0.Text).ToBe("08/17（火）")

	When(t, "-help given", func(t *testing.T) {
		sc := new(mockSlackClient)
		ev.Text = "@amesh forecast -h"
		err := cmd.Execute(ctx, sc, ev)
		Expect(t, err).ToBe(nil)
		msg := sc.messages[0]
		Expect(t, msg.Text).Match("天気予報コマンド")
	})

	When(t, "-list given", func(t *testing.T) {
		sc := new(mockSlackClient)
		ev.Text = "@amesh forecast -list"
		err := cmd.Execute(ctx, sc, ev)
		Expect(t, err).ToBe(nil)
		msg := sc.messages[0]
		Expect(t, msg.Text).Match("soya")
	})

	When(t, "city name given but not found", func(t *testing.T) {
		sc := new(mockSlackClient)
		ev.Text = "@amesh forecast newyork"
		err := cmd.Execute(ctx, sc, ev)
		Expect(t, err).ToBe(nil)
		msg := sc.messages[0]
		Expect(t, msg.Text).Match("クエリ「newyork」に対する観測所を発見できませんでした.")
	})
}
