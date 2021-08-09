package bot

import "fmt"

type CommandError struct {
	Err     string // `json:"err"`
	CmdName string // `json:"cmd_name"`
	// Cmd     interface{} // `json:"cmd"`
	Event   interface{} // `json:"event"`
	Message interface{} // `json:"message,omitempty"`
}

func errwrap(err error, cmd interface{}, event interface{}) *CommandError {
	if err == nil {
		return nil
	}
	switch v := cmd.(type) {
	case string:
		return &CommandError{Err: err.Error(), CmdName: v, Event: event}
	default:
		return &CommandError{Err: err.Error(), CmdName: fmt.Sprintf("%T", cmd), Event: event}
	}
}

func (cmderr *CommandError) labels() map[string]string {
	return map[string]string{
		"command": cmderr.CmdName,
	}
}
