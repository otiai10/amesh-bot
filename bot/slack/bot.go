package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/otiai10/marmoset"
	"github.com/otiai10/spell"

	"github.com/otiai10/amesh/bot/middleware"

	. "github.com/otiai10/amesh/bot/middleware"
)

var directMentionExpression = regexp.MustCompile("^<?@")

// Bot ...
type Bot struct {
	Commands []Command
}

// Command ...
type Command interface {
	Match(*Payload) bool
	Handle(context.Context, *Payload) *Message
	Help(*Payload) *Message
}

// OAuth handles oauth request from Slack.
func (bot Bot) OAuth(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	params := url.Values{
		"code":          {code},
		"client_id":     {os.Getenv("SLACK_APP_CLIENT_ID")},
		"client_secret": {os.Getenv("SLACK_APP_CLIENT_SECRET")},
	}
	req, err := http.NewRequest("POST", "https://slack.com/api/oauth.v2.access", strings.NewReader(params.Encode()))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if res.StatusCode >= 400 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	var oauth OAuthResponse
	if err := json.NewDecoder(res.Body).Decode(&oauth); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := bot.setTeam(context.Background(), oauth); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Success!")
}

// Webhook handles webhook request from Slack.
func (bot Bot) Webhook(w http.ResponseWriter, r *http.Request) {
	render := marmoset.Render(w, true)
	payload := &Payload{}
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		render.JSON(http.StatusBadRequest, marmoset.P{"message": err.Error()})
		return
	}

	if payload.Token != os.Getenv("SLACK_VERIFICATION_TOKEN") {
		render.JSON(http.StatusBadRequest, marmoset.P{"message": "invalid verification"})
		return
	}

	if payload.Type == "url_verification" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(payload.Challenge))
		return
	}

	// https://api.slack.com/events-api#the-events-api__responding-to-events
	render.JSON(http.StatusAccepted, marmoset.P{"message": "ok"})
	go bot.handle(context.Background(), payload)

	return
}

func (bot Bot) handle(ctx context.Context, payload *Payload) {

	log := Log(ctx, os.Getenv("GOOGLE_PROJECT_ID"), "amesh")
	defer log.Close()

	team, err := bot.getTeam(ctx, payload)
	if err != nil {
		log.Critical(err.Error(), Labels{"service": "firestore"})
		return
	}

	message := bot.createResponseMessage(context.Background(), payload)
	if message == nil {
		return
	}

	if err := postMessage(message, team); err != nil {
		log.Critical(err.Error(), Labels{"service": "slack"})
		return
	}
}

func (bot Bot) getTeam(ctx context.Context, payload *Payload) (team *Team, err error) {
	team = &Team{}
	kvs := middleware.KVS(ctx, os.Getenv("GOOGLE_PROJECT_ID"))
	defer kvs.Close()
	path := "Teams/" + payload.TeamID
	err = kvs.Get(path, team)
	return
}

func (bot Bot) setTeam(ctx context.Context, oauth OAuthResponse) error {
	kvs := middleware.KVS(ctx, os.Getenv("GOOGLE_PROJECT_ID"))
	defer kvs.Close()
	path := "Teams/" + oauth.Team.ID
	return kvs.Set(path, oauth)
}

func (bot Bot) createResponseMessage(ctx context.Context, payload *Payload) (message *Message) {

	if !directMentionExpression.MatchString(payload.Event.Text) {
		return nil
	}

	payload.Ext.Words = spell.Words(strings.Fields(payload.Event.Text)[1:])

	defer func() {
		if r := recover(); r != nil {
			message = &Message{
				Channel: payload.Event.Channel,
				Text:    fmt.Sprintf("🤪\n> %v\n```\n%s\n```", payload.Ext.Words, r),
				// Text: fmt.Sprintf("🤪\n> %v\n```\n%s\n```", payload.Ext.Words, debug.Stack()),
			}
		}
	}()

	for _, cmd := range bot.Commands {
		if cmd.Match(payload) {
			return cmd.Handle(ctx, payload)
		}
	}

	if payload.Ext.Words.Flag("-h") || payload.Ext.Words.Flag("help") {
		return bot.createHelpMessage(ctx, payload)
	}

	return &Message{
		Channel: payload.Event.Channel,
		Text:    fmt.Sprintf("ちょっと何言ってるかわからない\n> %v", payload.Ext.Words),
	}
}

func (bot Bot) createHelpMessage(ctx context.Context, payload *Payload) *Message {
	message := &Message{Channel: payload.Event.Channel}
	for _, cmd := range bot.Commands {
		message.Text += cmd.Help(payload).Text + "\n"
	}
	message.Text += "これ\n```@amesh [help|-h]```"
	return message
}
