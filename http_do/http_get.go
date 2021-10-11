package http_do

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/textproto"
	"strings"
	"time"

	"github.com/augmentable-dev/vtab"
	"go.riyazali.net/sqlite"
)

const sqliteDatetimeFormat = "2006-01-02 15:04:05.999"

func formatSqliteDatetime(t *time.Time) *string {
	s := t.UTC().Format(sqliteDatetimeFormat)
	return &s
}

var SharedDoTableColumns = []vtab.Column{
	{Name: "timings", Type: sqlite.SQLITE_TEXT.String()},
	{Name: "remote_address", Type: sqlite.SQLITE_TEXT.String()},

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

type DoCursorMeta struct {
	RemoteAddr string
}

type HttpDoCursor struct {
	current int

	request  *http.Request
	response *http.Response
	timing Timings
	meta DoCursorMeta

	columns []vtab.Column
}

type Timings struct {
	Started *time.Time
	FirstResponseByte *time.Time  
	DNSStart *time.Time 
	DNSDone *time.Time 
	GotConn *time.Time
	ConnectStart *time.Time 
	ConnectDone *time.Time 
	TLSHandshakeStart *time.Time 
	TLSHandshakeDone *time.Time 
	WroteHeaders *time.Time 
}
func (cur *HttpDoCursor) Column(ctx *sqlite.Context, c int) error {
	col := cur.columns[c]
	switch col.Name {
	case "url":
		ctx.ResultText("")
	case "headers":
		ctx.ResultText("")
	case "cookies":
		ctx.ResultText("")

	case "timings":
		buf, _ := json.Marshal(cur.timing)
		ctx.ResultText(string(buf))
	case "remote_address":
		ctx.ResultText(cur.meta.RemoteAddr)
	case "request_url":
		ctx.ResultText(cur.request.URL.String())
	case "request_method":
		ctx.ResultText(cur.request.Method)
	case "request_headers":
		buf := new(bytes.Buffer)
		cur.request.Header.Write(buf)
		ctx.ResultText(buf.String())
	case "request_cookies":
		cookies  := make([]string, len(cur.request.Cookies()))
		for _, c := range cur.request.Cookies() {
			cookies  = append(cookies, c.Raw)
		}
		buf, err := json.Marshal(cookies)
		if err != nil {
			ctx.ResultError(err)
		}else {
			ctx.ResultText(string(buf))
		}
	case "request_body":
		body, _ := io.ReadAll(cur.request.Body)
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
		cookies  := make([]string, len(cur.response.Cookies()))
		for _, c := range cur.response.Cookies() {
			cookies  = append(cookies, c.Raw)
		}
		buf, err := json.Marshal(cookies)
		if err != nil {
			ctx.ResultError(err)
		}else {
			ctx.ResultText(string(buf))
		}
	case "response_body":
		body, _ := io.ReadAll(cur.response.Body)
		ctx.ResultBlob(body)
	default:
		fmt.Println("what the fuck")
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

func createRequest(method string, url string, headers string, body []byte, cookies string) (*http.Request, error) {
	bodyReader := bytes.NewReader(body)
	
	request, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		fmt.Println("Error making request", url)
		fmt.Println(err)
		return nil, sqlite.SQLITE_ABORT
	}

	if headers != "" {
		h, err := readHeader(headers)

		if err != nil {
			fmt.Println("invalid headers")
			return nil, sqlite.SQLITE_ABORT
		}
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
	if cookies != "" {
		var parsed map[string]string
		err := json.Unmarshal([]byte(cookies), &parsed)
		if err != nil {
			fmt.Println("invalid cookies")
			return nil, sqlite.SQLITE_ABORT
		}

		for name, value := range parsed {
			request.AddCookie(&http.Cookie{
				Name:    name,
				Value:   value,
			})
		}
	}
	return request, nil
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
	client := &http.Client{}

	request, err := createRequest("GET", url, headers, nil, cookies)
	if err != nil {
		fmt.Println(err)
		return nil, sqlite.SQLITE_ABORT
	}

	request = traceAndInclude(request, &cursor)

	
	started := time.Now()
	cursor.timing.Started = &started

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return nil, sqlite.SQLITE_ABORT
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
	client := &http.Client{}

	request, err := createRequest("POST", url, headers, body, cookies)
	if err != nil {
		fmt.Println(err)
		return nil, sqlite.SQLITE_ABORT
	}

	request = traceAndInclude(request, &cursor)

	
	started := time.Now()
	cursor.timing.Started = &started

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return nil, sqlite.SQLITE_ABORT
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
	client := &http.Client{}

	request, err := createRequest(method, url, headers, body, cookies)
	if err != nil {
		fmt.Println(err)
		return nil, sqlite.SQLITE_ABORT
	}

	request = traceAndInclude(request, &cursor)

	t := time.Now()
	cursor.timing.Started = &t

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return nil, sqlite.SQLITE_ABORT
	}

	cursor.current = -1
	cursor.request = request
	cursor.response = response

	return &cursor, nil
}

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
			cursor.timing.ConnectStart =  &t
		},

		ConnectDone: func(network, addr string, err error) {
			t := time.Now()
			cursor.timing.ConnectDone =  &t
		},

		GotConn: func(g httptrace.GotConnInfo) {
			t := time.Now()
			cursor.timing.GotConn =  &t
			cursor.meta.RemoteAddr = g.Conn.RemoteAddr().String()
		},
		TLSHandshakeStart: func() {
			t := time.Now()
			cursor.timing.TLSHandshakeStart =  &t
		},

		TLSHandshakeDone: func(c tls.ConnectionState, e error) {
			t := time.Now()
			cursor.timing.TLSHandshakeDone =  &t
		},

		WroteHeaders: func() {
			t := time.Now()
			cursor.timing.WroteHeaders =  &t
		},
	}
	return request.WithContext(httptrace.WithClientTrace(request.Context(), trace))

}

