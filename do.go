package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"github.com/augmentable-dev/vtab"
	"go.riyazali.net/sqlite"
)

// timestamp layout to match SQLite's 'datetime()' format, ISO8601 subset
const sqliteDatetimeFormat = "2006-01-02 15:04:05.999"

// Format ghe given time as a SQLite date timestamp
func formatSqliteDatetime(t *time.Time) *string {
	s := t.UTC().Format(sqliteDatetimeFormat)
	return &s
}

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

/* http_get_headers(method, url, headers, body, cookies)
* Perform a HTTP request with the given method, URL, headers,
* body, and cookies.
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
		fmt.Println(err)
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
		fmt.Println(err)
		c.ResultError(err)
	}

	resultResponseHeaders(client, request, c)
}

var SharedDoTableColumns = []vtab.Column{

	{Name: "request_url", Type: sqlite.SQLITE_TEXT.String()},
	{Name: "request_method", Type: sqlite.SQLITE_TEXT.String()},
	{Name: "request_headers", Type: sqlite.SQLITE_TEXT.String()},
	{Name: "request_cookies", Type: sqlite.SQLITE_TEXT.String()},
	{Name: "request_body", Type: sqlite.SQLITE_TEXT.String()},

	{Name: "response_status", Type: sqlite.SQLITE_TEXT.String()},
	{Name: "response_status_code", Type: sqlite.SQLITE_INTEGER.String()},
	{Name: "response_headers", Type: sqlite.SQLITE_TEXT.String()},
	{Name: "response_cookies", Type: sqlite.SQLITE_TEXT.String()},
	{Name: "response_body", Type: sqlite.SQLITE_BLOB.String()},
	{Name: "remote_address", Type: sqlite.SQLITE_TEXT.String()},
	{Name: "timings", Type: sqlite.SQLITE_TEXT.String()},
	{Name: "meta", Type: sqlite.SQLITE_TEXT.String()},
}

var GetTableColumns = append([]vtab.Column{
	{Name: "url", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
	{Name: "headers", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
	{Name: "cookies", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
}, SharedDoTableColumns...)

var PostTableColumns = append([]vtab.Column{
	{Name: "url", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
	{Name: "headers", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
	{Name: "body", Type: sqlite.SQLITE_BLOB.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
	{Name: "cookies", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
}, SharedDoTableColumns...)

var DoTableColumns = append([]vtab.Column{
	{Name: "method", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
	{Name: "url", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
	{Name: "headers", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
	{Name: "body", Type: sqlite.SQLITE_BLOB.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
	{Name: "cookies", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
}, SharedDoTableColumns...)

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
			fmt.Println("invalid cookies")
			return nil, nil, sqlite.SQLITE_ERROR
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

type HttpDoCursor struct {
	current int

	// Original HTTP request made
	request *http.Request
	// HTTP response to the original request
	response *http.Response
	// Timestamps of various lower-level events for the timings column
	timing Timings
	// Remote network address, filled in when HTTP connection is made, IP address
	RemoteAddr string

	columns []vtab.Column
}

func (cur *HttpDoCursor) Column(ctx *sqlite.Context, c int) error {
	col := cur.columns[c]

	// TODO this should be a compile-time (?) option. response is nil bc commented out error handling in client.Do in GET
	if strings.HasPrefix(col.Name, "response_") && cur.response == nil {
		ctx.ResultNull()
		return nil
	}
	switch col.Name {
	case "url":
		ctx.ResultText("")
	case "headers":
		ctx.ResultText("")
	case "cookies":
		ctx.ResultText("")

	case "request_url":
		ctx.ResultText(cur.request.URL.String())
	case "request_method":
		ctx.ResultText(cur.request.Method)
	case "request_headers":
		buf := new(bytes.Buffer)
		cur.request.Header.Write(buf)
		ctx.ResultText(buf.String())
	case "request_cookies":
		cookies := make([]string, len(cur.request.Cookies()))
		for _, c := range cur.request.Cookies() {
			cookies = append(cookies, c.Raw)
		}
		buf, err := json.Marshal(cookies)
		if err != nil {
			ctx.ResultError(err)
		} else {
			ctx.ResultText(string(buf))
		}
	case "request_body":
		body, err := ioutil.ReadAll(cur.request.Body)
		if err != nil {
			ctx.ResultError(err)
		}

		ctx.ResultBlob(body)
	case "response_status":
		ctx.ResultText(cur.response.Status)
	case "response_status_code":
		ctx.ResultInt(cur.response.StatusCode)
	case "response_headers":
		buf := new(bytes.Buffer)
		cur.response.Header.Write(buf)
		ctx.ResultText(buf.String())
	case "response_cookies":
		cookies := make([]string, len(cur.response.Cookies()))
		for _, c := range cur.response.Cookies() {
			cookies = append(cookies, c.Raw)
		}
		buf, err := json.Marshal(cookies)
		if err != nil {
			ctx.ResultError(err)
		} else {
			ctx.ResultText(string(buf))
		}
	case "response_body":
		start := time.Now()
		cur.timing.BodyStart = &start

		body, err := ioutil.ReadAll(cur.response.Body)
		end := time.Now()
		cur.timing.BodyEnd = &end

		if err != nil {
			ctx.ResultError(err)
		} else {
			ctx.ResultBlob(body)
		}
	case "remote_address":
		ctx.ResultText(cur.RemoteAddr)
	case "timings":
		buf, err := json.Marshal(cur.timing)
		if err != nil {
			ctx.ResultError(err)
			return nil
		}
		ctx.ResultText(string(buf))
	case "meta":
		ctx.ResultNull()
	}
	return nil
}

// one row for now
func (cur *HttpDoCursor) Next() (vtab.Row, error) {
	cur.current += 1
	if cur.current >= 1 {
		return nil, io.EOF
	}
	return cur, nil
}

func GetTableIterator(constraints []*vtab.Constraint, order []*sqlite.OrderBy) (vtab.Iterator, error) {
	var headers string
	var cookies string
	url := ""

	for _, constraint := range constraints {
		if constraint.Op == sqlite.INDEX_CONSTRAINT_EQ {
			column := GetTableColumns[constraint.ColIndex]
			switch column.Name {
			case "url":
				url = constraint.Value.Text()
			case "headers":
				headers = constraint.Value.Text()
			case "cookies":
				cookies = constraint.Value.Text()
			}
		}
	}

	cursor := HttpDoCursor{
		columns: GetTableColumns,
	}
	client, request, err := prepareRequest(&PrepareRequestParams{method: "GET", url: url, headers: headers, body: nil, cookies: cookies})
	if err != nil {
		return nil, sqlite.SQLITE_ERROR
	}

	request = traceAndInclude(request, &cursor)

	started := time.Now()
	cursor.timing.Started = &started

	response, err := client.Do(request)
	if err != nil {
		return nil, sqlite.SQLITE_ERROR
	}

	cursor.current = -1
	cursor.request = request
	cursor.response = response

	return &cursor, nil
}

func PostTableIterator(constraints []*vtab.Constraint, order []*sqlite.OrderBy) (vtab.Iterator, error) {
	var headers string
	var cookies string
	var body []byte
	url := ""

	for _, constraint := range constraints {
		if constraint.Op == sqlite.INDEX_CONSTRAINT_EQ {
			column := PostTableColumns[constraint.ColIndex]
			switch column.Name {
			case "url":
				url = constraint.Value.Text()
			case "headers":
				headers = constraint.Value.Text()
			case "body":
				body = constraint.Value.Blob()
			case "cookies":
				cookies = constraint.Value.Text()
			}
		}
	}

	cursor := HttpDoCursor{
		columns: PostTableColumns,
	}
	client, request, err := prepareRequest(&PrepareRequestParams{method: "POST", url: url, headers: headers, body: body, cookies: cookies})
	if err != nil {
		fmt.Println("Error preparing request", err)
		return nil, sqlite.SQLITE_ERROR
	}

	request = traceAndInclude(request, &cursor)

	started := time.Now()
	cursor.timing.Started = &started

	response, err := client.Do(request)

	// TODO make this configurable. I don't want it to always error
	// if there's some connection error, but maybe other want that
	if err != nil {
		fmt.Println("error on client.Do", err)
		return nil, sqlite.SQLITE_ERROR
	}

	cursor.current = -1
	cursor.request = request
	cursor.response = response

	return &cursor, nil
}

func DoTableIterator(constraints []*vtab.Constraint, order []*sqlite.OrderBy) (vtab.Iterator, error) {
	var method string
	var headers string
	var cookies string
	var body []byte
	url := ""

	for _, constraint := range constraints {
		if constraint.Op == sqlite.INDEX_CONSTRAINT_EQ {
			column := DoTableColumns[constraint.ColIndex]
			switch column.Name {
			case "method":
				method = constraint.Value.Text()
			case "url":
				url = constraint.Value.Text()
			case "headers":
				headers = constraint.Value.Text()
			case "body":
				body = constraint.Value.Blob()
			case "cookies":
				cookies = constraint.Value.Text()
			}
		}
	}

	cursor := HttpDoCursor{
		columns: DoTableColumns,
	}
	client, request, err := prepareRequest(&PrepareRequestParams{method: method, url: url, headers: headers, body: body, cookies: cookies})
	if err != nil {
		fmt.Println(err)
		return nil, sqlite.SQLITE_ERROR
	}

	request = traceAndInclude(request, &cursor)

	t := time.Now()
	cursor.timing.Started = &t

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return nil, sqlite.SQLITE_ERROR
	}

	cursor.current = -1
	cursor.request = request
	cursor.response = response

	return &cursor, nil
}

// For the given HTTP request, write all timing info to the given
// cursor's "timing" object, so we can surface as a column later
func traceAndInclude(request *http.Request, cursor *HttpDoCursor) *http.Request {

	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			t := time.Now()
			cursor.timing.FirstResponseByte = &t
		},

		DNSStart: func(i httptrace.DNSStartInfo) {
			t := time.Now()
			cursor.timing.DNSStart = &t
		},

		DNSDone: func(i httptrace.DNSDoneInfo) {
			t := time.Now()
			cursor.timing.DNSDone = &t
		},

		ConnectStart: func(network string, addr string) {
			t := time.Now()
			cursor.timing.ConnectStart = &t
		},

		ConnectDone: func(network, addr string, err error) {
			t := time.Now()
			cursor.timing.ConnectDone = &t
		},

		GotConn: func(g httptrace.GotConnInfo) {
			t := time.Now()
			cursor.timing.GotConn = &t
			cursor.RemoteAddr = g.Conn.RemoteAddr().String()
		},
		TLSHandshakeStart: func() {
			t := time.Now()
			cursor.timing.TLSHandshakeStart = &t
		},

		TLSHandshakeDone: func(c tls.ConnectionState, e error) {
			t := time.Now()
			cursor.timing.TLSHandshakeDone = &t
		},

		WroteHeaders: func() {
			t := time.Now()
			cursor.timing.WroteHeaders = &t
		},
	}
	return request.WithContext(httptrace.WithClientTrace(request.Context(), trace))

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

var DoModules = map[string]sqlite.Module{
	"http_get":  vtab.NewTableFunc("http_get", GetTableColumns, GetTableIterator),
	"http_post": vtab.NewTableFunc("http_post", PostTableColumns, PostTableIterator),
	"http_do":   vtab.NewTableFunc("http_do", DoTableColumns, DoTableIterator),
}
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
	for name, module := range DoModules {
		if err := api.CreateModule(name, module); err != nil {
			return err
		}
	}
	for name, function := range DoFunctions {
		if err := api.CreateFunction(name, function); err != nil {
			return err
		}
	}
	return nil
}
