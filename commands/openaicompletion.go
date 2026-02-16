package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared"
	"github.com/otiai10/amesh-bot/bot"
	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/largo"
	"github.com/slack-go/slack/slackevents"
)

type AICompletion struct {
	APIKey  string
	BaseURL string
}

var (
	channelChatModeOnmemoryCache sync.Map
)

const (
	mentionPrefix    = "<@"
	mentionSuffix    = ">"
	openaiPricingURL = "https://openai.com/pricing"
	openaiStatusURL  = "https://status.openai.com/"

	// GPT-5-nano max output tokens: 128k まで対応。
	// Slack Bot用途では過剰なため 16k に設定。
	// https://platform.openai.com/docs/models/gpt-5-nano
	openaiMaxOutputTokens = int64(16000)
	// GPT-5-nano: GPT-5ファミリーの最軽量モデル。
	// https://platform.openai.com/docs/models/gpt-5-nano
	openaiDefaultModel = "gpt-5-nano"
)

func (cmd AICompletion) getChannelTopic(ctx context.Context, client service.ISlackClient, id string) (string, error) {
	if val, ok := channelChatModeOnmemoryCache.Load(id); ok {
		fmt.Println("[INFO] topic cache hit for channel id: " + id)
		return val.(string), nil
	}
	info, err := client.GetChannelInfo(ctx, id)
	if err != nil {
		return "", nil
	}
	channelChatModeOnmemoryCache.Store(id, info.Topic.Value)
	return info.Topic.Value, nil
}

func (cmd AICompletion) shouldForceThreadReply(ctx context.Context, client service.ISlackClient, channelID string) (bool, error) {
	topic, err := cmd.getChannelTopic(ctx, client, channelID)
	if err != nil {
		return true, err
	}
	if strings.Contains(topic, "-amesh-chat-mode=flat") {
		return false, nil
	}
	return true, nil
}

// Match ...
func (cmd AICompletion) Match(event slackevents.AppMentionEvent) bool {
	return strings.HasPrefix(event.Text, mentionPrefix) // Only replies to direct mentions.
}

func (cmd AICompletion) Execute(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) *bot.CommandError {

	forceThreadReply, err := cmd.shouldForceThreadReply(ctx, client, event.Channel)
	if err != nil {
		return commandError(err)
	}
	msg := inreply(event, forceThreadReply)

	tokens := largo.Tokenize(event.Text)[1:]

	inputItems := []responses.ResponseInputItemUnionParam{}
	// Thread内の会話なので、会話コンテキストを取得しにいく
	if event.ThreadTimeStamp != "" {
		myself := event.Text[len(mentionPrefix):strings.Index(event.Text, mentionSuffix)]
		myid := mentionPrefix + myself + mentionSuffix
		history, err := client.GetThreadHistory(ctx, event.Channel, event.ThreadTimeStamp)
		if err != nil {
			return commandErrorWithMessage(
				fmt.Errorf("slack: failed to fetch thread history: %v", err),
				":warning: スレッドの会話履歴を取得できませんでした。時間をおいて再試行してください。",
			)
		}
		for _, m := range history {
			role := responses.EasyInputMessageRoleUser
			if m.User == myself {
				role = responses.EasyInputMessageRoleAssistant
			}
			cleaned := strings.TrimSpace(strings.ReplaceAll(m.Text, myid, ""))
			if cleaned == "" {
				continue
			}
			inputItems = append(inputItems, responses.ResponseInputItemParamOfMessage(cleaned, role))
		}
	} else {
		prompt := strings.TrimSpace(strings.Join(tokens, "\n"))
		if prompt != "" {
			inputItems = append(inputItems, responses.ResponseInputItemParamOfMessage(prompt, responses.EasyInputMessageRoleUser))
		}
	}
	if len(inputItems) == 0 {
		return commandErrorWithMessage(
			fmt.Errorf("openai: no valid content to send"),
			":warning: 送信するテキストが見つかりませんでした。メッセージ内容を含めてメンションしてください。",
		)
	}

	options := []option.RequestOption{}
	if cmd.APIKey != "" {
		options = append(options, option.WithAPIKey(cmd.APIKey))
	}
	if cmd.BaseURL != "" {
		options = append(options, option.WithBaseURL(cmd.BaseURL))
	}
	ai := openai.NewClient(options...)
	res, err := ai.Responses.New(ctx, responses.ResponseNewParams{
		Model: shared.ResponsesModel(openaiDefaultModel),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: responses.ResponseInputParam(inputItems),
		},
		Instructions:    openai.String(openaiSlackInstructions),
		MaxOutputTokens: openai.Int(openaiMaxOutputTokens),
		User:            openai.String(fmt.Sprintf("%s:%s", event.Channel, event.TimeStamp)),
	})
	if err != nil {
		text := fmt.Sprintf(":pleading_face: %v", openaiStatusURL)
		var apierr *openai.Error
		if errors.As(err, &apierr) && apierr.Message != "" {
			text = fmt.Sprintf(":pleading_face: %s\n%v", apierr.Message, openaiPricingURL)
		}
		cerr := commandError(err)
		if cerr != nil {
			cerr.Message = text
			if forceThreadReply && event.ThreadTimeStamp == "" {
				cerr.ThreadTimestamp = event.TimeStamp
			}
		}
		return cerr
	}
	output := strings.TrimSpace(res.OutputText())
	if output == "" {
		cerr := commandErrorWithMessage(
			fmt.Errorf("openai Responses returned empty output"),
			":warning: 返答を生成できませんでした。もう一度お試しください。",
		)
		if cerr != nil && forceThreadReply && event.ThreadTimeStamp == "" {
			cerr.ThreadTimestamp = event.TimeStamp
		}
		return cerr
	}
	markdown := true
	msg.Mrkdwn = &markdown
	msg.Text = output
	if _, err := client.PostMessage(ctx, msg); err != nil {
		cerr := commandError(err)
		if cerr != nil && forceThreadReply && event.ThreadTimeStamp == "" {
			cerr.ThreadTimestamp = event.TimeStamp
		}
		return cerr
	}
	return nil
}

func (cmd AICompletion) Help() string {
	return ""
}
