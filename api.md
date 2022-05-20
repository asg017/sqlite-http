# API Reference

A full reference to every function and module that `sqlite-http` offers.

As a reminder, `sqlite-http` follows [semver](https://semver.org/) and is pre v1, so breaking changes are to be expected.

## Overview

- Request all contents from a URL (headers, body, timings, request metadata, etc)
  - [http_get](#)(_url, headers, cookies_)
  - [http_post](#)(_url, headers, body, cookies_)
  - [http_do](#)(_method, url, headers, body, cookies_)
- Request the body contents from a URL
  - [http_get_body](#http_get_body)(_url, headers, cookies_)
  - [http_post_body](#http_post_body)(_url, headers, body, cookies_)
  - [http_do_body](#http_do_body)(_method, url, headers, body, cookies_)
- Request the header contents from a URL
  - [http_get_headers](#http_get_headers)(\_url, headers,
  - [http_post_headers](#)(_url, headers, body, cookies_)
  - [http_do_headers](#)(_method, url, headers, body, cookies_)
- Utilities for crafting request bodies
  - [http_post_form_urlencoded](#)(_name1, value1, ..._)
- Create, query, and manipulate HTTP headers in wire format
  - [http_headers](#)(_label1, value1_)
  - [http_headers_has](#)(_headers, key_)
  - [http_headers_get](#)(_headers, get_)
  - [http_headers_all](#)(_headers, key_)
  - [http_headers_each](#)(_headers, onlyHeader_)
- Create, query, and manipulate HTTP cookies
  - [http_cookies](#)(_label1, value1_)
- Configure `sqlite-http` behavior
  - [http_rate_limit](#)(_duration_)
  - [http_timeout_set](#)(_duration_)
- `sqlite-http` information
  - [http_version](#http_version)()
  - [http_debug](#http_debug)()

## Interface Overview

### Making Requests

A single HTTP request is made with all of the following functions:

- `http_get` (table function)
- `http_post` (table function)
- `http_do` (table function)
- `http_get_body`
- `http_post_body`
- `http_do_body`
- `http_get_headers`
- `http_post_headers`
- `http_do_headers`

Refer to each function's documentation below for more specifics. A version of this extension called `http0-no-net.so` removes these functions, in case you want to distribute the utility functions found in this project (`http_headers_get`, `http_headers_each`, etc.) to a broader audience, but don't want to become a vector for DDoS attacks. This is especially helpful when using [Datasette](https://datasette.io/).

By default, no rate-limiting is set in this library. So if you run:

```sql
select http_get_body(
  printf("http://localhost:8000/%d", value)
from generate_series(1, 1e7);
);
```

This will sent 1 million GET requests to `localhost:8000` sequentially with no delay. If you want to introduce a delay between requests, use `http_rate_limit` like so:

```sql
select http_rate_limit(250);

select http_get_body(
  printf("http://localhost:8000/%d", value)
from generate_series(1, 1e7);
)
```

This will still send 1 million GET requests, but will pause 250 milliseconds between request.

Similarly, the default timeout for HTTP requests is 30 seconds. To change that, use `http_timeout_set`:

```sql
select http_timeout_set(5 * 1000);

-- will return an error if this takes more than 5 seconds
select http_get_body("http://localhost:8000");
```

### HEADERS arguments

All request functions take some form of a "headers" parameter. Headers should be provided in "wire" format. The `http_headers` function can provide this, like so:

```sql
select http_get_body(
  "http://localhost:8000",
  http_headers("Name", "alex", "Username", "asg017")
);
```

### COOKIES arguments

All request methods also support a cookies argument, to send cookies alongside a request. This is still being worked on so it's unstable, but the `http_cookies` function creates cookies that can be sent along.

## Details

### Request Everything

The `http_get`, `http_post`, and `http_do` table functions will make HTTP requests and create a single-row table of information on the generated request, response, and metadata. These functions are [table-valued functions](https://www.sqlite.org/vtab.html#tabfunc2), not scalar/window functions, so they work differently than other functions in this library.

Every call to one of these table functions returns a single row. Each of these functions generate a table with this schema:

```SQL
-- same for http_post and http_do
CREATE TABLE http_get(
  request_url TEXT,         -- URL that is request
  request_method TEXT,      -- HTTP method used in request
  request_headers TEXT,     -- Request HTTP headers, in wire format
  request_cookies TEXT,     -- Cookies sent in request (unstable)
  request_body BLOB,        -- Body sent in request
  response_status TEXT,     -- Status text of the response ("200 OK")
  response_status_code INT, -- HTTP status code of the response (100-999)
  response_headers TEXT,    -- Response HTTP headers, in wire format
  response_cookies TEXT,    -- Cookies received in response (unstable)
  response_body,            -- Body received in response
  remote_address TEXT,      -- IP address of responding server
  timings TEXT,             -- JSON of timestamp of various events
  meta TEXT                 -- Metadata of request
);
```

The `request_url` column contains the URL of the generated request (aka the 1st or 2nd argument to the table function).

The `request_method` column is the HTTP method used by the request, typically `"GET"` if using `http_get`, `"POST"` for `"http_post"`, or a custom method when using `http_do` (`"PATCH"`, `"DELETE"`, etc).

The `request_headers` column

The `request_cookies` column

The `request_body` column

The `response_status` column contains the HTTP status received by the response, like `"200 OK"`.

The `response_status_code` column

The `response_headers` column

The `response_cookies` column

The `response_body` column

The `remote_address` column is the IP address that the HTTP request connected to.

The `timings` column contains timestamps of when specific events happened while making the request. It is a JSON object, where the keys are the name of the event that happened, and the values are the string timestamps of when it occured (in ISO-8601, same format as sqlite's [date functions](https://www.sqlite.org/lang_datefunc.html)). Most of these are obtained using Go's [httptrace](https://pkg.go.dev/net/http/httptrace), so refer to that for more information. The events are:

- `"start"` - When the request is first made
- `"first_byte"` - _"when the first byte of the response headers is available"_
- `"connection"` - _""_
- `"wrote_headers"` - _""_
- `"dns_start"` - _"when a DNS lookup begins"_
- `"dns_end"` - _"when a DNS lookup ends"_
- `"connect_start"` - _"when a new connection's Dial begins"_
- `"connect_end"` - _"when a new connection's Dial completes"_
- `"tls_handshake_start"` - _"when the TLS handshake is started"_
- `"tls_handshake_end"` - _"after the TLS handshake with either the successful handshake's connection state, or a non-nil error on handshake failure."_
- `"body_start"` - When the `response_body` column is accessed, if at all. This is when `sqlite-http` reads in the body to memory
- `"body_end"` - After the entire reponse body is read into memory, right before it's returned to SQLite

The `meta` column is null for now. In the future, this may include more metadata about a request.

These table functions can be used like so:

```sql
sqlite> select request_url, response_status, length(response_body) from http_get("https://google.com");
https://google.com,"200 OK",14995
```

### Requesting only body

`http_get_body`, `http_post_body`, and `http_do_body` are similar to their table function counterparts, but instead they are scalar functions that only return the response body of the given request. These are good to use for one-off requests, or if you don't care about other information like headers, cookies, timings, etc.

```sql
sqlite> select http_();

```

```sql
sqlite> select http_get_body("https://dog.ceo/api/breeds/image/random");
{"message":"https:\/\/images.dog.ceo\/breeds\/komondor\/n02105505_3967.jpg","status":"success"}
```

#### `http_get_body`

#### `http_post_body`

#### `http_do_body`

### Requesting only headers

with `http_get_headers`, `http_post_headers`, and `http_do_headers`

```sql
sqlite> select http_();

```

```sql
sqlite> select http_();

```

### Request body utilities

More utility functions may be added in the future. Follow [#3](https://github.com/asg017/sqlite-http/issues/3) for more info.

#### `http_post_form_urlencoded`

### HTTP Headers

More header utilities may be added in the future. Follow [#23](https://github.com/asg017/sqlite-http/issues/23) for more info.

#### `http_headers`

#### `http_headers_has`

`http_headers_has(headers, name)`

#### `http_headers_get`

`http_headers_get(headers, name)`

#### `http_headers_all`

`http_headers_has(???)`

#### `http_headers_each`

### HTTP Cookies

More cookie utilities may be added in the future. Follow [#24](https://github.com/asg017/sqlite-http/issues/24) for more info.

#### `http_cookies`

### Configuring `sqlite-http` Behavior

#### `http_rate_limit`

#### `http_timeout_set`

### `sqlite-http` Information

#### `http_version`

Returns the version string of the `sqlite-http` library, modeled after [`sqlite_version()`](https://www.sqlite.org/lang_corefunc.html#sqlite_version).

```sql
sqlite> select http_version();
"v0.0.0"
```

#### `http_debug`

Returns debug information of the `sqlite-http` library, including the version string. Subject to change.

```sql
sqlite> select http_debug();
Version: "v0.0.0"
Commit: 85445c7b9d539e2626731275da6496d96d6dbb05
Runtime: go1.17 darwin/amd64
Date: 2021-11-17T16:20:06Z-0800
```
