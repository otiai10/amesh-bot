package commands

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/ja"
	"github.com/otiai10/jma"
	"github.com/otiai10/jma/api"
	"github.com/otiai10/largo"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type ForecastCommand struct {
	SourceURL string
	Timezone  *time.Location
}

func (cmd ForecastCommand) Match(event slackevents.AppMentionEvent) bool {
	tokens := largo.Tokenize(event.Text)
	if len(tokens) < 2 {
		return false
	}
	if tokens[1] == "forecast" || tokens[1] == "予報" {
		return true
	}
	return false
}

func (cmd ForecastCommand) Execute(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) error {
	city := "tokyo"
	list := false
	help := bytes.NewBuffer(nil)
	fset := largo.NewFlagSet("forecast", largo.ContinueOnError)
	fset.BoolVar(&list, "list", false, "対応都市・観測所のリスト")
	fset.Output = help

	msg := service.SlackMsg{Channel: event.Channel}

	if err := fset.Parse(largo.Tokenize(event.Text)[2:]); err != nil {
		msg.Text = err.Error()
		_, err = client.PostMessage(ctx, msg)
		return err
	}

	if fset.HelpRequested() {
		msg.Text = fmt.Sprintf("天気予報コマンド\n```@amesh forecast {都市の名前=tokyo} [-list]\n%v```", help.String())
		_, err := client.PostMessage(ctx, msg)
		return err
	}

	if list {
		for _, o := range jma.Offices {
			msg.Text += fmt.Sprintf("%v %v\n", o.NameEnLower, o.OfficeName)
		}
		_, err := client.PostMessage(ctx, msg)
		return err
	}

	if rest := fset.Rest(); len(rest) > 0 {
		city = rest[0]
	}

	areas := jma.SearchOffice(city)
	if len(areas) == 0 {
		msg.Text = fmt.Sprintf("クエリ「%s」に対する観測所を発見できませんでした.\n以下のコマンドを試してください.\n```@amesh forecast -list```\n", city)
		_, err := client.PostMessage(ctx, msg)
		return err
	}

	jmaclient := &api.Client{BaseURL: cmd.SourceURL}
	entries, err := jmaclient.Forecast(areas[0].Code)

	if err != nil {
		msg.Text = err.Error()
		_, err := client.PostMessage(ctx, msg)
		return err
	}
	overview, _ := jmaclient.Overview(areas[0].Code)
	// Overviewは無くていいので、エラーは無視する.
	// if err != nil {
	// 	msg.Text = err.Error()
	// 	_, err := client.PostMessage(ctx, msg)
	// 	return err
	// }

	blocks := cmd.FormatForecastToSlackBlocks(entries, overview)
	msg.Blocks = blocks

	// json.NewEncoder(os.Stderr).Encode(msg)

	_, err = client.PostMessage(ctx, msg)
	if err != nil {
		msg.Text = err.Error()
		_, err := client.PostMessage(ctx, msg)
		return err
	}

	return nil
}

func (cmd ForecastCommand) Help() string {
	return "天気予報コマンド\n```@amesh forecast|予報 {都市の名前=tokyo}```"
}

