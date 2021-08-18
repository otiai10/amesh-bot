package commands

import (
	"github.com/otiai10/amesh-bot/service"
	"github.com/slack-go/slack/slackevents"
)

func inreply(event slackevents.AppMentionEvent) service.SlackMsg {
	return service.SlackMsg{
		Channel:         event.Channel,
		ThreadTimestamp: event.ThreadTimeStamp,
	}
}
