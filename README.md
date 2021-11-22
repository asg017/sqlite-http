# sqlite-http

A SQLite extension for making HTTP requests purely in SQL.

- Create GET, POST, or any other HTTP requests and download responses, like `curl`, `wget`, and `fetch`
- Query HTTP headers, cookies, timing information
- Set rate limits, timeouts

## 🚧🚧 Work In Progress! 🚧🚧

This library is experimental and subject to change. I plan to make a stable beta release and subsequent v0+v1 in the near future, so use with caution.

When v0 is ready (with a mostly stable API), I will make a release (so watch this repo for that) and will make a blog post, feel free to [follow me on twitter](https://twitter.com/agarcia_me) to get notified of that.

## Installing

TODO

## Documentation

See [`api.md`](./api.md) for a full API Reference.

## Examples

First, let's load the extension using the [`.load`](https://www.sqlite.org/cli.html#loading_extensions) command in SQLite's CLI.

```sql
.load http0.so
```

> Note: Loading extensions may be disabled by default in your computer's builtin `sqlite3` CLI. Consider downloading a [newer version of SQLite](https://sqlite.org/download.html). Also look into your favorite SQLite client library API for how to load extensions, like Python's [`load_extension`](https://docs.python.org/3/library/sqlite3.html#sqlite3.Connection.load_extension), Node.js's [`loadExtesion`](https://github.com/JoshuaWise/better-sqlite3/blob/master/docs/api.md#loadextensionpath-entrypoint---this) in `better-sqlite`, or the [`load_extension`](https://www.sqlite.org/lang_corefunc.html#load_extension) function in some SQLite libraries.

## Testing

Testing the output `.so` is in `test.py`. Tests require a local instance of [httpbin](https://httpbin.org/) to work. If you have docker, run `docker httpbin` to start an instance on port 8080. If you want to skip tests that require httpbin (ex CI scripts), then set a `SKIP_DO` environment varaible to `""`, like sp:

```
SKIP_DO=1; python3 test.py
```

## See also

- [sqlite-html](https://github.com/asg017/sqlite-html), for parsing and querying HTML using CSS selectors in SQLite (pairs great with this tool)
- [pgsql-http](https://github.com/pramsey/pgsql-http), a similar yet very different HTTP libraryt for POstgreSQL (didn't know about this before I started this, but interestingly enough came up with a very similar API)
- [riyaz-ali/sqlite](https://github.com/riyaz-ali/sqlite), the brilliant Go library that this library depends on
- [nalgeon/sqlean](https://github.com/nalgeon/sqlean), several pre-compiled handy SQLite functions, in C
