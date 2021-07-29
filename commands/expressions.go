package commands

import "regexp"

var (
	uidexp = regexp.MustCompile("<@[a-zA-Z0-9]+>")
)
