package main

import (
	"time"

	"go.riyazali.net/sqlite"
)

// #cgo linux LDFLAGS: -Wl,--unresolved-symbols=ignore-in-object-files
// #cgo darwin LDFLAGS: -Wl,-undefined,dynamic_lookup
import "C"

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

// timestamp layout to match SQLite's 'datetime()' format, ISO8601 subset
const sqliteDatetimeFormat = "2006-01-02 15:04:05.999"

// Format ghe given time as a SQLite date timestamp
func formatSqliteDatetime(t *time.Time) *string {
	s := t.UTC().Format(sqliteDatetimeFormat)
	return &s
}


func RegisterDefault(api *sqlite.ExtensionApi) (sqlite.ErrorCode, error) {
	if err := RegisterMeta(api); err != nil {
		return sqlite.SQLITE_ERROR, err
	}

	if err := RegisterDo(api); err != nil {
		return sqlite.SQLITE_ERROR, err
	}
	if err := RegisterSettings(api); err != nil {
		return sqlite.SQLITE_ERROR, err
	}
	if err := RegisterHeaders(api); err != nil {
		return sqlite.SQLITE_ERROR, err
	}
	if err := RegisterCookies(api); err != nil {
		return sqlite.SQLITE_ERROR, err
	}

	return sqlite.SQLITE_OK, nil
}

func RegisterNoNetwork(api *sqlite.ExtensionApi) (sqlite.ErrorCode, error) {

	if err := RegisterHeaders(api); err != nil {
		return sqlite.SQLITE_ERROR, err
	}

	return sqlite.SQLITE_OK, nil
}

func init() {
	sqlite.RegisterNamed("default", RegisterDefault)
	sqlite.RegisterNamed("no_network", RegisterNoNetwork)
}

func main() {}
