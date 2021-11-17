package meta

import (
	"go.riyazali.net/sqlite"
)

var modules = map[string]sqlite.Module{}

type RegisterParams struct {
	Version string
	Commit  string
	Date    string
}

func buildFunctions(params RegisterParams) map[string]sqlite.Function {
	return map[string]sqlite.Function{
		"http_version": &VersionFunc{version: params.Version},
		"http_debug":   &DebugFunc{version: params.Version, date: params.Date, commit: params.Commit},
	}
}

func Register(api *sqlite.ExtensionApi, params RegisterParams) error {
	functions := buildFunctions(params)

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
