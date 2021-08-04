package bot

import (
	"context"

	"github.com/otiai10/amesh-bot/service"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// Helpコマンドは、botに登録されたcommandsを知る必要があるので、Commandインターフェースには押し込められない.
// msgのBlocksの作り方は、以下参考:
// 		https://app.slack.com/block-kit-builder/T02N4356M#%7B%22blocks%22:%5B%7B%22type%22:%22section%22,%22text%22:%7B%22type%22:%22mrkdwn%22,%22text%22:%22%E3%83%87%E3%83%95%E3%82%A9%E3%83%AB%E3%83%88%E3%82%B3%E3%83%9E%E3%83%B3%E3%83%89%5Cn%60%60%60@amesh%20%5B-a%5D%60%60%60%22%7D%7D,%7B%22type%22:%22section%22,%22text%22:%7B%22type%22:%22mrkdwn%22,%22text%22:%22%E7%94%BB%E5%83%8F%E6%A4%9C%E7%B4%A2%E3%82%B3%E3%83%9E%E3%83%B3%E3%83%89%5Cn%60%60%60@amesh%20img%20%7Bquery%7D%60%60%60%22%7D%7D%5D%7D
func (b *Bot) Help(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) error {
	msg := service.SlackMsg{Channel: event.Channel}
	for _, cmd := range append([]Command{b.Default}, b.Commands...) {
		if cmd != nil {
			block := slack.NewSectionBlock(
				slack.NewTextBlockObject(slack.MarkdownType, cmd.Help(), false, false),
				nil, nil,
			)
			msg.Blocks = append(msg.Blocks, block)
		}
	}
	_, err := client.PostMessage(ctx, msg)
	return err
}
