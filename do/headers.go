package http_do

import (
	"bytes"
	"errors"
	"fmt"

	"go.riyazali.net/sqlite"
)

// http_get_headers(url, headers, cookies)
type HttpGetHeadersFunc struct{}

func (*HttpGetHeadersFunc) Deterministic() bool { return true }
func (*HttpGetHeadersFunc) Args() int           { return -1 }
func (*HttpGetHeadersFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {

	if len(values) < 1 || len(values) > 3 {
		c.ResultError(errors.New("usage: http_get_headers(url, headers, cookies)"))
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
		buf := new(bytes.Buffer)
		response.Header.Write(buf)
		c.ResultText(buf.String())
	}
}

// http_post_headers(url, headers, body, cookies)
type HttpPostHeadersFunc struct{}

func (*HttpPostHeadersFunc) Deterministic() bool { return true }
func (*HttpPostHeadersFunc) Args() int           { return -1 }
func (*HttpPostHeadersFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {

	if len(values) < 1 || len(values) > 4 {
		c.ResultError(errors.New("usage: http_post_headers(url, headers, body, cookies)"))
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
	if len(values) >= 3 {
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
		buf := new(bytes.Buffer)
		response.Header.Write(buf)
		c.ResultText(buf.String())
	}
}

// http_do_headers(method, url, headers, body, cookies)
type HttpDoHeadersFunc struct{}

func (*HttpDoHeadersFunc) Deterministic() bool { return true }
func (*HttpDoHeadersFunc) Args() int           { return -1 }
func (*HttpDoHeadersFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {

	if len(values) < 2 || len(values) > 5 {
		c.ResultError(errors.New("usage: http_do_headers(method, url, headers, body, cookies)"))
		return
	}

	url := values[0].Text()
	method := values[1].Text()
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
		buf := new(bytes.Buffer)
		response.Header.Write(buf)
		c.ResultText(buf.String())
	}
}
