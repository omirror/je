package pool

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Config contains the bitcask configuration parameters
type Config struct {
	Backlog int    `json:"backlog"`
	MaxJobs uint32 `json:"maxjobs"`
}

// Load loads a configuration from the given path
func Load(path string) (*Config, error) {
	var cfg Config

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save saves the configuration to the provided path
func (c *Config) Save(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	if _, err = f.Write(data); err != nil {
		return err
	}

	if err = f.Sync(); err != nil {
		return err
	}

	return f.Close()
}
