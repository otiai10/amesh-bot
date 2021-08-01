package bot

import "fmt"

type CommandError struct {
	Err     error
	CmdName string
	Cmd     interface{} // Command
	Event   interface{} // slackevents.AppMentionEvent
	Message interface{} // service.SlackMsg
}

func errwrap(err error, cmd interface{}, event interface{}) *CommandError {
	if err == nil {
		return nil
	}
	switch v := cmd.(type) {
	case string:
		return &CommandError{Err: err, CmdName: v, Cmd: cmd, Event: event}
	default:
		return &CommandError{Err: err, CmdName: fmt.Sprintf("%T", cmd), Cmd: cmd, Event: event}
	}
}

func (cmderr *CommandError) labels() map[string]string {
	return map[string]string{
		"command": cmderr.CmdName,
	}
}
