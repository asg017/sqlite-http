package main

import (
	"encoding/json"
	"errors"

	"go.riyazali.net/sqlite"
)

// http_cookies(name1, value2, ...)
// Returns a JSON object where keys nameN, and vlaues are valueN
// Meant as a parameter for request functions like http_get and http_get_body
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

func RegisterCookies(api *sqlite.ExtensionApi) error {
	if err := api.CreateFunction("http_cookies", &CookiesFunc{}); err != nil {
		return err
	}

	return nil
}
