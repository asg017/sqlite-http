package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"go.riyazali.net/sqlite"
)

// Give the result of the given HTTP request as a SQLite response, the body
func resultResponseBody(client *http.Client, request *http.Request, ctx *sqlite.Context) {
	response, err := client.Do(request)

	if err != nil {
		ctx.ResultError(err)
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ctx.ResultError(err)
		return
	}
	ctx.ResultBlob(body)
}

// Give the result of the given HTTP request as a SQLite response, the headers
func resultResponseHeaders(client *http.Client, request *http.Request, ctx *sqlite.Context) {
	response, err := client.Do(request)

	if err != nil {
		ctx.ResultError(err)
		return
	}
	buf := new(bytes.Buffer)
	response.Header.Write(buf)
	ctx.ResultText(buf.String())
}

/* http_do_body(method, url, headers, body, cookies)
* Perform a HTTP request with the given method, URL, headers,
* body, and cookies. Returns the HTTP body as a BLOB, errors if fails.
 */
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
		c.ResultError(err)
		return
	}

	resultResponseBody(client, request, c)

}

/* http_post_body(url, headers, body, cookies)
* Perform a POST request with the given URL, headers,
* body, and cookies. Returns the HTTP body as a BLOB, errors if fails.
 */
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
		c.ResultError(err)
		return
	}

	resultResponseBody(client, request, c)
}

/* http_get_body(url, headers, cookies)
* Perform a HTTP request with the given URL, headers, and cookies.
* Returns the HTTP body as a BLOB, errors if fails.
 */
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
		c.ResultError(err)
		return
	}

	resultResponseBody(client, request, c)
}

/* http_get_headers(url, headers, body, cookies)
* Perform a GET request on the given URL, headers, body, and cookies.
* Returns the HTTP response headers in wire format, errors if fails.
 */
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
		c.ResultError(err)
		return
	}

	resultResponseHeaders(client, request, c)
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
		c.ResultError(err)
	}

	resultResponseHeaders(client, request, c)
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
		c.ResultError(err)
	}

	resultResponseHeaders(client, request, c)
}

type Timings struct {
	Started           *time.Time
	FirstResponseByte *time.Time
	DNSStart          *time.Time
	DNSDone           *time.Time
	GotConn           *time.Time
	ConnectStart      *time.Time
	ConnectDone       *time.Time
	TLSHandshakeStart *time.Time
	TLSHandshakeDone  *time.Time
	WroteHeaders      *time.Time
	BodyStart         *time.Time
	BodyEnd           *time.Time
}

type TimingJSON struct {
	Started           *string `json:"start"`
	FirstResponseByte *string `json:"first_byte"`
	GotConn           *string `json:"connection"`
	WroteHeaders      *string `json:"wrote_headers"`
	DNSStart          *string `json:"dns_start"`
	DNSDone           *string `json:"dns_end"`
	ConnectStart      *string `json:"connect_start"`
	ConnectDone       *string `json:"connect_end"`
	TLSHandshakeStart *string `json:"tls_handshake_start"`
	TLSHandshakeDone  *string `json:"tls_handshake_end"`
	BodyStart         *string `json:"body_start"`
	BodyEnd           *string `json:"body_end"`
}

func (t Timings) MarshalJSON() ([]byte, error) {
	tj := TimingJSON{}
	if t.Started != nil {
		tj.Started = formatSqliteDatetime(t.Started)
	}
	if t.FirstResponseByte != nil {
		tj.FirstResponseByte = formatSqliteDatetime(t.FirstResponseByte)
	}

	if t.DNSStart != nil {
		tj.DNSStart = formatSqliteDatetime(t.DNSStart)
	}
	if t.DNSDone != nil {
		tj.DNSDone = formatSqliteDatetime(t.DNSDone)
	}
	if t.GotConn != nil {
		tj.GotConn = formatSqliteDatetime(t.GotConn)
	}
	if t.ConnectStart != nil {
		tj.ConnectStart = formatSqliteDatetime(t.ConnectStart)
	}
	if t.ConnectDone != nil {
		tj.ConnectDone = formatSqliteDatetime(t.ConnectDone)
	}
	if t.TLSHandshakeStart != nil {
		tj.TLSHandshakeStart = formatSqliteDatetime(t.TLSHandshakeStart)
	}
	if t.TLSHandshakeDone != nil {
		tj.TLSHandshakeDone = formatSqliteDatetime(t.TLSHandshakeDone)
	}
	if t.WroteHeaders != nil {
		tj.WroteHeaders = formatSqliteDatetime(t.WroteHeaders)
	}
	if t.BodyStart != nil {
		tj.BodyStart = formatSqliteDatetime(t.BodyStart)
	}
	if t.BodyEnd != nil {
		tj.BodyEnd = formatSqliteDatetime(t.BodyEnd)
	}

	return json.Marshal(tj)
}

type PrepareRequestParams struct {
	method  string
	url     string
	headers string
	body    []byte
	cookies string
}

// helper functions around http.NewRequest, takes  headers/body in sqlite-http formats
func prepareRequest(params *PrepareRequestParams) (*http.Client, *http.Request, error) {
	bodyReader := bytes.NewReader(params.body)

	request, err := http.NewRequest(params.method, params.url, bodyReader)
	if err != nil {
		return nil, nil, err
	}

	if params.headers != "" {
		h := readHeader(params.headers)

		// step 1: clear default headers
		for key := range request.Header {
			request.Header.Del(key)
		}

		for key, values := range h {
			for i := range values {
				request.Header.Add(key, values[i])
			}

		}

	}
	if params.cookies != "" {
		var parsed map[string]string
		err := json.Unmarshal([]byte(params.cookies), &parsed)
		if err != nil {
			return nil, nil, errors.New("invalid cookes")
		}

		for name, value := range parsed {
			request.AddCookie(&http.Cookie{
				Name:  name,
				Value: value,
			})
		}
	}
	client := &http.Client{
		Timeout: DoTimeout,
	}

	// block to rate limit properly
	<-DoTicker.C
	return client, request, nil
}



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


var DoFunctions = map[string]sqlite.Function{
	"http_get_body":             &HttpGetBodyFunc{},
	"http_post_body":            &HttpPostBodyFunc{},
	"http_do_body":              &HttpDoBodyFunc{},
	"http_get_headers":          &HttpGetHeadersFunc{},
	"http_post_headers":         &HttpPostHeadersFunc{},
	"http_do_headers":           &HttpDoHeadersFunc{},
	"http_post_form_urlencoded": &HttpPostFormUrlEncoded{},
	"http_rate_limit":           &HttpRateLimit{},
	"http_timeout_set":          &HttpTimeoutSet{},
}

func RegisterDo(api *sqlite.ExtensionApi) error {
	for name, function := range DoFunctions {
		if err := api.CreateFunction(name, function); err != nil {
			return err
		}
	}
	return nil
}
