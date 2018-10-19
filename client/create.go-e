package client

import (
	"fmt"
	"io"

	"git.mills.io/prologic/je"
)

// Create ...
func (c *Client) Create(name string, args []string, input io.Reader, interactive, wait bool) (res []*je.Job, err error) {

	url := fmt.Sprintf("%s/create/%s?args=%s", c.url, name, JoinArgs(args))

	if interactive {
		url += "&interactive=1"
	}

	if wait {
		url += "&wait=1"
	}

	return c.request("POST", url, input)
}
