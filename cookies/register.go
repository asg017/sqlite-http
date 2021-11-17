package cookies

import (
	"go.riyazali.net/sqlite"
)

var modules = map[string]sqlite.Module{}

var functions = map[string]sqlite.Function{
	"http_cookies": &CookiesFunc{},
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