func readHeader(rawHeader string) (textproto.MIMEHeader, error) {
	headerReader := textproto.NewReader(bufio.NewReader(strings.NewReader(rawHeader)))
	header, _ := headerReader.ReadMIMEHeader()
	return header, nil
}
type TimingJSON struct {
	Started * string `json:"started"`
	FirstResponseByte * string `json:"first_byte"`
	DNSStart * string `json:"dns_start"`
	DNSDone * string `json:"dns_done"`
	GotConn * string `json:"got_conn"`
	ConnectStart * string `json:"connect_start"`
	ConnectDone * string `json:"connect_done"`
	TLSHandshakeStart * string `json:"tls_handshake_start"`
	TLSHandshakeDone * string `json:"tls_handshake_done"`
	WroteHeaders * string `json:"wrote_headers"`
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
	
	return json.Marshal(tj)
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

	client := &http.Client{}
	url := values[0].Text()
	var headers string;
	var cookies string;

	if len(values) >= 2 {
		headers = values[1].Text()
	}

	if len(values) >= 3 {
		cookies = values[2].Text()
	}
	
	request, err := createRequest("GET", url, headers, nil, cookies)
	if err != nil {
		fmt.Println(err)
		c.ResultError(err)
	}

	response, err := client.Do(request)
	
	if err != nil {
		c.ResultError(err)
	}else {
		body, _ := io.ReadAll(response.Body)
		c.ResultBlob(body)
	}
}

