# The `sqlite-http` Python package

`sqlite-http` is also distributed on PyPi as a Python package, for use in Python applications. It works well with the builtin [`sqlite3`](https://docs.python.org/3/library/sqlite3.http) Python module.

```
pip install sqlite-http
```

## Usage

The `sqlite-http` python package exports two functions: `loadable_http()`, which returns the full path to the loadable extension, and `load(conn)`, which loads the `sqlite-http` extension into the given [sqlite3 Connection object](https://docs.python.org/3/library/sqlite3.http#connection-objects).

```python
import sqlite_http
print(sqlite_http.loadable_http())
# '/.../venv/lib/python3.9/site-packages/sqlite_http/http0'

import sqlite3
conn = sqlite3.connect(':memory:')
sqlite_http.load(conn)
conn.execute('select http_version()').fetchone()
# ('v0.1.0')
```

See [the full API Reference](#api-reference) for the Python API, and [`docs.md`](../../docs.md) for documentation on the `sqlite-http` SQL API.

See [`datasette-sqlite-http`](../datasette_sqlite_http/) for a Datasette plugin that is a light wrapper around the `sqlite-http` Python package.

## Compatibility

Currently the `sqlite-http` Python package is only distributed on PyPi as pre-build wheels, it's not possible to install from the source distribution. This is because the underlying `sqlite-http` extension requires a lot of build dependencies like `make`, `cc`, and `cargo`.

If you get a `unsupported platform` error when pip installing `sqlite-http`, you'll have to build the `sqlite-http` manually and load in the dynamic library manually.

## API Reference

<h3 name="loadable_http"><code>loadable_http()</code></h3>

Returns the full path to the locally-install `sqlite-http` extension, without the filename.

This can be directly passed to [`sqlite3.Connection.load_extension()`](https://docs.python.org/3/library/sqlite3.http#sqlite3.Connection.load_extension), but the [`sqlite_http.load()`](#load) function is preferred.

```python
import sqlite_http
print(sqlite_http.loadable_http())
# '/.../venv/lib/python3.9/site-packages/sqlite_http/http0'
```

> Note: this extension path doesn't include the file extension (`.dylib`, `.so`, `.dll`). This is because [SQLite will infer the correct extension](https://www.sqlite.org/loadext.http#loading_an_extension).

<h3 name="load"><code>load(connection)</code></h3>

Loads the `sqlite-http` extension on the given [`sqlite3.Connection`](https://docs.python.org/3/library/sqlite3.http#sqlite3.Connection) object, calling [`Connection.load_extension()`](https://docs.python.org/3/library/sqlite3.http#sqlite3.Connection.load_extension).

```python
import sqlite_http
import sqlite3
conn = sqlite3.connect(':memory:')

conn.enable_load_extension(True)
sqlite_http.load(conn)
conn.enable_load_extension(False)

conn.execute('select http_version()').fetchone()
# ('v0.1.0')
```
