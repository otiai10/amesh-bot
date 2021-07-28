package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/slack-go/slack"
)

type (
	SlackClient struct {
		AccessToken string
	}
	SlackOAuthClient struct{}
)

func NewSlackClient(accessToken string) *SlackClient {
	return &SlackClient{
		AccessToken: accessToken,
	}
}

func (c *SlackClient) PostMessage(ctx context.Context, msg interface{}) (*slack.SlackResponse, error) {
	body := bytes.NewBuffer(nil)
	if err := json.NewEncoder(body).Encode(msg); err != nil {
		return nil, err
	}

	// https://api.slack.com/methods/chat.postMessage
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf(res.Status)
	}

	response := &slack.SlackResponse{}
	if err := json.NewDecoder(res.Body).Decode(response); err != nil {
		return nil, err
	}
	if !response.Ok {
		// TODO: Improve
		return response, fmt.Errorf(response.Error)
	}
	return response, nil
}

func (o *SlackOAuthClient) ExchangeOAuthCodeWithAccessToken(ctx context.Context, code string) (*http.Response, error) {
	params := url.Values{
		"code":          {code},
		"client_id":     {os.Getenv("SLACK_APP_CLIENT_ID")},
		"client_secret": {os.Getenv("SLACK_APP_CLIENT_SECRET")},
	}
	req, err := http.NewRequest("POST", "https://slack.com/api/oauth.v2.access", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
