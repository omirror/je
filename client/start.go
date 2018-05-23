package client

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"git.mills.io/prologic/je"
)

// Start ...
func (c *Client) Start(name string, args []string, input io.Reader) (res []*je.Job, err error) {
	url := fmt.Sprintf("%s/%s?args=%s", c.url, name, url.QueryEscape(strings.Join(args, " ")))

	return c.request("POST", url, input)
}
