package pool

import "honnef.co/go/tools/config"

const (
	// DefaultBacklog is the default backlog size of the queue
	DefaultBacklog = 32

	// DefaultMaxJobs is the default maximum number of parallel jobs
	DefaultMaxJobs = 2
)

// Option is a function that takes a config struct and modifies it
type Option func(*config.Config) error

// WithBacklog sets the backlog queue size
func WithBacklog(backlog int) Option {
	return func(cfg *config.Config) error {
		cfg.MaxBacklog = backlog
		return nil
	}
}

// WithMaxJobs sets the maximum number of parallel jobs
func WithMaxJobs(maxjobs uint32) Option {
	return func(cfg *config.Config) error {
		cfg.MaxJobs = maxjobs
		return nil
	}
}

func newDefaultConfig() *config.Config {
	return &config.Config{
		MaxBacklog: DefaultBacklog,
		MaxJobs:    DefaultMaxJobs,
	}
}
