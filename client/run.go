package client

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"git.mills.io/prologic/je"
)

// Run ...
func (c *Client) Run(name string, args []string, input io.Reader) (res []*je.Job, err error) {
	url := fmt.Sprintf("%s/%s?args=%s&wait=1", c.url, name, url.QueryEscape(strings.Join(args, " ")))

	return c.request("POST", url, input)
}
