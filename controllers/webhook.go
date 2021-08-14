package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/marmoset"
	"github.com/slack-go/slack/slackevents"
)

func (c *Controller) Webhook(w http.ResponseWriter, req *http.Request) {

	render := marmoset.Render(w, true)

	payload := Payload{}
	defer req.Body.Close()

	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		render.JSON(http.StatusBadRequest, marmoset.P{"message": err.Error()})
		return
	}

	if payload.Token != os.Getenv("SLACK_VERIFICATION_TOKEN") {
		render.JSON(http.StatusBadRequest, marmoset.P{"message": "invalid verification"})
		return
	}

	if payload.Type == slackevents.URLVerification {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(payload.Challenge))
		return
	}

	// OK, it's valid event callback
	// https://api.slack.com/events-api#the-events-api__responding-to-events
	render.JSON(http.StatusAccepted, marmoset.P{"message": "ok"})

	// Fetch oauth information and recover Slack client
	team := service.OAuthResponse{}
	key := fmt.Sprintf("Teams/%s", payload.TeamID)
	if err := c.Datastore.Get(req.Context(), key, &team); err != nil {
		// TODO: Fix
		panic(fmt.Errorf("datastore.Get failed: %v", err))
	}

	if payload.APIAppID == os.Getenv("DEV_SLACK_APP_ID") {
		team.AccessToken = os.Getenv("DEV_SLACK_BOT_USER_OAUTH_TOKEN")
	}

	// TODO: BotがRequestの情報を参照できるよう、Contextに含める
	ctx := context.WithValue(context.Background(), "webhook_request", req)
	go c.Bot.Handle(ctx, team, payload.Event)

	return

}
