package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/otiai10/amesh-bot/service"
)

var (
	successHTML = `<!DOCTYPE html><html>
		<head>
			<meta name="viewport" content="width=device-width,initial-scale=1.0,minimum-scale=1.0">
			<style>
				body { display: flex; flex-direction: column; align-items: center; font-family: Helvetica; text-align: center; }
				code { background-color: #f0f0f0; padding: 4px; font-weight: bold; } a { color: #5E58C7; } footer { margin: 32px; }
			</style>
		</head>
		<body>
			<h1>amesh is successfully installed!</h1>
			<div>Invite <b>@amesh</b> to your channel and mention <code>@amesh help</code> 🤖</div>
			<a href="https://app.slack.com/client/%s">Back to your Slack.</a>
			<footer>Know more about <a href="https://github.com/otiai10/amesh-bot">amesh-bot</a> on GitHub.</footer>
		</body>
	</html>`
)

func (c *Controller) OAuth(w http.ResponseWriter, req *http.Request) {
	code := req.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// params := url.Values{
	// 	"code":          {code},
	// 	"client_id":     {os.Getenv("SLACK_APP_CLIENT_ID")},
	// 	"client_secret": {os.Getenv("SLACK_APP_CLIENT_SECRET")},
	// }
	// req, err := http.NewRequest("POST", "https://slack.com/api/oauth.v2.access", strings.NewReader(params.Encode()))
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// res, err := http.DefaultClient.Do(req)

	res, err := c.Slack.ExchangeOAuthCodeWithAccessToken(req.Context(), code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if res.StatusCode >= 400 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	oauth := service.OAuthResponse{}
	if err := json.NewDecoder(res.Body).Decode(&oauth); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	key := fmt.Sprintf("Teams/%s", oauth.Team.ID)
	if err := c.Datastore.Set(req.Context(), key, oauth); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf8")
	fmt.Fprintf(w, successHTML, oauth.Team.ID)

}
