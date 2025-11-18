package bot

import (
	"fmt"

	"github.com/slack-go/slack/slackevents"
)

const defaultCommandErrorMessage = ":warning: コマンドの実行に失敗しました。時間をおいて再試行してください。"

// CommandError は Slack 応答とログ用の情報を併せ持つエラーコンテナ。
type CommandError struct {
	Err             string      `json:"err"`
	CmdName         string      `json:"cmd_name"`
	Event           interface{} `json:"event"`
	Message         string      `json:"message"`
	ThreadTimestamp string      `json:"thread_ts,omitempty"`
}

// NewCommandError は CommandError を生成するヘルパー。
func NewCommandError(err error, message string) *CommandError {
	if err == nil {
		return nil
	}
	if message == "" {
		message = defaultCommandErrorMessage
	}
	return &CommandError{Err: err.Error(), Message: message}
}

// withContext はコマンド名やイベント情報を後付けする。
func (cmderr *CommandError) withContext(cmd interface{}, event interface{}) *CommandError {
	if cmderr == nil {
		return nil
	}
	if cmderr.CmdName == "" {
		cmderr.CmdName = commandName(cmd)
	}
	if cmderr.Event == nil {
		cmderr.Event = event
	}
	if cmderr.Message == "" {
		cmderr.Message = defaultCommandErrorMessage
	}
	if cmderr.ThreadTimestamp == "" {
		if ev, ok := event.(slackevents.AppMentionEvent); ok {
			cmderr.ThreadTimestamp = ev.ThreadTimeStamp
		}
	}
	return cmderr
}

func wrapWithContext(cmderr *CommandError, cmd interface{}, event interface{}) *CommandError {
	if cmderr == nil {
		return nil
	}
	return cmderr.withContext(cmd, event)
}

func commandName(cmd interface{}) string {
	switch v := cmd.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%T", cmd)
	}
}

func (cmderr *CommandError) labels() map[string]string {
	return map[string]string{
		"command": cmderr.CmdName,
	}
}
