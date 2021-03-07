package retry

import (
	"errors"
	"time"

	"github.com/valyala/fasthttp"
)

// Config is a HTTP configuration.
type Config struct {
	Attempts      int
	Timeouts      time.Duration
	Interval      time.Duration
	RetryAttempts int
}

// New to create a new instance for http retry.
func New(cfg *Config) *Config {
	attempts := 1
	if cfg.Attempts > 0 {
		attempts = cfg.Attempts
	}

	timeouts := time.Second * 1
	if cfg.Timeouts > 0 {
		timeouts = cfg.Timeouts
	}

	interval := time.Millisecond * 500
	if cfg.Interval > 0 {
		interval = cfg.Interval
	}

	return &Config{
		Attempts: attempts,
		Timeouts: timeouts,
		Interval: interval,
	}
}

// Do to running HTTP retry.
func (r *Config) Do(req *fasthttp.Request, resp *fasthttp.Response, c *fasthttp.Client) (*fasthttp.Response, error) {
	r.RetryAttempts++

	if e := c.DoTimeout(req, resp, r.Timeouts); e != nil {
		if r.RetryAttempts <= r.Attempts {
			time.Sleep(r.Interval)
			r.Do(req, resp, c)
		}

		if r.RetryAttempts > r.Attempts {
			return resp, errors.New("Request Timeout")
		}
	}

	return resp, nil
}
