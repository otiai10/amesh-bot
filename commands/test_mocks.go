package commands

import (
	"context"
	"fmt"

	"github.com/otiai10/amesh-bot/service"
	"github.com/slack-go/slack"
)

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

type mockSlackClient struct {
	messages []service.SlackMsg
}

func (sc *mockSlackClient) PostMessage(ctx context.Context, msg interface{}) (*slack.SlackResponse, error) {
	sc.messages = append(sc.messages, msg.(service.SlackMsg))
	return nil, nil
}
