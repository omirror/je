package client

import (
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Logs ...
func (c *Client) Logs(id string, follow bool) (r io.Reader, err error) {
	var url string

	if follow {
		url = fmt.Sprintf("%s/logs/%s?follow=1", c.url, id)
	} else {
		url = fmt.Sprintf("%s/logs/%s", c.url, id)
	}

	client := &http.Client{}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("error constructing request to %s: %s", url, err)
		return
	}

	response, err := client.Do(request)
	if err != nil {
		log.Errorf("error sending request to %s: %s", url, err)
		return
	}

	return response.Body, nil
}
