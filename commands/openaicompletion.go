package commands

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

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
	openaiMaxContext = 2048 // 4096
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

func (cmd AICompletion) Execute(ctx context.Context, client service.ISlackClient, event slackevents.AppMentionEvent) (err error) {

	forceThreadReply, err := cmd.shouldForceThreadReply(ctx, client, event.Channel)
	if err != nil {
		return err
	}
	msg := inreply(event, forceThreadReply)

	tokens := largo.Tokenize(event.Text)[1:]

	messages := []openaigo.ChatMessage{}
	// Thread内の会話なので、会話コンテキストを取得しにいく
	if event.ThreadTimeStamp != "" {
		myself := event.Text[len(mentionPrefix):strings.Index(event.Text, mentionSuffix)]
		myid := mentionPrefix + myself + mentionSuffix
		history, err := client.GetThreadHistory(ctx, event.Channel, event.ThreadTimeStamp)
		if err != nil {
			return fmt.Errorf("slack: failed to fetch thread history: %v", err)
		}
		total := 0
		for _, m := range history {
			role := "user"
			if m.User == myself {
				role = "assistant"
			}
			cleaned := strings.ReplaceAll(m.Text, myid, "")
			messages = append(messages, openaigo.ChatMessage{Role: role, Content: cleaned})
			if total+len(cleaned) > openaiMaxContext {
				total -= len(messages[1].Content)
				messages = append(messages[:1], messages[1:]...)
			}
		}
	} else {
		messages = append(messages, openaigo.ChatMessage{Role: "user", Content: strings.Join(tokens, "\n")})
	}

	ai := &openaigo.Client{APIKey: cmd.APIKey, BaseURL: cmd.BaseURL}
	res, err := ai.Chat(ctx, openaigo.ChatCompletionRequestBody{
		Model:     "gpt-3.5-turbo",
		Messages:  messages,
		MaxTokens: openaiMaxContext * 2,
		User:      fmt.Sprintf("%s:%s", event.Channel, event.TimeStamp),
	})
	if err != nil {
		if e, ok := err.(openaigo.APIError); ok {
			msg.Text = fmt.Sprintf(":pleading_face: %s\n%v", e.Message, openaiPricingURL)
		} else {
			msg.Text = fmt.Sprintf(":pleading_face: %v", openaiStatusURL)
		}
		_, foerr := client.PostMessage(ctx, msg)
		return fmt.Errorf("openai.Ask failed with: %v (and error on failover: %v)", err, foerr)
	}
	if len(res.Choices) == 0 {
		nferr := NotFound{}.Execute(ctx, client, event)
		return fmt.Errorf("openai.Ask returns zero choice (and NotFound Cmd error: %v)", nferr)
	}
	msg.Text = res.Choices[rand.Intn(len(res.Choices))].Message.Content
	_, err = client.PostMessage(ctx, msg)
	return err
}

func (cmd AICompletion) Help() string {
	return ""
}
