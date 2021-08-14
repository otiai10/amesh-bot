package commands

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
		if strings.Contains(req.URL.Path, "/data/forecast") {
			w.Write([]byte(`[
				{},
				{
					"timeSeries": [
						{
							"timeDefines": ["2021-08-06T11:00:00+09:00", "2021-08-07T00:00:00+09:00", "2021-08-08T00:00:00+09:00"],
							"areas": [{
								"weatherCodes": ["111", "203", "300"],
								"pops": ["", "90", "50"]
							}]
						},
						{
							"areas": [{
								"tempsMin": ["", "23", "25"],
								"tempsMax": ["", "33", "33"]
							}]
						}
					]
				}
			]`))
		} else if strings.Contains(req.URL.Path, "/overview_week") {
			w.Write([]byte(`{
				"publishingOffice": "気象庁",
				"reportDatetime": "2021-08-06T10:46:00+09:00",
				"headTitle": "関東甲信地方週間天気予報",
				"text": "予報期間　８月７日から８月１３日まで\n　向こう一週間は、台風第１０号や湿った空気の影響で雲が広がりやすく、雨の降る日があるでしょう。なお、７日から８日にかけては台風第１０号の影響で荒れた天気となり、大しけとなるおそれがあります。また、台風の進路や発達の程度等によっては大雨のおそれもあります。\n　最高気温と最低気温はともに、期間の前半は平年並か平年より高い日が多いですが、期間の後半は平年並か平年より低い日が多い見込みです。\n　降水量は、平年より多いでしょう。"
			}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})
	s := httptest.NewServer(m)

	sc := new(mockSlackClient)
	cmd := ForecastCommand{SourceURL: s.URL, Timezone: timezone}
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
