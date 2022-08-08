# sqlite-http

A SQLite extension for making HTTP requests purely in SQL.

- Create GET, POST, or other HTTP requests and download responses, like `curl`, `wget`, and `fetch`
- Query HTTP headers, cookies, timing information
- Set rate limits, timeouts

## Usage

```sql
.load ./http0
select http_get_body('https://text.npr.org/');
/*
<!DOCTYPE html>
<html lang="en">
<head>
  <title>NPR : National Public Radio</title>
  ....
*/
```

## Documentation

See [`docs.md`](./docs.md) for a full API reference.

## Installing

The [Releases page](https://github.com/asg017/sqlite-lines/releases) contains pre-built binaries for Linux amd64, MacOS amd64 (no arm), and Windows.

### As a loadable extension

If you want to use `sqlite-http` as a [Runtime-loadable extension](https://www.sqlite.org/loadext.html), Download the `http0.dylib` (for MacOS), `http0.so` (Linux), or `http0.dll` (Windows) file from a release and load it into your SQLite environment.

> **Note:**
> The `0` in the filename (`http0.dylib`/ `http0.so`/`http0.dll`) denotes the major version of `sqlite-http`. Currently `sqlite-http` is pre v1, so expect breaking changes in future versions.

For example, if you are using the [SQLite CLI](https://www.sqlite.org/cli.html), you can load the library like so:

```sql
.load ./http0
select http_version();
-- v0.0.1
```

Or in Python, using the builtin [sqlite3 module](https://docs.python.org/3/library/sqlite3.html):

```python
import sqlite3

con = sqlite3.connect(":memory:")

con.enable_load_extension(True)
con.load_extension("./http0")

print(con.execute("select http_version()").fetchone())
# ('v0.0.1',)
```

Or in Node.js using [better-sqlite3](https://github.com/WiseLibs/better-sqlite3):

```javascript
const Database = require("better-sqlite3");
const db = new Database(":memory:");

db.loadExtension("./lines0");

console.log(db.prepare("select http_version()").get());
// { 'http_version()': 'v0.0.1' }
```

Or with [Datasette](https://datasette.io/) TODO:

```
datasette data.db --load-extension ./http0
```

## Testing

Testing the output `.so` is in `test.py`. Tests require a local instance of [httpbin](https://httpbin.org/) to work. If you have docker, run `docker httpbin` to start an instance on port 8080. If you want to skip tests that require httpbin (ex CI scripts), then set a `SKIP_DO` environment varaible to `""`, like sp:

```
SKIP_DO=1; python3 test.py
```

## See also

- [sqlite-html](https://github.com/asg017/sqlite-html), for parsing and querying HTML using CSS selectors in SQLite (pairs great with this tool)
- [pgsql-http](https://github.com/pramsey/pgsql-http), a similar yet very different HTTP library for POstgreSQL (didn't know about this before I started this, but interestingly enough came up with a very similar API)
- [riyaz-ali/sqlite](https://github.com/riyaz-ali/sqlite), the brilliant Go library that this library depends on
- [nalgeon/sqlean](https://github.com/nalgeon/sqlean), several pre-compiled handy SQLite functions, in C
