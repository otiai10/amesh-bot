package controllers

import (
	"context"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type (
	Payload struct {
		slackevents.EventsAPIEvent
		slackevents.ChallengeResponse
		Event slackevents.AppMentionEvent
	}
	Bot interface {
		Handle(ctx context.Context, team slack.OAuthV2Response, event slackevents.AppMentionEvent)
	}
	Slack interface {
		ExchangeOAuthCodeWithAccessToken(ctx context.Context, code string) (*http.Response, error)
	}
	Datastore interface {
		Set(ctx context.Context, path string, val interface{}) error
		Get(ctx context.Context, path string, dest interface{}) error
	}
)

type Controller struct {
	Bot       Bot
	Slack     Slack
	Datastore Datastore
}
