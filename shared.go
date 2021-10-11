package main

import (
	headers "github.com/asg017/sqlite-http/headers"
	http_do "github.com/asg017/sqlite-http/http_do"
	"go.riyazali.net/sqlite"
)

func init() {

	sqlite.Register(func(api *sqlite.ExtensionApi) (sqlite.ErrorCode, error) {
		if err := http_do.Register(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}
		if err := headers.Register(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}
		
		return sqlite.SQLITE_OK, nil
	})
}

func main() {}
