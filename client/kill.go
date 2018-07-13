package client

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Kill ...
func (c *Client) Kill(id string, force bool) (err error) {
	var url string

	if force {
		url = fmt.Sprintf("%s/kill/%s?force=1", c.url, id)
	} else {
		url = fmt.Sprintf("%s/kill/%s", c.url, id)
	}
	client := &http.Client{}

	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Errorf("error constructing request to %s: %s", url, err)
		return
	}

	response, err := client.Do(request)
	if err != nil {
		log.Errorf("error sending request to %s: %s", url, err)
		return
	}

	if response.StatusCode != 200 {
		err = fmt.Errorf("error killing job #%s: %s", id, response.Status)
		log.Error(err)
		return
	}

	return
}
