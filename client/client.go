package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

func (c *Client) request(method, url string, body io.Reader) (res []*je.Job, err error) {
	client := &http.Client{}

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Errorf("error constructing request to %s: %s", url, err)
		return
	}

	response, err := client.Do(request)
	if err != nil {
		log.Errorf("error sending request to %s: %s", url, err)
		return
	}

	if response.StatusCode == http.StatusNotFound {
		return
	} else if response.StatusCode == http.StatusOK {
		defer response.Body.Close()
		err = json.NewDecoder(response.Body).Decode(&res)
		if err != nil {
			log.Errorf("error decoding response from %s: %s", url, err)
			return
		}
	} else {
		err = fmt.Errorf("unexpected response %s from %s %s", response.Status, method, url)
		log.Error(err)
		return
	}

	// Impossible
	return
}

// GetJobByID returns the matching job by id
func (c *Client) GetJobByID(id string) (res []*je.Job, err error) {
	i, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		log.Errorf("invalud id %s: %s", id, err)
		return
	}

	return c.Search(&SearchOptions{
		Filter: &SearchFilter{
			ID: i,
		},
	})
}
