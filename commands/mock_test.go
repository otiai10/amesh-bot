package commands

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/mint"
)

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

func (ml *mockLGTM) Random() (string, error) {
	if ml.err != nil {
		return "", ml.err
	}
	return ml.imgurl, nil
}
