package commands

import (
	"github.com/otiai10/amesh-bot/service"
	"github.com/slack-go/slack/slackevents"
)

// ameshの哲学として、
// public to public: publicメンションにはpublic投稿で返す
// thread to thread: thread内メンションにはthread内で返す
// ではあるが、forceThreadReply true を与えられた場合は、
// publicメンションにも、そのスレッドで返信するようにする.
func inreply(event slackevents.AppMentionEvent, forceThreadReply ...bool) service.SlackMsg {
	msg := service.SlackMsg{
		Channel:         event.Channel,
		ThreadTimestamp: event.ThreadTimeStamp,
	}
	if len(forceThreadReply) > 0 && forceThreadReply[0] == true {
		msg.ThreadTimestamp = event.TimeStamp
	}
	return msg
}