// http_get_headers(url, headers, cookies)
type HttpGetHeadersFunc struct{}
func (*HttpGetHeadersFunc) Deterministic() bool { return true }
func (*HttpGetHeadersFunc) Args() int           { return -1 }
func (*HttpGetHeadersFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {

	if len(values) < 1 || len(values) > 3 {
		c.ResultError(errors.New("usage: http_get_headers(url, headers, cookies)"))
		return
	}

	client := &http.Client{}
	url := values[0].Text()
	var headers string;
	var cookies string;

	if len(values) >= 2 {
		headers = values[1].Text()
	}

	if len(values) >= 3 {
		cookies = values[2].Text()
	}
	
	request, err := createRequest("GET", url, headers, nil, cookies)
	if err != nil {
		fmt.Println(err)
		c.ResultError(err)
	}

	response, err := client.Do(request)
	
	if err != nil {
		c.ResultError(err)
	}else {
		buf := new(bytes.Buffer)
		response.Header.Write(buf)
		c.ResultText(buf.String())
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

	client := &http.Client{}
	url := values[0].Text()
	var headers string;
	var cookies string;
	var body []byte;

	if len(values) >= 2 {
		headers = values[1].Text()
	}

	if len(values) >= 3 {
		body = values[2].Blob()
	}
	if len(values) >= 3 {
		cookies = values[3].Text()
	}
	
	request, err := createRequest("POST", url, headers, body, cookies)
	if err != nil {
		fmt.Println(err)
		c.ResultError(err)
	}

	response, err := client.Do(request)
	
	if err != nil {
		c.ResultError(err)
	}else {
		body, _ := io.ReadAll(response.Body)
		c.ResultBlob(body)
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

	client := &http.Client{}
	url := values[0].Text()
	var headers string;
	var cookies string;
	var body []byte;

	if len(values) >= 2 {
		headers = values[1].Text()
	}

	if len(values) >= 3 {
		body = values[2].Blob()
	}
	if len(values) >= 3 {
		cookies = values[3].Text()
	}
	
	request, err := createRequest("POST", url, headers, body, cookies)
	if err != nil {
		fmt.Println(err)
		c.ResultError(err)
	}

	response, err := client.Do(request)
	
	if err != nil {
		c.ResultError(err)
	}else {
		buf := new(bytes.Buffer)
		response.Header.Write(buf)
		c.ResultText(buf.String())
	}
}

// http_do_body(method, url, headers, body, cookies)
type HttpDoBodyFunc struct{}
func (*HttpDoBodyFunc) Deterministic() bool { return true }
func (*HttpDoBodyFunc) Args() int           { return -1 }
func (*HttpDoBodyFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {

	if len(values) < 2 || len(values) > 5 {
		c.ResultError(errors.New("usage: http_do_body(method, url, headers, body, cookies)"))
		return
	}

	client := &http.Client{}
	url := values[0].Text()
	method := values[1].Text()
	var headers string;
	var cookies string;
	var body []byte;

	if len(values) >= 3 {
		headers = values[2].Text()
	}

	if len(values) >= 4 {
		body = values[3].Blob()
	}
	if len(values) >= 5 {
		cookies = values[4].Text()
	}
	
	request, err := createRequest(method, url, headers, body, cookies)

	if err != nil {
		fmt.Println(err)
		c.ResultError(err)
	}

	response, err := client.Do(request)
	
	if err != nil {
		c.ResultError(err)
	}else {
		body, _ := io.ReadAll(response.Body)
		c.ResultBlob(body)
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

	client := &http.Client{}
	url := values[0].Text()
	method := values[1].Text()
	var headers string;
	var cookies string;
	var body []byte;

	if len(values) >= 3 {
		headers = values[2].Text()
	}

	if len(values) >= 4 {
		body = values[3].Blob()
	}
	if len(values) >= 5 {
		cookies = values[4].Text()
	}
	
	request, err := createRequest(method, url, headers, body, cookies)

	if err != nil {
		fmt.Println(err)
		c.ResultError(err)
	}

	response, err := client.Do(request)
	
	if err != nil {
		c.ResultError(err)
	}else {
		buf := new(bytes.Buffer)
		response.Header.Write(buf)
		c.ResultText(buf.String())
	}
}


var modules = map[string]sqlite.Module{
	"http_get": vtab.NewTableFunc("http_get", GetTableColumns, GetTableIterator),
	"http_post": vtab.NewTableFunc("http_post", PostTableColumns, PostTableIterator),
	"http_do": vtab.NewTableFunc("http_do", DoTableColumns, DoTableIterator),
}
var functions = map[string]sqlite.Function{
	"http_get_body": &HttpGetBodyFunc{},
	"http_get_headers": &HttpGetHeadersFunc{},
	"http_post_body": &HttpPostBodyFunc{},
	"http_post_headers": &HttpPostHeadersFunc{},
	"http_do_body": &HttpDoBodyFunc{},
	"http_do_headers": &HttpDoHeadersFunc{},
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
