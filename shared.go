package main

import (
	"fmt"
	"runtime"

	headers "github.com/asg017/sqlite-http/headers"
	http_do "github.com/asg017/sqlite-http/http_do"
	"go.riyazali.net/sqlite"
)

// Set in Makefile
var (
	Commit  string
	Date    string
	Version string
)


// http_version
type VersionFunc struct{}
func (*VersionFunc) Deterministic() bool { return true }
func (*VersionFunc) Args() int           { return 0 }
func (*VersionFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	c.ResultText(Version)
}

// http_debug
type DebugFunc struct{}
func (*DebugFunc) Deterministic() bool { return true }
func (*DebugFunc) Args() int           { return 0 }
func (*DebugFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	c.ResultText(fmt.Sprintf("Version: %s\nCommit: %s\nRuntime: %s %s/%s\nDate: %s\n",
	Version,
	Commit,
	runtime.Version(),
	runtime.GOOS,
	runtime.GOARCH,
	Date,
))
}

var functions = map[string]sqlite.Function{
	"http_version": &VersionFunc{},
	"http_debug": &DebugFunc{},
}

func init() {

	sqlite.Register(func(api *sqlite.ExtensionApi) (sqlite.ErrorCode, error) {
		for name, function := range functions {
			if err := api.CreateFunction(name, function); err != nil {
				return sqlite.SQLITE_ERROR, err
			}
		}
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
