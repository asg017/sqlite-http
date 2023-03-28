<!--- Generated with the deno_generate_package.sh script, don't edit by hand! -->

# `x/sqlite_http` Deno Module

[![Tags](https://img.shields.io/github/release/asg017/sqlite-http)](https://github.com/asg017/sqlite-http/releases)
[![Doc](https://doc.deno.land/badge.svg)](https://doc.deno.land/https/deno.land/x/sqlite-http@0.0.1/mod.ts)

The [`sqlite-http`](https://github.com/asg017/sqlite-http) SQLite extension is available to Deno developers with the [`x/sqlite_http`](https://deno.land/x/sqlite-http) Deno module. It works with [`x/sqlite3`](https://deno.land/x/sqlite3), the fastest and native Deno SQLite3 module.

```js
import { Database } from "https://deno.land/x/sqlite3@0.8.0/mod.ts";
import * as sqlite_http from "https://deno.land/x/sqlite_http@v0.0.1/mod.ts";

const db = new Database(":memory:");

  db.enableLoadExtension = true;
  db.loadExtension(sqlite_http.getLoadablePath());

  const [version] = db
    .prepare("select http_version()")
    .value<[string]>()!;

  console.log(version);

```

Like `x/sqlite3`, `x/sqlite_http` requires network and filesystem permissions to download and cache the pre-compiled SQLite extension for your machine. Though `x/sqlite3` already requires `--allow-ffi` and `--unstable`, so you might as well use `--allow-all`/`-A`.

```bash
deno run -A --unstable <file>
```

`x/sqlite_http` does not work with [`x/sqlite`](https://deno.land/x/sqlite@v3.7.0), which is a WASM-based Deno SQLite module that does not support loading extensions.