// https://app.slack.com/block-kit-builder/T02N4356M#%7B%22blocks%22:%5B%7B%22type%22:%22context%22,%22elements%22:%5B%7B%22type%22:%22plain_text%22,%22text%22:%22%E4%BA%88%E5%A0%B1%E6%9C%9F%E9%96%93%E3%80%80%EF%BC%98%E6%9C%88%EF%BC%93%E6%97%A5%E3%81%8B%E3%82%89%EF%BC%98%E6%9C%88%EF%BC%99%E6%97%A5%E3%81%BE%E3%81%A7%5Cn%E5%90%91%E3%81%93%E3%81%86%E4%B8%80%E9%80%B1%E9%96%93%E3%81%AF%E3%80%81%E6%9C%9F%E9%96%93%E3%81%AE%E5%89%8D%E5%8D%8A%E3%81%AF%E9%AB%98%E6%B0%97%E5%9C%A7%E3%81%AB%E8%A6%86%E3%82%8F%E3%82%8C%E3%81%A6%E6%99%B4%E3%82%8C%E3%82%8B%E6%97%A5%E3%82%82%E3%81%82%E3%82%8A%E3%81%BE%E3%81%99%E3%81%8C%E3%80%81%E6%B0%97%E5%9C%A7%E3%81%AE%E8%B0%B7%E3%82%84%E6%B9%BF%E3%81%A3%E3%81%9F%E7%A9%BA%E6%B0%97%E3%81%AE%E5%BD%B1%E9%9F%BF%E3%81%A7%E9%9B%B2%E3%81%8C%E5%BA%83%E3%81%8C%E3%82%8A%E3%82%84%E3%81%99%E3%81%84%E3%81%A7%E3%81%97%E3%82%87%E3%81%86%E3%80%82%E6%9C%80%E9%AB%98%E6%B0%97%E6%B8%A9%E3%81%A8%E6%9C%80%E4%BD%8E%E6%B0%97%E6%B8%A9%E3%81%AF%E3%81%A8%E3%82%82%E3%81%AB%E3%80%81%E5%B9%B3%E5%B9%B4%E4%B8%A6%E3%81%8B%E5%B9%B3%E5%B9%B4%E3%82%88%E3%82%8A%E9%AB%98%E3%81%84%E8%A6%8B%E8%BE%BC%E3%81%BF%E3%81%A7%E3%81%99%E3%80%82%E7%86%B1%E4%B8%AD%E7%97%87%E3%81%AA%E3%81%A9%E5%81%A5%E5%BA%B7%E7%AE%A1%E7%90%86%E3%81%AB%E6%B3%A8%E6%84%8F%E3%81%97%E3%81%A6%E3%81%8F%E3%81%A0%E3%81%95%E3%81%84%E3%80%82%E9%99%8D%E6%B0%B4%E9%87%8F%E3%81%AF%E3%80%81%E5%B9%B3%E5%B9%B4%E4%B8%A6%E3%81%8B%E5%B9%B3%E5%B9%B4%E3%82%88%E3%82%8A%E5%B0%91%E3%81%AA%E3%81%84%E3%81%A7%E3%81%97%E3%82%87%E3%81%86%E3%80%82%22,%22emoji%22:true%7D%5D%7D,%7B%22type%22:%22section%22,%22fields%22:%5B%7B%22type%22:%22mrkdwn%22,%22text%22:%22*%E6%9D%B1%E4%BA%AC%E5%9C%B0%E6%96%B9*%22%7D%5D%7D,%7B%22type%22:%22context%22,%22elements%22:%5B%7B%22type%22:%22mrkdwn%22,%22text%22:%228%E6%9C%8802%E6%97%A5%EF%BC%88%E6%9C%88%E6%9B%9C%EF%BC%89%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22:sunny:%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%2233/25%E2%84%83%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22%7C%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22-/-/50/20%22%7D%5D%7D,%7B%22type%22:%22context%22,%22elements%22:%5B%7B%22type%22:%22mrkdwn%22,%22text%22:%228%E6%9C%8802%E6%97%A5%EF%BC%88%E6%9C%88%E6%9B%9C%EF%BC%89%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22:sunny:%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%2233/25%E2%84%83%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22%7C%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%2210/20/10/10%22%7D%5D%7D,%7B%22type%22:%22context%22,%22elements%22:%5B%7B%22type%22:%22mrkdwn%22,%22text%22:%228%E6%9C%8802%E6%97%A5%EF%BC%88%E6%9C%88%E6%9B%9C%EF%BC%89%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22:sunny:%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%2233/25%E2%84%83%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22%7C%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%2220%22%7D%5D%7D,%7B%22type%22:%22context%22,%22elements%22:%5B%7B%22type%22:%22mrkdwn%22,%22text%22:%228%E6%9C%8802%E6%97%A5%EF%BC%88%E6%9C%88%E6%9B%9C%EF%BC%89%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22:sunny:%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%2233/25%E2%84%83%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22%7C%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%2230%22%7D%5D%7D,%7B%22type%22:%22context%22,%22elements%22:%5B%7B%22type%22:%22mrkdwn%22,%22text%22:%228%E6%9C%8802%E6%97%A5%EF%BC%88%E6%9C%88%E6%9B%9C%EF%BC%89%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22:sunny:%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%2233/25%E2%84%83%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22%7C%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%2210%22%7D%5D%7D,%7B%22type%22:%22context%22,%22elements%22:%5B%7B%22type%22:%22mrkdwn%22,%22text%22:%228%E6%9C%8802%E6%97%A5%EF%BC%88%E6%9C%88%E6%9B%9C%EF%BC%89%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22:sunny:%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%2233/25%E2%84%83%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22%7C%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%2220%22%7D%5D%7D,%7B%22type%22:%22context%22,%22elements%22:%5B%7B%22type%22:%22mrkdwn%22,%22text%22:%228%E6%9C%8802%E6%97%A5%EF%BC%88%E6%B0%B4%E6%9B%9C%EF%BC%89%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22:sunny:%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%2233/25%E2%84%83%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%22%7C%22%7D,%7B%22type%22:%22mrkdwn%22,%22text%22:%2230%22%7D%5D%7D%5D%7D
func (cmd ForecastCommand) FormatForecastToSlackBlocks(entries []api.ComprehensiveForecastEntry, overview *api.OverviewWeek) (blocks []slack.Block) {

	k := 0 // まずは一番うえのAreaだけ見る, which means 伊豆諸島・小笠原諸島を無視している

	// {{{ エリア「k」における情報をまず抽出
	weekly := entries[1]
	codes := weekly.TimeSeries[0].Areas[k].WeatherCodes
	pops := weekly.TimeSeries[0].Areas[k].Pops
	temps := weekly.TimeSeries[1].Areas[k]
	area := weekly.TimeSeries[0].Areas[k]
	// }}}

	// 地域タイトル
	title := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*%s*", area.Area.Name), false, false)
	blocks = append(blocks, slack.NewSectionBlock(title, nil, nil))

	rows := []Row{}

	// {{{ 最初の数日分を作成
	// 実はweeklyは、明日からの予報しかないうえに、明日の予報の情報を欠損している.
	// したがって、entries[0]を解釈し、今日と明日の予報を生成する必要がある.
	coming := entries[0]
	for i, t := range coming.TimeSeries[0].TimeDefines {
		row := Row{
			Date:    t,
			Weather: jma.Weathers[coming.TimeSeries[0].Areas[k].WeatherCodes[i]],
		}
		rows = append(rows, row)
	}
	// }}}

	// {{{ 最初の数日の最高気温最低気温を補完
	for i, t := range coming.TimeSeries[2].TimeDefines {
		tmps := coming.TimeSeries[2].Areas[k].Temps
		for j, r := range rows {
			if DateDecimal(r.Date) == DateDecimal(t) {
				// 同じ日付のエントリだが、最高気温なのか最低気温なのか、判断基準が無い
				if r.TempMax == "" { // とりあえず埋める
					rows[j].TempMax, rows[j].TempMin = tmps[i], tmps[i]
				} else if r.TempMax < tmps[i] { // 来たものがMaxより上ならMax
					rows[j].TempMax = tmps[i]
				} else if r.TempMin > tmps[i] { // 来たものがMinより下ならMin
					rows[j].TempMin = tmps[i]
				}
			}
		}
	}
	// }}}

	// {{{ 最初の数日の降水確率を補完
	for i, t := range coming.TimeSeries[1].TimeDefines {
		pops := coming.TimeSeries[1].Areas[k].Pops
		for j, r := range rows {
			if DateDecimal(r.Date) == DateDecimal(t) {
				// 降水確率は時系列なので、単純にその日の降水確率にappendでよい
				rows[j].PoPs = append(rows[j].PoPs, pops[i])
			}
		}
	}
	// }}}

	// 7日間分作成. k == 0 で固定していることに注意
	// 1日を1行（= 1block）で表現している
	for i, t := range weekly.TimeSeries[0].TimeDefines {
		// すでに上記のブロックで最初の数日登録されている場合があるので、チェック
		done := false
		for j, r := range rows {
			if DateDecimal(r.Date) == DateDecimal(t) {
				// 補完だけする
				if r.TempMax == "" || r.TempMin == "" {
					rows[j].TempMax = temps.TempsMax[i]
					rows[j].TempMin = temps.TempsMin[i]
				}
				if len(r.PoPs) == 0 {
					rows[j].PoPs = append(r.PoPs, pops[i])
				}
				done = true
				break
			}
		}
		if !done {
			rows = append(rows, Row{
				Date:    t,
				Weather: jma.Weathers[codes[i]],
				TempMax: temps.TempsMax[i],
				TempMin: temps.TempsMin[i],
				PoPs:    []string{pops[i]},
			})
		}
	}

	for _, row := range rows {
		blocks = append(blocks, row.ToSlackBlock())
	}

	// 広域のOverviewをヘッドラインとして表示
	if overview != nil {
		chunks := ja.Cut(overview.Text, true)
		text := "> " + chunks[0] + "\n" + "> " + strings.Join(chunks[1:], "")
		headline := slack.NewTextBlockObject(slack.MarkdownType, text, false, false)
		blocks = append(blocks, slack.NewContextBlock("", headline))
	}

	return blocks
}

