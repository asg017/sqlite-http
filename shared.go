package main

import (
	"go.riyazali.net/sqlite"
)

// Set in Makefile
var (
	Commit  string
	Date    string
	Version string
	// TODO should be in seperate script - this works, but the generated binary
	// should have no reference to do functions at all, should be slimmed-down binary
	// Maybe a separate entrypoint?
	OmitNet string
)

func init() {
	sqlite.Register(func(api *sqlite.ExtensionApi) (sqlite.ErrorCode, error) {

		if err := RegisterMeta(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}
		
		// If the "-X main.OmitNet=1" flag was provided, then don't
		// include funcs that make network calls.
		if OmitNet != "1" {
			if err := RegisterDo(api); err != nil {
				return sqlite.SQLITE_ERROR, err
			}
			if err := RegisterSettings(api); err != nil {
				return sqlite.SQLITE_ERROR, err
			}	
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
