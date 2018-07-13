package client

import (
	"fmt"
)

// Close ...
func (c *Client) Close(id string) (err error) {
	url := fmt.Sprintf("%s/close/%s", c.url, id)
	_, err = c.request("POST", url, nil)
	return
}
