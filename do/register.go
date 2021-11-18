package http_do

import (
	"errors"
	"net/url"
	"time"

	"github.com/augmentable-dev/vtab"
	"go.riyazali.net/sqlite"
)

// http_post_form_url_encoded(name1, value1, ...)
type HttpPostFormUrlEncoded struct{}

func (*HttpPostFormUrlEncoded) Deterministic() bool { return true }
func (*HttpPostFormUrlEncoded) Args() int           { return -1 }
func (*HttpPostFormUrlEncoded) Apply(c *sqlite.Context, values ...sqlite.Value) {

	if len(values)%2 != 0 {
		c.ResultError(errors.New("http_post_form_url_encoded must have even-numbered arguments"))
		return
	}

	data := url.Values{}

	for i := 0; i < len(values); i = i + 2 {
		key := values[i].Text()
		value := values[i+1].Text()

		data.Set(key, value)
	}

	c.ResultText(data.Encode())
}

// TODO HttpPostMultipartForm

// http_rate_limit(delay, [n])
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
func (*HttpTimeoutSet) Apply(c *sqlite.Context, values ...sqlite.Value) {
	// TODO see if a duration string was passed, e.g. "10s". Else, parse in milliseconds
	ms := values[0].Int()
	newDuration := time.Duration(ms) * time.Millisecond
	DoTimeout = newDuration
	c.ResultInt(ms)
}

var modules = map[string]sqlite.Module{
	"http_get":  vtab.NewTableFunc("http_get", GetTableColumns, GetTableIterator),
	"http_post": vtab.NewTableFunc("http_post", PostTableColumns, PostTableIterator),
	"http_do":   vtab.NewTableFunc("http_do", DoTableColumns, DoTableIterator),
}
var functions = map[string]sqlite.Function{
	"http_get_body":              &HttpGetBodyFunc{},
	"http_post_body":             &HttpPostBodyFunc{},
	"http_do_body":               &HttpDoBodyFunc{},
	"http_get_headers":           &HttpGetHeadersFunc{},
	"http_post_headers":          &HttpPostHeadersFunc{},
	"http_do_headers":            &HttpDoHeadersFunc{},
	"http_post_form_urlencoded": &HttpPostFormUrlEncoded{},
	"http_rate_limit":            &HttpRateLimit{},
	"http_timeout_set":           &HttpTimeoutSet{},
}

func Register(api *sqlite.ExtensionApi) error {
	for name, module := range modules {
		if err := api.CreateModule(name, module); err != nil {
			return err
		}
	}
	for name, function := range functions {
		if err := api.CreateFunction(name, function); err != nil {
			return err
		}
	}
	return nil
}
