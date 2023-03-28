# The `datasette-sqlite-http` Datasette Plugin

`datasette-sqlite-http` is a [Datasette plugin](https://docs.datasette.io/en/stable/plugins.http) that loads the [`sqlite-http`](https://github.com/asg017/sqlite-http) extension in Datasette instances, allowing you to generate and work with [TODO](https://github.com/http/spec) in SQL.

```
datasette install datasette-sqlite-http
```

See [`docs.md`](../../docs.md) for a full API reference for the http SQL functions.

Alternatively, when publishing Datasette instances, you can use the `--install` option to install the plugin.

```
datasette publish cloudrun data.db --service=my-service --install=datasette-sqlite-http

```
