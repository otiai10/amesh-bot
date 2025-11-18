package commands

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/otiai10/amesh-bot/bot"
	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/largo"
	"github.com/otiai10/openaigo"
	"github.com/slack-go/slack/slackevents"
)

type AICompletion struct {
	APIKey  string
	BaseURL string
}

var (
	channelChatModeOnmemoryCache = map[string]string{}
)

const (
	mentionPrefix    = "<@"
	mentionSuffix    = ">"
	openaiMaxContext = 2048
	openaiPricingURL = "https://openai.com/pricing"
	openaiStatusURL  = "https://status.openai.com/"
)

func (cmd AICompletion) getChannelTopic(ctx context.Context, client service.ISlackClient, id string) (string, error) {
	if val, ok := channelChatModeOnmemoryCache[id]; ok {
		fmt.Println("[INFO] topic cache hit for channel id: " + id)
		return val, nil
	}
	info, err := client.GetChannelInfo(ctx, id)
	if err != nil {
		return "", nil
	}
	channelChatModeOnmemoryCache[id] = info.Topic.Value
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

	messages := []openaigo.Message{}
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
			role := "user"
			if m.User == myself {
				role = "assistant"
			}
			cleaned := strings.ReplaceAll(m.Text, myid, "")
			messages = append(messages, openaigo.Message{Role: role, Content: cleaned})
			// if total+len(cleaned) > openaiMaxContext {
			// 	total -= len(messages[1].Content)
			// 	messages = append(messages[:1], messages[1:]...)
			// }
		}
	} else {
		messages = append(messages, openaigo.Message{Role: "user", Content: strings.Join(tokens, "\n")})
	}

	ai := &openaigo.Client{APIKey: cmd.APIKey, BaseURL: cmd.BaseURL}
	res, err := ai.Chat(ctx, openaigo.ChatRequest{
		Model:     openaigo.GPT4o,
		Messages:  messages,
		MaxTokens: openaiMaxContext,
		User:      fmt.Sprintf("%s:%s", event.Channel, event.TimeStamp),
	})
	if err != nil {
		text := fmt.Sprintf(":pleading_face: %v", openaiStatusURL)
		if e, ok := err.(openaigo.APIError); ok {
			text = fmt.Sprintf(":pleading_face: %s\n%v", e.Message, openaiPricingURL)
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
	if len(res.Choices) == 0 {
		cerr := commandErrorWithMessage(
			fmt.Errorf("openai Chat returns zero choice"),
			":warning: 返答を生成できませんでした。もう一度お試しください。",
		)
		if cerr != nil && forceThreadReply && event.ThreadTimeStamp == "" {
			cerr.ThreadTimestamp = event.TimeStamp
		}
		return cerr
	}
	msg.Text = res.Choices[rand.Intn(len(res.Choices))].Message.Content
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
