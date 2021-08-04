package controllers

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/otiai10/amesh-bot/service"
	. "github.com/otiai10/mint"
	"github.com/slack-go/slack/slackevents"
)

type (
	mockBot       struct{}
	mockSlack     struct{}
	mockDatastore struct{}
)

func (mb *mockBot) Handle(ctx context.Context, oauth service.OAuthResponse, event slackevents.AppMentionEvent) {

}

func (ms *mockSlack) ExchangeOAuthCodeWithAccessToken(context.Context, string) (*http.Response, error) {
	res := &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{}`))}
	return res, nil
}

func (md *mockDatastore) Set(ctx context.Context, path string, val interface{}) error {
	return nil
}
func (md *mockDatastore) Get(ctx context.Context, path string, dest interface{}) error {
	return nil
}

func TestController_OAuth(t *testing.T) {

	bot := &mockBot{}
	slack := &mockSlack{}
	datastore := &mockDatastore{}

	c := &Controller{
		Bot:       bot,
		Slack:     slack,
		Datastore: datastore,
	}

	s := httptest.NewServer(nil)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", s.URL+"?code=xxx", nil)
	c.OAuth(rec, req)
	Expect(t, rec.Code).ToBe(http.StatusOK)
}

func TestController_Webhook(t *testing.T) {
	c := &Controller{Bot: &mockBot{}, Slack: &mockSlack{}, Datastore: &mockDatastore{}}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`{}`))

	c.Webhook(rec, req)
	Expect(t, rec.Code).ToBe(http.StatusAccepted)
}
