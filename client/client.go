package client

import (
	"encoding/json"
	"fmt"
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
