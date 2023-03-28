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
type DebugFunc struct{noNetwork bool}

func (*DebugFunc) Deterministic() bool { return true }
func (*DebugFunc) Args() int           { return 0 }
func (f *DebugFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	var format string;
	if f.noNetwork {
		format = "Version: %s\nCommit: %s\nRuntime: %s %s/%s\nDate: %s\nNO NETWORK"
	} else {
	format = "Version: %s\nCommit: %s\nRuntime: %s %s/%s\nDate: %s"
	}
	c.ResultText(fmt.Sprintf(format,
		Version,
		Commit,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		Date,
	))
}

func RegisterMeta(api *sqlite.ExtensionApi, noNetwork bool) error {
	if err := api.CreateFunction("http_version", &VersionFunc{}); err != nil {
		return err
	}
	if err := api.CreateFunction("http_debug", &DebugFunc{noNetwork}); err != nil {
		return err
	}
	return nil
}
