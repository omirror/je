package client

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"git.mills.io/prologic/je"
)

// SearchFilter ...
type SearchFilter struct {
	ID    uint64
	Name  string
	State int
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
	case filter.ID != 0:
		url += fmt.Sprintf("/%d", filter.ID)
	case filter.Name != "":
		url += fmt.Sprintf("?name=%s", filter.Name)
	case filter.State != 0:
		url += fmt.Sprintf("?state=%d", filter.State)
	default:
		err = fmt.Errorf("unsupported search filter: %+v", filter)
		log.Error(err)
		return
	}

	return c.request("GET", url, nil)
}
