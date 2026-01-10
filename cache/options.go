package cache

import "time"

var defaultConfig = options{
	cleanUpInterval: 10 * time.Second,
}

type options struct {
	cleanUpInterval time.Duration
}

func WithCleanUpInterval(interval time.Duration) Option {
	return func(cfg *options) {
		cfg.cleanUpInterval = interval
	}
}

type Option func(*options)
