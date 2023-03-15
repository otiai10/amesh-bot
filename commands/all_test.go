package commands

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/mint"
	"github.com/slack-go/slack"
)

var (
	timezone *time.Location
)

func TestMain(m *testing.M) {
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Fatalf("failed to load location: %v", err)
	}
	timezone = tokyo
	code := m.Run()
	os.Exit(code)
}

// CloudStorage
type mockStorage struct{}

func (mocks *mockStorage) Exists(ctx context.Context, bucket, name string) (bool, error) {
	return false, nil
	// return true, nil
}

func (mocks *mockStorage) URL(bucket, name string) string {
	return fmt.Sprintf("%s/%s", bucket, name)
}

func (mocks *mockStorage) Upload(ctx context.Context, bucket, name string, contents []byte) error {
	return nil
}

// ISlackClient
type mockSlackClient struct {
	messages []service.SlackMsg
}

func (sc *mockSlackClient) PostMessage(ctx context.Context, msg interface{}) (*service.PostMessageResponse, error) {
	sc.messages = append(sc.messages, msg.(service.SlackMsg))
	return nil, nil
}

// func (sc *mockSlackClient) DeleteMessage(ctx context.Context, msg interface{}) error {
// 	return nil
// }

func (sc *mockSlackClient) UpdateMessage(ctx context.Context, msg interface{}) error {
	return nil
}

func (sc *mockSlackClient) GetChannelInfo(ctx context.Context, id string) (info slack.Channel, err error) {
	return
}

func (sc *mockSlackClient) GetThreadHistory(ctx context.Context, channel, thread string) (a []slack.Msg, e error) {
	return
}

// Google
type mockGoogleClient struct {
	mint.HTTPClientMock
}

func (gc *mockGoogleClient) CustomSearch(url.Values) (*http.Response, error) {
	res, err, ok := gc.Handle()
	if ok {
		return res, err
	}
	return nil, fmt.Errorf("invalid mocking")
}

// LGTM
type mockLGTM struct {
	imgurl string
	err    error
}

func (ml *mockLGTM) Random() (string, string, error) {
	if ml.err != nil {
		return "", "", ml.err
	}
	return ml.imgurl, "![xx](zz)", nil
}
