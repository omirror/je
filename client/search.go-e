package client

import (
	"fmt"

	"git.mills.io/prologic/je"
)

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
func (c *Client) Search(options *SearchOptions) (res []*je.Job, err error) {
	url := fmt.Sprintf("%s/search", c.url)

	filter := options.Filter

	switch {
	case filter.ID != "":
		url += fmt.Sprintf("/%s", filter.ID)
	case filter.Name != "":
		url += fmt.Sprintf("?q=name:%s", filter.Name)
	case filter.State != "":
		url += fmt.Sprintf("?q=state:%d", je.ParseState(filter.State))
	}

	return c.request("GET", url, nil)
}
