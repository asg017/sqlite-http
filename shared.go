package main

import (
	"go.riyazali.net/sqlite"
)

// following linker flags are needed to suppress missing symbol warning in intermediate stages

// #cgo linux LDFLAGS: -Wl,--unresolved-symbols=ignore-in-object-files
// #cgo darwin LDFLAGS: -Wl,-undefined,dynamic_lookup
// #cgo windows LDFLAGS: /ignore:4217
import "C"

// Set in Makefile
var (
	Commit  string
	Date    string
	Version string
)


func init() {
	sqlite.Register(func(api *sqlite.ExtensionApi) (sqlite.ErrorCode, error) {

		if err := RegisterMeta(api, false); err != nil {
			return sqlite.SQLITE_ERROR, err
		}	
		if err := RegisterHeaders(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}
		if err := RegisterCookies(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}
		if err := RegisterDo(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}
		if err := RegisterSettings(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}

		return sqlite.SQLITE_OK, nil
	})
	sqlite.RegisterNamed("http", func(api *sqlite.ExtensionApi) (sqlite.ErrorCode, error) {

		if err := RegisterMeta(api, false); err != nil {
			return sqlite.SQLITE_ERROR, err
		}	
		if err := RegisterHeaders(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}
		if err := RegisterCookies(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}
		if err := RegisterDo(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}
		if err := RegisterSettings(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}

		return sqlite.SQLITE_OK, nil
	})
	sqlite.RegisterNamed("http_no_network", func(api *sqlite.ExtensionApi) (sqlite.ErrorCode, error) {

		if err := RegisterMeta(api, true); err != nil {
			return sqlite.SQLITE_ERROR, err
		}

		if err := RegisterHeaders(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}
		if err := RegisterCookies(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}

		return sqlite.SQLITE_OK, nil
	})
}

func main() {}
