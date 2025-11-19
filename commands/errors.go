package commands

import "github.com/otiai10/amesh-bot/bot"

const defaultUserErrorMessage = ":warning: コマンドの実行に失敗しました。時間をおいて再試行してください。"

func commandError(err error) *bot.CommandError {
	return bot.NewCommandError(err, defaultUserErrorMessage)
}

func commandErrorWithMessage(err error, message string) *bot.CommandError {
	if message == "" {
		message = defaultUserErrorMessage
	}
	return bot.NewCommandError(err, message)
}
