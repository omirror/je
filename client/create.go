package client

import (
	"fmt"
	"io"

	"git.mills.io/prologic/je"
)

// Create ...
func (c *Client) Create(name string, args []string, input io.Reader, wait bool) (res []*je.Job, err error) {
	var url string

	if wait {
		url = fmt.Sprintf("%s/%s?args=%s&wait=1", c.url, name, JoinArgs(args))
	} else {
		url = fmt.Sprintf("%s/create/%s?args=%s", c.url, name, JoinArgs(args))
	}

	return c.request("POST", url, input)
}
