package main

import (
	"fmt"
	"net/textproto"
	"strings"

	"go.riyazali.net/sqlite"
)

var SharedDoTableColumns = []string{

	"request_url text",
	"request_method text",
	"request_headers text",
	"request_cookies json",
	"request_body blob",

	"response_status text",
	"response_status_code integer",
	"response_headers text",
	"response_cookies json",
	"response_body blob",

	"remote_address text",
	"timings json",
	"meta json",
}

var GetTableColumns = append([]string{
	"url text hidden",
	"headers text hidden",
	"cookies text hidden",
}, SharedDoTableColumns...)

var PostTableColumns = append([]string{
	"url text hidden",
	"headers text hidden",
	"body blob hidden",
	"cookies text hidden",
}, SharedDoTableColumns...)

var DoTableColumns = append([]string{
	"method text hidden",
	"url text hidden",
	"headers text hidden",
	"body blob hidden",
	"cookies text hidden",
}, SharedDoTableColumns...)

type TableType int64
const (
	GET TableType = iota
	POST
	DO
)

type HttpDoModule struct {
	Type TableType
	Columns []string
}

func (m *HttpDoModule) Connect(_ *sqlite.Conn, _ []string, declare func(string) error) (sqlite.VirtualTable, error) {
	filesCreate := strings.Builder{}
	filesCreate.WriteString("CREATE TABLE x(")
	for i, column := range m.Columns {
		if i != len(headersEachColumns)-1 {
			filesCreate.WriteString(fmt.Sprintf("%s,", column))
		} else {
			filesCreate.WriteString(column)
		}
	}
	filesCreate.WriteString(")")
	return &HttpDoTable{}, declare(filesCreate.String())
}

type HttpDoTable struct {}

func (e *HttpDoTable) BestIndex(input *sqlite.IndexInfoInput) (*sqlite.IndexInfoOutput, error) {
	var output = &sqlite.IndexInfoOutput{
		ConstraintUsage: make([]*sqlite.ConstraintUsage, len(input.Constraints)),
	}
	var hasMethod bool
	for i, constraint := range input.Constraints {
		if constraint.Op == sqlite.INDEX_CONSTRAINT_EQ {
			column := headersEachColumns[constraint.ColumnIndex]
			switch column {
			case "headers":
				if !constraint.Usable {
					return nil, sqlite.SQLITE_CONSTRAINT
				}
				output.ConstraintUsage[i] = &sqlite.ConstraintUsage{ArgvIndex: 1, Omit: true}
				hasHeaders = true
			}
		}
	}
	if !hasHeaders {
		return nil, sqlite.SQLITE_ERROR
	}
	
	output.EstimatedCost = 1000
	output.EstimatedRows = 1000
	output.IndexNumber = 1
	return output, nil
}

func (e *HttpDoTable) Open() (sqlite.VirtualCursor, error) {
	return &HttpDoCursor{}, nil
}
func (e *HttpDoTable) Disconnect() error { return e.Destroy() }
func (e *HttpDoTable) Destroy() error    { return nil }

type HttpDoCursor struct {
	rowid        int64
	header        textproto.MIMEHeader
	keyOrder      []string
	currentKeyI   int
	currentValueI int
	done bool
}


func (cur *HttpDoCursor) Filter(idxNum int, idxStr string, values ...sqlite.Value) error {
	
	cur.rowid = 1
	return nil
}

func (cur *HttpDoCursor) Next() error {
	cur.rowid++

	cur.currentValueI += 1

	if cur.currentKeyI >= len(cur.keyOrder) {
		cur.done = true
		return nil
	}
	if cur.currentValueI >= len(cur.header.Values(cur.keyOrder[cur.currentKeyI])) {
		cur.currentKeyI += 1
		cur.currentValueI = 0
	}
	if cur.currentKeyI >= len(cur.keyOrder) {
		cur.done = true
		return nil
	}
	return nil
}
func (cur *HttpDoCursor) Eof() bool {
	return cur.done
}


func (cur *HttpDoCursor) Column(context *sqlite.VirtualTableContext, i int) error {
	col := headersEachColumns[i]

	switch col {
	case "name":
		context.ResultText(cur.keyOrder[cur.currentKeyI])
	case "value":
		context.ResultText(cur.header.Values(cur.keyOrder[cur.currentKeyI])[cur.currentValueI])
	}
	return nil
}

func (cur *HttpDoCursor) Rowid() (int64, error) { return cur.rowid, nil }
func (cur *HttpDoCursor) Close() error          { return nil }


/*
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
		return nil, fmt.Errorf("error preparing request: %s", err)
	}

	request = traceAndInclude(request, &cursor)

	started := time.Now()
	cursor.timing.Started = &started

	response, err := client.Do(request)

	// TODO make this configurable. I don't want it to always error
	// if there's some connection error, but maybe other want that
	if err != nil {
		return nil, fmt.Errorf("error on client.Do: %s", err)
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
		return nil, fmt.Errorf("error preparing request: %s", err)
	}

	request = traceAndInclude(request, &cursor)

	t := time.Now()
	cursor.timing.Started = &t

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error on client.Do: %s", err)
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

var DoModules = map[string]sqlite.Module{
	"http_get":  vtab.NewTableFunc("http_get", GetTableColumns, GetTableIterator),
	"http_post": vtab.NewTableFunc("http_post", PostTableColumns, PostTableIterator),
	"http_do":   vtab.NewTableFunc("http_do", DoTableColumns, DoTableIterator),
}

func RegisterDoTables(api *sqlite.ExtensionApi) error {
	for name, module := range DoModules {
		if err := api.CreateModule(name, module); err != nil {
			return err
		}
	}
	return nil
}
*/