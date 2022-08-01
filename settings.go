package main

import (
	"time"

	"go.riyazali.net/sqlite"
)

// Ticker for minimum delay between each all HTTP requests.
var DoTicker = time.NewTicker(time.Millisecond)

// timeout duration for all HTTP requests. Configurable with http_timeout_set
var DoTimeout = 5 * time.Second

/* http_rate_limit(delay_ms)
* Set the rate limit of all HTTP requests, in milliseconds.
 */
type HttpRateLimit struct{}

func (*HttpRateLimit) Deterministic() bool { return true }
func (*HttpRateLimit) Args() int           { return 1 }
func (*HttpRateLimit) Apply(c *sqlite.Context, values ...sqlite.Value) {
	ms := values[0].Int()
	DoTicker.Reset(time.Duration(ms) * time.Millisecond)
	c.ResultInt(1)
}

// http_timeout_set(duration)
type HttpTimeoutSet struct{}

func (*HttpTimeoutSet) Deterministic() bool { return true }
func (*HttpTimeoutSet) Args() int           { return 1 }

// TODO see if a duration string was passed, ex "10s". Else, parse in milliseconds
func (*HttpTimeoutSet) Apply(c *sqlite.Context, values ...sqlite.Value) {
	ms := values[0].Int()
	newDuration := time.Duration(ms) * time.Millisecond
	DoTimeout = newDuration
	c.ResultInt(ms)
}

func RegisterSettings(api *sqlite.ExtensionApi) error {
	if err := api.CreateFunction("http_rate_limit", &HttpRateLimit{}); err != nil {
		return err
	}
	if err := api.CreateFunction("http_timeout_set", &HttpTimeoutSet{}); err != nil {
		return err
	}
	return nil
}
