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
		GetChannelInfo(context.Context, string) (slack.Channel, error)
		GetThreadHistory(ctx context.Context, channel, thread string) ([]slack.Msg, error)
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

func (c *SlackClient) GetThreadHistory(ctx context.Context, channel, thread string) ([]slack.Msg, error) {
	api := slack.New(c.AccessToken)
	api.GetConversationRepliesContext(ctx, &slack.GetConversationRepliesParameters{})
	query := url.Values{"channel": []string{channel}, "ts": []string{thread}}
	req, err := http.NewRequest("GET", "https://slack.com/api/conversations.replies?"+query.Encode(), nil)
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
	response := struct {
		OK       bool
		Messages []slack.Msg
	}{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}
	if !response.OK {
		// @see https://github.com/slack-go/slack/issues/939
		errres := slack.SlackResponse{}
		return nil, fmt.Errorf("%s", errres.Error)
	}
	return response.Messages, nil
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

func (c *SlackClient) GetChannelInfo(ctx context.Context, id string) (info slack.Channel, err error) {
	req, err := http.NewRequest("GET", "https://slack.com/api/conversations.info?channel="+id, nil)
	if err != nil {
		// return err
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AccessToken))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return info, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return info, fmt.Errorf(res.Status)
	}

	response := struct {
		OK      bool
		Channel slack.Channel
	}{Channel: info}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return info, err
	}
	if !response.OK {
		// @see https://github.com/slack-go/slack/issues/939
		errres := slack.SlackResponse{}
		return info, fmt.Errorf("%s", errres.Error)
	}

	return response.Channel, nil
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
