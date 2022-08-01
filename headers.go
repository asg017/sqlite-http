package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/augmentable-dev/vtab"
	"go.riyazali.net/sqlite"
)

// utility for reading headers in "wire" format into a query-able textproto.MIMEHeader
func readHeader(rawHeader string) textproto.MIMEHeader {
	headerReader := textproto.NewReader(bufio.NewReader(strings.NewReader(rawHeader)))
	// we intentionally ignore any errors here, not sure why...
	header, _ := headerReader.ReadMIMEHeader()
	return header
}

/** select key, value from http_headers_each(headers)
 * A table function for enumerating each header found in headers.
 */
var HeaderEachColumns = []vtab.Column{
	{Name: "headers", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
	{Name: "key", Type: sqlite.SQLITE_TEXT.String()},
	{Name: "value", Type: sqlite.SQLITE_TEXT.String()},
}

type HeaderEachCursor struct {
	header        textproto.MIMEHeader
	keyOrder      []string
	currentKeyI   int
	currentValueI int
}

func (cur *HeaderEachCursor) Column(ctx *sqlite.Context, c int) error {
	col := HeaderEachColumns[c]

	switch col.Name {
	case "key":
		ctx.ResultText(cur.keyOrder[cur.currentKeyI])
	case "value":
		ctx.ResultText(cur.header.Values(cur.keyOrder[cur.currentKeyI])[cur.currentValueI])
	}
	return nil
}

func (cur *HeaderEachCursor) Next() (vtab.Row, error) {
	cur.currentValueI += 1

	if cur.currentKeyI >= len(cur.keyOrder) {
		return nil, io.EOF
	}
	if cur.currentValueI >= len(cur.header.Values(cur.keyOrder[cur.currentKeyI])) {
		cur.currentKeyI += 1
		cur.currentValueI = 0
	}
	if cur.currentKeyI >= len(cur.keyOrder) {
		return nil, io.EOF
	}
	return cur, nil
}

func HeadersEachIterator(constraints []*vtab.Constraint, order []*sqlite.OrderBy) (vtab.Iterator, error) {
	var rawHeader string
	for _, constraint := range constraints {
		if constraint.Op == sqlite.INDEX_CONSTRAINT_EQ {
			column := HeaderEachColumns[constraint.ColIndex]
			switch column.Name {
			case "headers":
				rawHeader = constraint.Value.Text()
			}
		}
	}
	header := readHeader(rawHeader)

	cursor := HeaderEachCursor{
		header:        header,
		currentKeyI:   0,
		currentValueI: -1,
	}

	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	cursor.keyOrder = keys

	return &cursor, nil

}

/* http_headers_has(headers, key)
* Returns 1 if there is at least one header in headers with the given key,
* or, 0 otherwise.
* Key lookups are case-insensitive, like all HTTP headers
 */
type HeadersHasFunc struct{}

func (*HeadersHasFunc) Deterministic() bool { return true }
func (*HeadersHasFunc) Args() int           { return 2 }
func (*HeadersHasFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	headers := readHeader(values[0].Text())
	key := values[1].Text()

	matching := headers.Values(key)

	if len(matching) > 0 {
		c.ResultInt(1)
	} else {
		c.ResultInt(0)
	}
}

/* http_headers_get(headers, key)
 * Returns the first matching header's value, or null if none matches.
 * Key lookups are case-insensitive, like all HTTP headers
 */
type HeadersGetFunc struct{}

func (*HeadersGetFunc) Deterministic() bool { return true }
func (*HeadersGetFunc) Args() int           { return 2 }
func (*HeadersGetFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	headers := readHeader(values[0].Text())
	key := values[1].Text()

	matching := headers.Values(key)
	if len(matching) > 0 {
		c.ResultText(matching[0])
	} else {
		c.ResultNull()
	}
}

/* http_headers(name1, value1, ...)
 * Utilty for constructing headers in wire format.
 */
type HeadersFunc struct{}

func (*HeadersFunc) Deterministic() bool { return true }
func (*HeadersFunc) Args() int           { return -1 }
func (*HeadersFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {

	if len(values)%2 != 0 {
		c.ResultError(errors.New("http_headers must have even-numbered arguments"))
	}
	header := http.Header{}
	for i := 0; i < len(values); i = i + 2 {
		key := values[i].Text()
		value := values[i+1].Text()

		header.Add(key, value)
	}

	buf := new(bytes.Buffer)

	header.Write(buf)
	c.ResultText(buf.String())
}

func RegisterHeaders(api *sqlite.ExtensionApi) error {
	if err := api.CreateModule("http_headers_each", vtab.NewTableFunc("http_headers_each", HeaderEachColumns, HeadersEachIterator)); err != nil {
		return err
	}
	if err := api.CreateFunction("http_headers", &HeadersFunc{}); err != nil {
		return err
	}
	if err := api.CreateFunction("http_headers_has", &HeadersHasFunc{}); err != nil {
		return err
	}
	if err := api.CreateFunction("http_headers_get", &HeadersGetFunc{}); err != nil {
		return err
	}
	return nil
}
