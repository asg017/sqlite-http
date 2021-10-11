package headers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/textproto"
	"strings"

	"go.riyazali.net/sqlite"
)


func readHeader(rawHeader string) (textproto.MIMEHeader, error) {
	headerReader := textproto.NewReader(bufio.NewReader(strings.NewReader(rawHeader)))
	header, _ := headerReader.ReadMIMEHeader()
	return header, nil
}

type HeadersHasFunc struct{}
func (*HeadersHasFunc) Deterministic() bool { return true }
func (*HeadersHasFunc) Args() int           { return 2 }
func (*HeadersHasFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	headers, err := readHeader(values[0].Text())
	key := values[1].Text();

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
	key := values[1].Text();

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
	b, _ := json.Marshal(all)
	c.ResultText(string(b))
}

type HeadersFunc struct{}
func (*HeadersFunc) Deterministic() bool { return true }
func (*HeadersFunc) Args() int           { return -1 }
func (*HeadersFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {

	if len(values) % 2 != 0 {
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


type CookiesFunc struct{}
func (*CookiesFunc) Deterministic() bool { return true }
func (*CookiesFunc) Args() int           { return -1 }
func (*CookiesFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	cookies := make(map[string]string)
	if len(values) % 2 != 0 {
		c.ResultError(errors.New("http_cookies must have even-numbered arguments"))
	}
	for i := 0; i < len(values); i = i + 2 {
		key := values[i].Text()
		value := values[i+1].Text()

		cookies[key] = value
	}
	txt, err := json.Marshal(cookies)

	if err != nil {
		c.ResultError(err)
	}else {
		c.ResultText(string(txt))
	}
}


var functions = map[string]sqlite.Function{
	"http_headers": &HeadersFunc{},
	"http_headers_has": &HeadersHasFunc{},
	"http_headers_get": &HeadersGetFunc{},
	"http_headers_all": &HeadersAllFunc{},
	"http_cookies": &CookiesFunc{},
}
func Register(api *sqlite.ExtensionApi) (error) {
	for name, function := range functions {
		if err := api.CreateFunction(name, function); err != nil {
			return err
		}
	}
	
	return nil
}