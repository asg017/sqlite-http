package cookies

import (
	"encoding/json"
	"errors"

	"go.riyazali.net/sqlite"
)

type CookiesFunc struct{}

func (*CookiesFunc) Deterministic() bool { return true }
func (*CookiesFunc) Args() int           { return -1 }
func (*CookiesFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	cookies := make(map[string]string)
	if len(values)%2 != 0 {
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
	} else {
		c.ResultText(string(txt))
	}
}
