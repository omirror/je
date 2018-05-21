package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.mills.io/prologic/je"
)

// Client ...
type Client struct {
	url string
}

// Options ...
type Options struct {
}

// NewClient ...
func NewClient(url string, options *Options) *Client {
	url = strings.TrimSuffix(url, "/")

	return &Client{url: url}
}

// Info ...
func (c *Client) Info(id string) (res []*je.Job, err error) {
	url := fmt.Sprintf("%s/search/%s", c.url, id)
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

	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(&res)
	if err != nil {
		log.Errorf(
			"error decoding response from %s for job #%s: %s",
			url, id, err,
		)
		return
	}

	return
}

// Logs ...
func (c *Client) Logs(id string) (r io.Reader, err error) {
	url := fmt.Sprintf("%s/logs/%s", c.url, id)
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

// Run ...
func (c *Client) Run(name string) (res *je.Job, err error) {
	url := fmt.Sprintf("%s/job/%s", c.url, name)
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

	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(&res)
	if err != nil {
		log.Errorf(
			"error decoding response from %s for %s: %s",
			url, name, err,
		)
		return
	}

	return
}
