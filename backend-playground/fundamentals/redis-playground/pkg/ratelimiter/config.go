package ratelimiter

import "time"

type Config struct {
	Prefix      string
	Namespace   string
	MaxRequests int64
	Window      time.Duration
}

func (c *Config) setDefaults() {
	if c.Prefix == "" {
		c.Prefix = "rate_limit"
	}

	if c.Namespace == "" {
		c.Namespace = "default"
	}

	if c.MaxRequests <= 0 {
		c.MaxRequests = 100
	}

	if c.Window <= 0 {
		c.Window = time.Second * 10
	}
}
