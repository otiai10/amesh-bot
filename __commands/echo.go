package commands

import (
	"context"
	"strings"

	"github.com/otiai10/amesh-bot/slack"
)

// EchoCommand ...
type EchoCommand struct{}

// Match ...
func (cmd EchoCommand) Match(payload *slack.Payload) bool {
	if len(payload.Ext.Words) == 0 {
		return false
	}
	return payload.Ext.Words[0] == "echo"
}

// Handle ...
func (cmd EchoCommand) Handle(ctx context.Context, payload *slack.Payload) *slack.Message {

	if payload.Ext.Words.Flag("-h") {
		return cmd.Help(payload)
	}

	return &slack.Message{
		Channel: payload.Event.Channel,
		Text:    strings.Join(payload.Ext.Words[1:], " "),
	}
}

// Help ...
func (cmd EchoCommand) Help(payload *slack.Payload) *slack.Message {
	return &slack.Message{
		Channel: payload.Event.Channel,
		Text:    "オウム返しコマンド\n```@amesh echo {なにかしら発言}```",
	}
}
