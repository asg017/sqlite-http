package main

import (
	"fmt"
	"runtime"

	"go.riyazali.net/sqlite"
)

/* http_version()
* Return the semver version string of the loaded sqlite-http library.
 */
type VersionFunc struct{}

func (*VersionFunc) Deterministic() bool { return true }
func (*VersionFunc) Args() int           { return 0 }
func (f *VersionFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	c.ResultText(Version)
}

/* http_debug()
* Return a debug string of the loaded sqlite-http library.
* Includes version string, build commit hash, go runtime info, and build date.
 */
type DebugFunc struct{}

func (*DebugFunc) Deterministic() bool { return true }
func (*DebugFunc) Args() int           { return 0 }
func (f *DebugFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	c.ResultText(fmt.Sprintf("Version: %s\nCommit: %s\nRuntime: %s %s/%s\nDate: %s\n",
		Version,
		Commit,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		Date,
	))
}

func RegisterMeta(api *sqlite.ExtensionApi) error {
	if err := api.CreateFunction("http_version", &VersionFunc{}); err != nil {
		return err
	}
	if err := api.CreateFunction("http_debug", &DebugFunc{}); err != nil {
		return err
	}
	return nil
}
