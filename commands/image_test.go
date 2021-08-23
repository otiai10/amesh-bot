package commands

import (
	"context"
	"testing"

	. "github.com/otiai10/mint"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func TestImageCommand_Match(t *testing.T) {
	cmd := ImageCommand{}
	m := cmd.Match(slackevents.AppMentionEvent{Text: "@amesh img foo"})
	Expect(t, m).ToBe(true)
	m = cmd.Match(slackevents.AppMentionEvent{Text: "@amesh"})
	Expect(t, m).ToBe(false)
	m = cmd.Match(slackevents.AppMentionEvent{Text: "@amesh foobaa"})
	Expect(t, m).ToBe(false)
}

func TestImageCommand_Help(t *testing.T) {
	cmd := ImageCommand{}
	help := cmd.Help()
	Expect(t, help).ToBe("画像検索コマンド\n```@amesh img|image {query}```")
}

func TestImageCommand_Execute(t *testing.T) {
	sc := &mockSlackClient{}
	search := &mockGoogleClient{}
	search.ResponseBody = `{
		"items": [
			{"title":"hoge", "link":"qwerty"}
		],
		"queries": {"request": [{}]}
	}`
	ctx := context.Background()
	cmd := ImageCommand{Search: search}
	event := slackevents.AppMentionEvent{Text: "@amesh img pikachu"}
	err := cmd.Execute(ctx, sc, event)
	Expect(t, err).ToBe(nil)

	msg := sc.messages[0]
	Expect(t, msg.Blocks[0].(*slack.ImageBlock).ImageURL).ToBe("qwerty")

	When(t, "no item found", func(t *testing.T) {
		sc := &mockSlackClient{}
		search.ResponseBody = `{
			"items": [],
			"queries": {"request": [{"startIndex": 45}]}
		}`
		cmd.Search = search
		err := cmd.Execute(ctx, sc, event)
		Expect(t, err).ToBe(nil)
	})

	When(t, "help requested", func(t *testing.T) {
		sc := &mockSlackClient{}
		event.Text = "@amesh img -h"
		err := cmd.Execute(ctx, sc, event)
		Expect(t, err).ToBe(nil)
		Expect(t, sc.messages[0].Text).Match("NAME\n  img")
	})

	When(t, "unsafe given", func(t *testing.T) {
		sc := &mockSlackClient{}
		event.Text = "@amesh img foobaa -unsafe"
		err := cmd.Execute(ctx, sc, event)
		Expect(t, err).ToBe(nil)
	})

	When(t, "verbose given", func(t *testing.T) {
		sc := &mockSlackClient{}
		search.ResponseBody = `{
			"items": [
				{
				  "kind": "customsearch#result",
				  "title": "環境養殖技術開発センター",
				  "htmlTitle": "環境養殖\u003cb\u003e技術\u003c/b\u003e開発センター",
				  "link": "https://www.pref.nagasaki.jp/shared/uploads/2018/11/1543213119.pdf",
				  "displayLink": "www.pref.nagasaki.jp",
				  "snippet": "「水産業普及指導センター. 長崎市. 果水產振興課. 県畜産課. 科学技術振興課. II. 養殖衛生管理指導. III. 養殖場の調査・監視. 1. 水産用医薬品の適正使用指導.",
				  "htmlSnippet": "「水産業普及指導センター. \u003cb\u003e長崎市\u003c/b\u003e. 果水產振興課. 県畜産課. \u003cb\u003e科学技術\u003c/b\u003e振興課. II. 養殖衛生管理指導. III. 養殖場の調査・監視. 1. 水産用医薬品の適正使用指導.",
				  "cacheId": "0mOemwWwAroJ",
				  "formattedUrl": "https://www.pref.nagasaki.jp/shared/uploads/2018/11/1543213119.pdf",
				  "htmlFormattedUrl": "https://www.pref.nagasaki.jp/shared/uploads/2018/11/1543213119.pdf",
				  "pagemap": {
					"metatags": [
					  {
						"moddate": "Fri Nov 23 04:41:22 2018",
						"creationdate": "Fri Nov 23 04:41:22 2018",
						"producer": "iText® 5.3.2 ©2000-2012 1T3XT BVBA (AGPL-version)"
					  }
					]
				  },
				  "mime": "application/pdf",
				  "fileFormat": "PDF/Adobe Acrobat"
				}
			],
			"queries": {
				"request": [
				  {
					"title": "Google Custom Search - \"長崎市科学技術\"",
					"totalResults": "1",
					"searchTerms": "\"長崎市科学技術\"",
					"count": 1,
					"startIndex": 1,
					"inputEncoding": "utf8",
					"outputEncoding": "utf8",
					"safe": "off",
					"cx": "015322470755990492660:xfvoaurleaw",
					"orTerms": "長崎市科学技術"
				  }
				]
			}
		}`
		event.Text = "@amesh img foobaa -v"
		err := cmd.Execute(ctx, sc, event)
		Expect(t, err).ToBe(nil)
	})
}
