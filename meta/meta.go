package meta

import (
	"fmt"
	"io"
	"runtime"

	"github.com/augmentable-dev/vtab"
	"go.riyazali.net/sqlite"
)

// http_version
type VersionFunc struct {
	version string
}

func (*VersionFunc) Deterministic() bool { return true }
func (*VersionFunc) Args() int           { return 0 }
func (f *VersionFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	c.ResultText(f.version)
}

// http_debug
type DebugFunc struct {
	version string
	commit  string
	date    string
}

func (*DebugFunc) Deterministic() bool { return true }
func (*DebugFunc) Args() int           { return 0 }
func (f *DebugFunc) Apply(c *sqlite.Context, values ...sqlite.Value) {
	c.ResultText(fmt.Sprintf("Version: %s\nCommit: %s\nRuntime: %s %s/%s\nDate: %s\n",
		f.version,
		f.commit,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		f.date,
	))
}

var DocsColumns = []vtab.Column{
	{Name: "headers", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: true, OmitCheck: true}}},
	{Name: "onlyHeader", Type: sqlite.SQLITE_TEXT.String(), NotNull: true, Hidden: true, Filters: []*vtab.ColumnFilter{{Op: sqlite.INDEX_CONSTRAINT_EQ, Required: false, OmitCheck: true}}},
	{Name: "key", Type: sqlite.SQLITE_TEXT.String()},
	{Name: "value", Type: sqlite.SQLITE_TEXT.String()},
}

type DocsCursor struct {
	current int
}

func (cur *DocsCursor) Column(ctx *sqlite.Context, c int) error {
	col := DocsColumns[c]

	switch col.Name {
	case "key":
		ctx.ResultText("")
	}
	return nil
}

func (cur *DocsCursor) Next() (vtab.Row, error) {
	cur.current += 1

	if cur.current >= 10 {
		return nil, io.EOF
	}
	return cur, nil
}

func DocsEachIterator(constraints []*vtab.Constraint, order []*sqlite.OrderBy) (vtab.Iterator, error) {
	var name string
	for _, constraint := range constraints {
		if constraint.Op == sqlite.INDEX_CONSTRAINT_EQ {
			column := DocsColumns[constraint.ColIndex]
			switch column.Name {
			case "headers":
				name = constraint.Value.Text()
			}
		}
	}
	cursor := DocsCursor{
		current: 0,
	}
	fmt.Println(name)
	return &cursor, nil

}