type (
	Row struct {
		Date    time.Time
		Weather jma.Weather
		TempMax string
		TempMin string
		PoPs    []string
	}
)

func (row Row) ToSlackBlock() *slack.ContextBlock {
	date := row.Date.Format("01/02") + fmt.Sprintf("（%s）", ja.Weekday[row.Date.Weekday()])
	columns := []slack.MixedElement{
		slack.NewTextBlockObject(slack.MarkdownType, date, false, false),                    // 日付
		slack.NewTextBlockObject(slack.MarkdownType, row.Weather.Emoji.Slack, false, false), // 天気emoji
	}
	if row.TempMax != "" || row.TempMin != "" { // 気温を追加
		t := fmt.Sprintf("%s/%s℃", row.TempMax, row.TempMin)
		columns = append(columns, slack.NewTextBlockObject(slack.MarkdownType, t, false, false))
	}
	if len(row.PoPs) != 0 { // 降水確率を追加
		for i, pop := range row.PoPs {
			row.PoPs[i] = pop + "%"
		}
		columns = append(columns, slack.NewTextBlockObject(slack.MarkdownType, strings.Join(row.PoPs, "/"), false, false))
	}
	return slack.NewContextBlock("", columns...)
}

// 2021/08/15 -> 20210815
func DateDecimal(t time.Time) int {
	y, m, d := t.Date()
	return y*10000 + int(m)*100 + d
}
