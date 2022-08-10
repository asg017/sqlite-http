# sqlite-http

A SQLite extension for making HTTP requests purely in SQL.

- Create GET, POST, and other HTTP requests, like `curl`, `wget`, and `fetch`
- Download response bodies, header, status codes, timing info
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

Query for all custom headers in an endpoint.

```sql
select name, value
from http_headers_each(
  http_get_headers('https://api.github.com/')
)
where name like 'X-%';
/*
┌────────────────────────┬────────────────────────────────────┐
│          name          │               value                │
├────────────────────────┼────────────────────────────────────┤
│ X-Ratelimit-Limit      │ 60                                 │
│ X-Ratelimit-Used       │ 8                                  │
│ X-Content-Type-Options │ nosniff                            │
│ X-Github-Media-Type    │ github.v3; format=json             │
│ X-Github-Request-Id    │ CCCA:5FDF:1014BC2:10965F9:62F3DE4E │
│ X-Ratelimit-Remaining  │ 52                                 │
│ X-Ratelimit-Resource   │ core                               │
│ X-Frame-Options        │ deny                               │
│ X-Ratelimit-Reset      │ 1660152798                         │
│ X-Xss-Protection       │ 0                                  │
└────────────────────────┴────────────────────────────────────┘
*/
```

Scrape data from a JSON endpoint.

```sql
select http_get_body('https://api.github.com/repos/sqlite/sqlite')
  ->> '$.description' as description;
/*
┌───────────────────────────────────────────────┐
│                  description                  │
├───────────────────────────────────────────────┤
│ Official Git mirror of the SQLite source tree │
└───────────────────────────────────────────────┘
*/
```

Pass in specific headers in a request.

```sql
select
  value
from json_each(
  http_get_body(
    'https://api.github.com/issues',
    http_headers(
      'Authorization', 'token ghp_16C7e42F292c6912E7710c8'
    )
  )
);

```

## Documentation

See [`docs.md`](./docs.md) for a full API reference.

## Installing

The [Releases page](https://github.com/asg017/sqlite-http/releases) contains pre-built binaries for Linux amd64, MacOS amd64 (no arm), and Windows.

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

db.loadExtension("./http0");

console.log(db.prepare("select http_version()").get());
// { 'http_version()': 'v0.0.1' }
```

Or with [Datasette](https://datasette.io/), with the "no network" option to limit DDoS attacks:

```
datasette data.db --load-extension ./http0-no-net
```

## See also

- [sqlite-html](https://github.com/asg017/sqlite-html), for parsing and querying HTML using CSS selectors in SQLite (pairs great with this tool)
- [pgsql-http](https://github.com/pramsey/pgsql-http), a similar yet very different HTTP library for POstgreSQL (didn't know about this before I started this, but interestingly enough came up with a very similar API)
- [riyaz-ali/sqlite](https://github.com/riyaz-ali/sqlite), the brilliant Go library that this library depends on
- [nalgeon/sqlean](https://github.com/nalgeon/sqlean), several pre-compiled handy SQLite functions, in C
