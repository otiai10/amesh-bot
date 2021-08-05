package commands

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/slack-go/slack/slackevents"

	. "github.com/otiai10/mint"
)

func TestForecastCommand_Match(t *testing.T) {
	cmd := ForecastCommand{}
	m := cmd.Match(slackevents.AppMentionEvent{Text: "@amesh forecast"})
	Expect(t, m).ToBe(true)
	m = cmd.Match(slackevents.AppMentionEvent{Text: "@amesh"})
	Expect(t, m).ToBe(false)
	m = cmd.Match(slackevents.AppMentionEvent{Text: "@amesh foobaa"})
	Expect(t, m).ToBe(false)
}

func TestForecastCommand_Help(t *testing.T) {
	cmd := ForecastCommand{}
	help := cmd.Help()
	Expect(t, help).ToBe("天気予報コマンド\n```@amesh forecast|予報 {都市の名前=tokyo}```")
}

func TestForecastCommand_Execute(t *testing.T) {

	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(`[
			{},
			{
				"timeSeries": [
					{
						"areas": [
							{}
						]
					},
					{
						"areas": [
							{}
						]
					}
				]
			}
		]`))
	})
	s := httptest.NewServer(m)

	sc := new(mockSlackClient)
	cmd := ForecastCommand{SourceURL: s.URL}
	ctx := context.Background()
	ev := slackevents.AppMentionEvent{Text: "@amesh forecast"}
	err := cmd.Execute(ctx, sc, ev)
	Expect(t, err).ToBe(nil)

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
