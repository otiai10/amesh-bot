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

	ISlackClient interface {
		PostMessage(ctx context.Context, msg interface{}) (*PostMessageResponse, error)
		// DeleteMessage(ctx context.Context, msg interface{}) error
		UpdateMessage(ctx context.Context, msg interface{}) error
	}

	SlackMsg struct {
		Channel         string        `json:"channel"`
		Text            string        `json:"text,omitempty"`
		Blocks          []slack.Block `json:"blocks,omitempty"`
		Timestamp       string        `json:"ts,omitempty"`
		ThreadTimestamp string        `json:"thread_ts,omitempty"`
		UnfurlMedia     *bool         `json:"unfurl_media,omitempty"`
		// UnfurlLinks  *bool         `json:"unfurl_links,omitempty"`
	}

	// OAuthResponse ...
	// https://api.slack.com/methods/oauth.v2.access#response
	OAuthResponse struct {
		OK         bool   `json:"ok"     firestore:"ok"`
		AppID      string `json:"app_id" firestore:"app_id"`
		AuthedUser struct {
			ID string `json:"id" firestore:"id"`
		} `json:"authed_user" firestore:"authed_user"`
		Scope       string `json:"scope"        firestore:"scope"`
		TokenType   string `json:"token_type"   firestore:"token_type"`
		AccessToken string `json:"access_token" firestore:"access_token"`
		BotUserID   string `json:"bot_user_id"  firestore:"bot_user_id"`
		Team        struct {
			ID   string `json:"id"   firestore:"id"`
			Name string `json:"name" firestore:"name"`
		} `json:"team" firestore:"team"`
		Enterprise interface{} `json:"enterprise" firestore:"-"`
	}

	PostMessageResponse struct {
		slack.SlackResponse
		slack.Msg
	}
)

func NewSlackClient(accessToken string) *SlackClient {
	return &SlackClient{
		AccessToken: accessToken,
	}
}

func (c *SlackClient) PostMessage(ctx context.Context, msg interface{}) (*PostMessageResponse, error) {
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

	response := &PostMessageResponse{}
	if err := json.NewDecoder(res.Body).Decode(response); err != nil {
		return nil, err
	}
	if !response.Ok {
		// @see https://github.com/slack-go/slack/issues/939
		buf := bytes.NewBuffer(nil)
		json.NewEncoder(buf).Encode(response)
		return response, fmt.Errorf("%s: %s", response.Error, buf.String())
	}
	return response, nil
}

// https://api.slack.com/methods/chat.delete
func (c *SlackClient) DeleteMessage(ctx context.Context, msg interface{}) error {

	body := bytes.NewBuffer(nil)
	if err := json.NewEncoder(body).Encode(msg); err != nil {
		return err
	}

	// https://api.slack.com/methods/chat.delete
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.delete", body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf(res.Status)
	}

	response := &slack.SlackResponse{}
	if err := json.NewDecoder(res.Body).Decode(response); err != nil {
		return err
	}
	if !response.Ok {
		// @see https://github.com/slack-go/slack/issues/939
		buf := bytes.NewBuffer(nil)
		json.NewEncoder(buf).Encode(response)
		return fmt.Errorf("%s: %s", response.Error, buf.String())
	}
	return nil

}

func (c *SlackClient) UpdateMessage(ctx context.Context, msg interface{}) error {

	body := bytes.NewBuffer(nil)
	if err := json.NewEncoder(body).Encode(msg); err != nil {
		return err
	}

	// https://api.slack.com/methods/chat.update
	req, err := http.NewRequest("POST", "https://slack.com/api/chat.update", body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf(res.Status)
	}

	response := &slack.SlackResponse{}
	if err := json.NewDecoder(res.Body).Decode(response); err != nil {
		return err
	}
	if !response.Ok {
		// @see https://github.com/slack-go/slack/issues/939
		buf := bytes.NewBuffer(nil)
		json.NewEncoder(buf).Encode(response)
		return fmt.Errorf("%s: %s", response.Error, buf.String())
	}
	return nil

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
