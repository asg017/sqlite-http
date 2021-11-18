package http_do

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
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

func readHeader(rawHeader string) (textproto.MIMEHeader, error) {
	headerReader := textproto.NewReader(bufio.NewReader(strings.NewReader(rawHeader)))
	header, _ := headerReader.ReadMIMEHeader()
	return header, nil
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

var DoTicker = time.NewTicker(time.Millisecond)
var DoTimeout = 5 * time.Second

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
		fmt.Println("Error making request", params.url)
		fmt.Println(err)
		return nil, nil, sqlite.SQLITE_ERROR
	}

	if params.headers != "" {
		h, err := readHeader(params.headers)

		if err != nil {
			fmt.Println("invalid headers")
			return nil, nil, sqlite.SQLITE_ERROR
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

	<-DoTicker.C
	return client, request, nil
}

type DoCursorMeta struct {
	RemoteAddr string
}

type HttpDoCursor struct {
	current int

	request  *http.Request
	response *http.Response
	timing   Timings
	meta     DoCursorMeta

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

		body, _ := io.ReadAll(cur.response.Body)

		end := time.Now()
		cur.timing.BodyEnd = &end

		ctx.ResultBlob(body)
	case "remote_address":
		ctx.ResultText(cur.meta.RemoteAddr)
	case "timings":
		buf, _ := json.Marshal(cur.timing)
		ctx.ResultText(string(buf))
	case "meta":
		ctx.ResultNull()
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
		fmt.Println(err)
		return nil, sqlite.SQLITE_ERROR
	}

	request = traceAndInclude(request, &cursor)

	started := time.Now()
	cursor.timing.Started = &started

	response, err := client.Do(request)
	if err != nil {
		fmt.Println("client.Do error")
		fmt.Println(err)
		//return nil, sqlite.SQLITE_ERROR
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
		fmt.Println(err)
		return nil, sqlite.SQLITE_ERROR
	}

	request = traceAndInclude(request, &cursor)

	started := time.Now()
	cursor.timing.Started = &started

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
			cursor.meta.RemoteAddr = g.Conn.RemoteAddr().String()
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
