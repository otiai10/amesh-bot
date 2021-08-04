package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/otiai10/amesh-bot/service"
	. "github.com/otiai10/mint"
	"github.com/slack-go/slack/slackevents"
)

type (
	mockBot   struct{}
	mockSlack struct {
		exchangeerr    error
		exchangestatus int
		exchangeresstr string
	}
	mockDatastore struct {
		seterr error
		geterr error
	}
)

func (mb *mockBot) Handle(ctx context.Context, oauth service.OAuthResponse, event slackevents.AppMentionEvent) {

}

func (ms *mockSlack) ExchangeOAuthCodeWithAccessToken(context.Context, string) (*http.Response, error) {
	if ms.exchangeerr != nil {
		return nil, ms.exchangeerr
	}
	res := &http.Response{}
	if ms.exchangeresstr != "" {
		res.Body = ioutil.NopCloser(bytes.NewBufferString(ms.exchangeresstr))
	} else {
		res.Body = ioutil.NopCloser(bytes.NewBufferString(`{}`))
	}
	if ms.exchangestatus != 0 {
		res.StatusCode = ms.exchangestatus
	}
	return res, nil
}

func (md *mockDatastore) Set(ctx context.Context, path string, val interface{}) error {
	if md.seterr != nil {
		return md.seterr
	}
	return nil
}
func (md *mockDatastore) Get(ctx context.Context, path string, dest interface{}) error {
	if md.geterr != nil {
		return md.geterr
	}
	return nil
}

func TestController_OAuth(t *testing.T) {

	bot := &mockBot{}
	slack := &mockSlack{}
	datastore := &mockDatastore{}

	c := &Controller{Bot: bot, Slack: slack, Datastore: datastore}

	s := httptest.NewServer(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", s.URL+"?code=xxx", nil)

	c.OAuth(rec, req)
	Expect(t, rec.Code).ToBe(http.StatusOK)

	When(t, "code doesn't exist", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", s.URL+"?code=", nil)
		c.OAuth(rec, req)
		Expect(t, rec.Code).ToBe(http.StatusBadRequest)
	})
	When(t, "Exchange access_token failed", func(t *testing.T) {
		slack := &mockSlack{exchangeerr: fmt.Errorf("testtest")}
		c.Slack = slack
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", s.URL+"?code=xxx", nil)
		c.OAuth(rec, req)
		Expect(t, rec.Code).ToBe(http.StatusInternalServerError)

		slack = &mockSlack{exchangestatus: 402}
		c.Slack = slack
		rec = httptest.NewRecorder()
		req = httptest.NewRequest("POST", s.URL+"?code=xxx", nil)
		c.OAuth(rec, req)
		Expect(t, rec.Code).ToBe(http.StatusPaymentRequired)

		slack = &mockSlack{exchangeresstr: `invalid_json`}
		c.Slack = slack
		rec = httptest.NewRecorder()
		req = httptest.NewRequest("POST", s.URL+"?code=xxx", nil)
		c.OAuth(rec, req)
		Expect(t, rec.Code).ToBe(http.StatusBadRequest)
	})
	When(t, "datastore.set failed", func(t *testing.T) {
		bot := &mockBot{}
		slack := &mockSlack{}
		datastore := &mockDatastore{seterr: fmt.Errorf("too bad")}
		c := &Controller{Bot: bot, Slack: slack, Datastore: datastore}
		s := httptest.NewServer(nil)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", s.URL+"?code=xxx", nil)
		c.OAuth(rec, req)
		Expect(t, rec.Code).ToBe(http.StatusInternalServerError)
	})
}

func TestController_Webhook(t *testing.T) {
	c := &Controller{Bot: &mockBot{}, Slack: &mockSlack{}, Datastore: &mockDatastore{}}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`{}`))

	c.Webhook(rec, req)
	Expect(t, rec.Code).ToBe(http.StatusAccepted)

	When(t, "invalid request body", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`""`))
		c.Webhook(rec, req)
		Expect(t, rec.Code).Not().ToBe(http.StatusAccepted)
	})
	When(t, "wrong verification token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"token": "WRONG_TOKEN"}`))
		c.Webhook(rec, req)
		Expect(t, rec.Code).Not().ToBe(http.StatusAccepted)
	})
	When(t, "url_verification challenge given", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"type": "url_verification", "challenge":"xxxx"}`))
		c.Webhook(rec, req)
		Expect(t, rec.Code).ToBe(http.StatusOK)
		b, err := ioutil.ReadAll(rec.Body)
		Expect(t, err).ToBe(nil)
		Expect(t, string(b)).ToBe("xxxx")
	})
	When(t, "datastore.get failed", func(t *testing.T) {
		defer func() {
			Expect(t, recover()).Not().ToBe(nil)
		}()
		md := &mockDatastore{geterr: fmt.Errorf("too bad")}
		c.Datastore = md
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`{}`))
		c.Webhook(rec, req)
	})
}
