package je

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Client ...
type Client struct {
	url string
}

// ClientOptions ...
type ClientOptions struct {
}

// NewClient ...
func NewClient(url string, options *ClientOptions) *Client {
	url = strings.TrimSuffix(url, "/")

	return &Client{url: url}
}

func (c *Client) request(method, url string, body io.Reader) (res []*Job, err error) {
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
		if response.Header.Get("Content-Type") == "application/json" {
			err = json.NewDecoder(response.Body).Decode(&res)
			if err != nil {
				log.Errorf("error decoding response from %s: %s", url, err)
				return
			}
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
func (c *Client) GetJobByID(id string) (res []*Job, err error) {
	return c.Search(&SearchOptions{
		Filter: &SearchFilter{
			ID: id,
		},
	})
}

// UpdateJob updates a job to the remote store
func (c *Client) UpdateJob(job *Job) (res []*Job, err error) {
	return c.Update(job)
}

// Close ...
func (c *Client) Close(id string) (err error) {
	url := fmt.Sprintf("%s/close/%s", c.url, id)
	_, err = c.request("POST", url, nil)
	return
}

// Create ...
func (c *Client) Create(name string, args []string, input io.Reader, interactive, wait bool) (res []*Job, err error) {

	url := fmt.Sprintf("%s/create/%s?args=%s", c.url, name, JoinArgs(args))

	if interactive {
		url += "&interactive=1"
	}

	if wait {
		url += "&wait=1"
	}

	return c.request("POST", url, input)
}

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

// Logs ...
func (c *Client) Logs(id string, follow bool) (r io.ReadCloser, err error) {
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

// Output ...
func (c *Client) Output(id string, follow bool) (r io.ReadCloser, err error) {
	var url string

	if follow {
		url = fmt.Sprintf("%s/output/%s?follow=1", c.url, id)
	} else {
		url = fmt.Sprintf("%s/output/%s", c.url, id)
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

// SearchFilter ...
type SearchFilter struct {
	ID    string
	Name  string
	State string
}

// SearchOptions ...
type SearchOptions struct {
	Filter *SearchFilter
}

// Search ...
func (c *Client) Search(options *SearchOptions) (res []*Job, err error) {
	url := fmt.Sprintf("%s/search", c.url)

	filter := options.Filter

	switch {
	case filter.ID != "":
		url += fmt.Sprintf("/%s", filter.ID)
	case filter.Name != "":
		url += fmt.Sprintf("?q=name:%s", filter.Name)
	case filter.State != "":
		url += fmt.Sprintf("?q=state:%d", ParseState(filter.State))
	}

	return c.request("GET", url, nil)
}

// Update ...
func (c *Client) Update(job *Job) (res []*Job, err error) {
	url := fmt.Sprintf("%s/update/%d", c.url, job.ID)

	body, err := json.Marshal(job)
	if err != nil {
		log.Errorf("error encoding job: %s", err)
		return nil, err
	}

	return c.request("POST", url, bytes.NewBuffer(body))
}

// Read ...
func (c *Client) Read(id string) (r io.ReadCloser, err error) {
	url := fmt.Sprintf("%s/read/%s", c.url, id)

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

// Write ...
func (c *Client) Write(id string, input io.Reader) (err error) {
	url := fmt.Sprintf("%s/write/%s", c.url, id)
	_, err = c.request("POST", url, input)
	return
}
