package client

import (
	"fmt"
	"io"
)

// Write ...
func (c *Client) Write(id string, input io.Reader) (err error) {
	url := fmt.Sprintf("%s/write/%s", c.url, id)
	_, err = c.request("POST", url, input)
	return
}
