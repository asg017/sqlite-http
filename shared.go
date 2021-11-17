package main

import (
	cookies "github.com/asg017/sqlite-http/cookies"
	do "github.com/asg017/sqlite-http/do"
	headers "github.com/asg017/sqlite-http/headers"
	meta "github.com/asg017/sqlite-http/meta"
	"go.riyazali.net/sqlite"
)

// Set in Makefile
var (
	Commit  string
	Date    string
	Version string
	// TODO should be in seperate script - this works, but the generated binary 
	// should have no reference to do functions at all, should be slimmed-down binary
	OmitDo  string
)

func init() {
	omitDo := OmitDo == "1"
	sqlite.Register(func(api *sqlite.ExtensionApi) (sqlite.ErrorCode, error) {

		if err := meta.Register(api, meta.RegisterParams{
			Version: Version,
			Commit:  Commit,
			Date:    Date,
		}); err != nil {
			return sqlite.SQLITE_ERROR, err
		}

		if !omitDo {
			if err := do.Register(api); err != nil {
				return sqlite.SQLITE_ERROR, err
			}
		}
		if err := headers.Register(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}
		if err := cookies.Register(api); err != nil {
			return sqlite.SQLITE_ERROR, err
		}

		return sqlite.SQLITE_OK, nil
	})
}

func main() {}
