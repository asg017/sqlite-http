package headers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/augmentable-dev/vtab"
	"go.riyazali.net/sqlite"
)

//*
var HeaderEachColumns = []vtab.Column{
	{Name: "headers", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
	{Name: "onlyHeader", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: false, OmitCheck: true}}},
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
	onlyHeaderSet := false
	var onlyHeader string
	for _, constraint := range constraints {
		if constraint.Op == sqlite.INDEX_CONSTRAINT_EQ {
			column := HeaderEachColumns[constraint.ColIndex]
			switch column.Name {
			case "headers":
				rawHeader = constraint.Value.Text()
			case "onlyHeader":
				onlyHeader = constraint.Value.Text()
				onlyHeaderSet = true
			}
		}
	}
	header, err := readHeader(rawHeader)
	if err != nil {
		return nil, err
	} 
	
	cursor := HeaderEachCursor{
		header:        header,
		currentKeyI:   0,
		currentValueI: -1,
	}
	if onlyHeaderSet {
		cursor.keyOrder = []string{onlyHeader}
	} else {
		keys := make([]string, 0, len(header))
		for k := range header {
			keys = append(keys, k)
		}
		cursor.keyOrder = keys
	}
	return &cursor, nil

}

//*/

func readHeader(rawHeader string) (textproto.MIMEHeader, error) {
	headerReader := textproto.NewReader(bufio.NewReader(strings.NewReader(rawHeader)))
	return headerReader.ReadMIMEHeader()
}

type HeadersHasFunc struct{}

func (*HeadersHasFunc) Deterministic() bool { return true }
func (*HeadersHasFunc) Args() int           { return 2 }
func (*HeadersHasFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	headers, err := readHeader(values[0].Text())
	key := values[1].Text()

	if err != nil {
		c.ResultError(errors.New("1st argument not propery formatter header"))
		return
	}

	vals := headers.Values(key)

	if len(vals) > 0 {
		c.ResultInt(1)
	} else {
		c.ResultInt(0)
	}
}

type HeadersGetFunc struct{}

func (*HeadersGetFunc) Deterministic() bool { return true }
func (*HeadersGetFunc) Args() int           { return 2 }
func (*HeadersGetFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	headers, err := readHeader(values[0].Text())
	key := values[1].Text()

	if err != nil {
		c.ResultError(errors.New("1st argument not propery formatter header"))
		return
	}

	c.ResultText(headers.Get(key))
}

type HeadersAllFunc struct{}

func (*HeadersAllFunc) Deterministic() bool { return true }
func (*HeadersAllFunc) Args() int           { return 2 }
func (*HeadersAllFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	headers, err := readHeader(values[0].Text())
	key := values[1].Text()

	if err != nil {
		c.ResultError(errors.New("1st argument not propery formatter header"))
		return
	}

	all := headers.Values(key)
	b, err := json.Marshal(all)
	if err != nil {
		c.ResultError(err)
	}else {
		c.ResultText(string(b))
	}
}

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
