package client

import (
	"net/url"
	"strings"
)

// JoinArgs ...
func JoinArgs(args []string) string {
	return url.QueryEscape(strings.Join(args, " "))
}
