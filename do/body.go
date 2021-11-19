package http_do

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"go.riyazali.net/sqlite"
)

// http_do_body(method, url, headers, body, cookies)
type HttpDoBodyFunc struct{}

func (*HttpDoBodyFunc) Deterministic() bool { return true }
func (*HttpDoBodyFunc) Args() int           { return -1 }
func (*HttpDoBodyFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {

	if len(values) < 2 || len(values) > 5 {
		c.ResultError(errors.New("usage: http_do_body(method, url, headers, body, cookies)"))
		return
	}

	method := values[0].Text()
	url := values[1].Text()
	var headers string
	var cookies string
	var body []byte

	if len(values) >= 3 {
		headers = values[2].Text()
	}

	if len(values) >= 4 {
		body = values[3].Blob()
	}
	if len(values) >= 5 {
		cookies = values[4].Text()
	}

	client, request, err := prepareRequest(&PrepareRequestParams{method: method, url: url, headers: headers, body: body, cookies: cookies})

	if err != nil {
		fmt.Println(err)
		c.ResultError(err)
	}

	response, err := client.Do(request)

	if err != nil {
		c.ResultError(err)
	} else {
		body, _ := io.ReadAll(response.Body)
		c.ResultBlob(body)
	}
}

// http_post_body(url, headers, body, cookies)
type HttpPostBodyFunc struct{}

func (*HttpPostBodyFunc) Deterministic() bool { return true }
func (*HttpPostBodyFunc) Args() int           { return -1 }
func (*HttpPostBodyFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {

	if len(values) < 1 || len(values) > 4 {
		c.ResultError(errors.New("usage: http_post_body(url, headers, body, cookies)"))
		return
	}

	url := values[0].Text()
	var headers string
	var cookies string
	var body []byte

	if len(values) >= 2 {
		headers = values[1].Text()
	}

	if len(values) >= 3 {
		body = values[2].Blob()
	}
	if len(values) >= 4 {
		cookies = values[3].Text()
	}

	client, request, err := prepareRequest(&PrepareRequestParams{method: "POST", url: url, headers: headers, body: body, cookies: cookies})
	if err != nil {
		fmt.Println(err)
		c.ResultError(err)
	}

	response, err := client.Do(request)

	if err != nil {
		c.ResultError(err)
	} else {
		body, _ := io.ReadAll(response.Body)
		c.ResultBlob(body)
	}
}

// http_get_body(url, headers, cookies)
type HttpGetBodyFunc struct{}

func (*HttpGetBodyFunc) Deterministic() bool { return true }
func (*HttpGetBodyFunc) Args() int           { return -1 }
func (*HttpGetBodyFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {

	if len(values) < 1 || len(values) > 3 {
		c.ResultError(errors.New("usage: http_get_body(url, headers, cookies)"))
		return
	}

	url := values[0].Text()
	var headers string
	var cookies string

	if len(values) >= 2 {
		headers = values[1].Text()
	}

	if len(values) >= 3 {
		cookies = values[2].Text()
	}

	client, request, err := prepareRequest(&PrepareRequestParams{method: "GET", url: url, headers: headers, body: nil, cookies: cookies})
	if err != nil {
		fmt.Println(err)
		c.ResultError(err)
	}

	response, err := client.Do(request)

	if err != nil {
		c.ResultError(err)
	} else {
		body, _ := ioutil.ReadAll(response.Body)
		c.ResultBlob(body)
	}
}
