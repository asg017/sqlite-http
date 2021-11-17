package headers

import (
	"github.com/augmentable-dev/vtab"
	"go.riyazali.net/sqlite"
)

var functions = map[string]sqlite.Function{
	"http_headers":     &HeadersFunc{},
	"http_headers_has": &HeadersHasFunc{},
	"http_headers_get": &HeadersGetFunc{},
	"http_headers_all": &HeadersAllFunc{},
}

var modules = map[string]sqlite.Module{
	"http_headers_each": vtab.NewTableFunc("http_headers_each", HeaderEachColumns, HeadersEachIterator),
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
