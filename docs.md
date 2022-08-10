# API Reference

A full reference to every function and module that `sqlite-http` offers.

As a reminder, `sqlite-http` follows [semver](https://semver.org/) and is pre v1, so breaking changes are to be expected.

## Overview

- Request all contents from a URL (headers, body, timings, request metadata, etc)
  - [http_get](#http_get)(_url, [headers], [cookies]_)
  - [http_post](#http_post)(_url, [headers], [body], [cookies]_)
  - [http_do](#http_do)(_method, url, [headers], [body], [cookies]_)
- Request the body contents from a URL
  - [http_get_body](#http_get_body)(_url, [headers], [cookies]_)
  - [http_post_body](#http_post_body)(_url, [headers], [body], [cookies]_)
  - [http_do_body](#http_do_body)(_method, url, [headers], [body], [cookies]_)
- Request the header contents from a URL
  - [http_get_headers](#http_get_headers)(_url, [headers], [cookies]_)
  - [http_post_headers](#http_post_headers)(_url, [headers], [body], [cookies]_)
  - [http_do_headers](#http_do_headers)(_method, url, [headers], [body], [cookies]_)
- Utilities for crafting request bodies
  - [http_post_form_urlencoded](#http_post_form_urlencoded)(_name1, value1, ..._)
- Create, query, and manipulate HTTP headers in wire format
  - [http_headers](#http_headers)(_name1, value1_)
  - [http_headers_has](#http_headers_has)(_headers, name_)
  - [http_headers_get](#http_headers_get)(_headers, get_)
  - [http_headers_each](#http_headers_each)(_headers_)
- Create, query, and manipulate HTTP cookies
  - [http_cookies](#http_cookies)(_label1, value1, [...]_)
- Configure `sqlite-http` behavior
  - [http_rate_limit](#http_rate_limit)(_duration_ms_)
  - [http_timeout_set](#http_timeout_set)(_duration_ms_)
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

Refer to each function's documentation below for more specifics. This extension can be compiled to remove these functions, in case you want to distribute the utility functions found in this project (`http_headers_get`, `http_headers_each`, etc.) to a broader audience, but don't want to become a vector for DDoS attacks. This is especially helpful when using [Datasette](https://datasette.io/).

By default, no rate-limiting is set in this library. So if you run:

```sql
select http_get_body(
  printf('http://localhost:8000/%d', value)
from generate_series(1, 1e7);
);
```

This will sent 1 million GET requests to `localhost:8000` sequentially with no delay. If you want to introduce a delay between requests, use `http_rate_limit` like so:

```sql
select http_rate_limit(250);

select http_get_body(
  printf('http://localhost:8000/%d', value)
from generate_series(1, 1e7);
)
```

This will still send 1 million GET requests, but will pause 250 milliseconds between request.

Similarly, the default timeout for HTTP requests is )0 seconds. To change that, use [`http_timeout_set`](#http_timeout_set):

```sql
select http_timeout_set(5 * 1000);

-- will return an error if this takes more than 5 seconds
select http_get_body('http://localhost:8000');
```

### HEADERS arguments

All request functions take some form of a "headers" parameter. Headers should be provided in "wire" format. The [`http_headers`](http_headers) function can provide this, like so:

```sql
select http_get_body(
  'http://localhost:8000',
  http_headers(
    'Name', 'alex',
    'Username', 'asg017'
  )
);
```

### COOKIES arguments

All request methods also support a cookies argument, to send cookies alongside a request. This is still being worked on so it's unstable, but the `http_cookies` function creates cookies that can be sent along.

<h3 name="no-net"> "No network" compile time option</h3>

sqlite-http can be compiled with the `-X main.OmitNet=1` option, which disables all functions that make HTTP requests like `http_get()`, `http_get_body()`, etc. This is because in some SQLite environments, untrusted users can execute arbitrary SQL code, which can become a security issue. However, it can still be useful to include other sqlite-http functions like `http_headers_each()` or `http_headers_date()`, which don't make HTTP requests.

The [Releases page](https://github.com/asg017/sqlite-http/releases) contain these "no network" compiled extensions with a `-no-net` suffix, if your use-case requires it.

## Error States

All table and scalar functions that make HTTP requests will fail and error in the following circumstances:

- Network errors (DNS, connections, etc.)
- Timeout errors (default 5 seconds)

Other "errors" like 500 or 400 status codes will _not_ result in a SQLite error. If you need to track and perform special behavior on non-200 status codes, consider something like:

```sql
select
  iif(
    response_status_code not between 200 and 299,
    'An error happened!',
    null
  ) as status,
  response_body
from http_get("http://httpbin.org/status/503");

/*
┌────────────────────┬───────────────┐
│       status       │ response_body │
├────────────────────┼───────────────┤
│ An error happened! │               │
└────────────────────┴───────────────┘
*/
```

## Details

### Request Everything

The `http_get`, `http_post`, and `http_do` table functions will make HTTP requests and create a single-row table of information on the generated request, response, and metadata. These are [table functions](https://www.sqlite.org/vtab.html#tabfunc2), so they work differently than other functions in this library.

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
  response_body BLOB,       -- Body received in response
  remote_address TEXT,      -- IP address of responding server
  timings TEXT,             -- JSON of various event timestamps
  meta TEXT                 -- Metadata of request
);
```

The `request_url` column contains the URL of the generated request (aka the 1st or 2nd argument to the table function).

The `request_method` column is the HTTP method used by the request, typically `"GET"` if using `http_get`, `"POST"` for `"http_post"`, or a custom method when using `http_do` (`"PATCH"`, `"DELETE"`, etc).

The `request_headers` column contain the HTTP headers sent with the initial request. Keep in mind, if no `User-Agent` is provided, the default is `"Go-http-client/1.1"`, and will _not_ appear in this column.

The `request_cookies` column is a work-in-progress. Will eventually be a JSON array of cookies sent with the request.

The `request_body` column contains the raw HTTP body sent along with the request, or NULL or none was sent.

The `response_status` column contains the HTTP status received by the response, like `"200 OK"`.

The `response_status_code` column contains the [HTTP status code](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status) received by the response, like `200`.

The `response_headers` column contains the HTTP headers from the response, in wire format.

The `response_cookies` column contains a JSON array of the HTTP cookies sent with the response.

The `response_body` column contains the response body as a BLOB.

The `remote_address` column is the IP address that the HTTP request connected to.

The `timings` column contains timestamps of when specific events happened while making the request. It is a JSON object, where the keys are the name of the event that happened, and the values are the string timestamps of when it occured (in ISO-8601, same format as sqlite's [date functions](https://www.sqlite.org/lang_datefunc.html)). Most of these are obtained using Go's [httptrace](https://pkg.go.dev/net/http/httptrace), so refer to that for more information. The events are:

- `"start"` - _When the request is first made_
- `"dns_start"` - _"when a DNS lookup begins"_
- `"dns_end"` - _"when a DNS lookup ends"_
- `"connect_start"` - _"when a new connection's Dial begins"_
- `"connect_end"` - _"when a new connection's Dial completes"_
- `"tls_handshake_start"` - _"when the TLS handshake is started"_
- `"connection"` - _"after a successful connection is obtained"_
- `"wrote_headers"` - _"after the Transport has written all request headers"_
- `"tls_handshake_end"` - _"after the TLS handshake with either the successful handshake's connection state, or a non-nil error on handshake failure."_
- `"first_byte"` - _"when the first byte of the response headers is available"_
- `"body_start"` - _When the `response_body` column is accessed, if at all. This is when `sqlite-http` reads in the body to memory_
- `"body_end"` - _After the entire reponse body is read into memory, right before it's returned to SQLite_

The `meta` column is null for now. In the future, this may include more metadata about a request.

These table functions can be used like so:

```sql
select
  request_url,
  response_status,
  length(response_body)
from http_get('https://google.com');
/*
┌────────────────────┬─────────────────┬───────────────────────┐
│    request_url     │ response_status │ length(response_body) │
├────────────────────┼─────────────────┼───────────────────────┤
│ https://google.com │ 200 OK          │ 14043                 │
└────────────────────┴─────────────────┴───────────────────────┘
*/
```

<h4 name="http_get"> <code>http_get(url, [headers], [cookies])</code></h4>

```sql
select * from http_get('http://httpbin.org/get');
```

<h4 name="http_post"> <code>http_post(url, [headers], [body], cookies])</code></h4>

```sql
select * from http_post('http://httpbin.org/post');
```

<h4 name="http_do"> <code>http_do(method, url, [headers], [body], [cookies])</code></h4>

```sql
select * from http_do('delete', 'http://httpbin.org/delete');
```

### Requesting only body

`http_get_body()`, `http_post_body()`, and `http_do_body()` are similar to their table function counterparts, but instead are scalar functions that only return the response body of the given request. These are good to use for one-off requests, or if you don't care about other information like headers, cookies, timings, etc.

<h4 name="http_get_body"> <code>http_get_body(url, [headers], [cookies])</code></h4>

Perform a GET request on the given URL, and return the response body.

```sql
select http_get_body('https://dog.ceo/api/breeds/image/random');
/*
{
  "message":"https://images.dog.ceo/breeds/komondor/n02105505_3967.jpg",
  "status":"success"
}
*/
```

<h4 name="http_post_body"> <code>http_post_body(url, [headers], [body], [cookies])</code></h4>

Perform a POST request on the given URL, and return the response body.

```sql
select http_post_body(
  'https://httpbin.org/post',
  http_headers('X-foo', 'bar'),
  json_object('name', 'alex')
);
/*
  "args": {},
  "data": "{\"name\":\"alex\"}",
  "files": {},
  "form": {},
  "headers": {
    "Accept-Encoding": "gzip",
    "Content-Length": "15",
    "Host": "httpbin.org",
    "User-Agent": "Go-http-client/2.0",
    "X-Amzn-Trace-Id": "Root=1-62eaee84-3b9060c003da2dca27ce8160",
    "X-Foo": "bar"
  },
  "json": {
    "name": "alex"
  },
  "origin": "XXXXX",
  "url": "https://httpbin.org/post"
}
*/
```

<h4 name="http_do_body"> <code>http_do_body(method, url, [headers], [body], [cookies])</code></h4>

Perform a request on the given URL with the given method, and return the response body.

```sql
select http_do_body(
  'DELETE',
  'https://httpbin.org/delete',
  http_headers('X-foo', 'bar'),
  json_object('name', 'alex')
);
/*
{
  "args": {},
  "data": "{\"name\":\"alex\"}",
  "files": {},
  "form": {},
  "headers": {
    "Accept-Encoding": "gzip",
    "Content-Length": "15",
    "Host": "httpbin.org",
    "User-Agent": "Go-http-client/2.0",
    "X-Amzn-Trace-Id": "Root=1-62eaeebc-5cef389c08b97210615d3a38",
    "X-Foo": "bar"
  },
  "json": {
    "name": "alex"
  },
  "origin": "XXXXXX",
  "url": "https://httpbin.org/delete"
}*/
```

### Requesting only headers

`http_get_headers()`, `http_post_headers()`, and `http_do_headers()` are similar to the "body" counterparts, but instead return only the headers of the reponse in wire format.

<h4 name="http_get_headers"> <code>http_get_headers(url, [headers], [cookies])</code></h4>

Perform a GET request on the given URL, and return the response headers.

```sql
select http_get_headers('http://httpbin.org/get');
/*
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: *
Connection: keep-alive
Content-Length: 271
Content-Type: application/json
Date: Mon, 08 Aug 2022 17:41:03 GMT
Server: gunicorn/19.9.0
*/
```

<h4 name="http_post_headers"> <code>http_post_headers(url, [headers], [body], [cookies])</code></h4>

Perform a POST request on the given URL, and return the response headers.

```sql
select http_post_headers('http://httpbin.org/post');
/*
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: *
Connection: keep-alive
Content-Length: 363
Content-Type: application/json
Date: Mon, 08 Aug 2022 17:41:31 GMT
Server: gunicorn/19.9.0
*/
```

<h4 name="http_do_headers"> <code>http_do_headers(method, url, [headers], [body], [cookies])</code></h4>

Perform a request on the given URL with the given method, and return the response headers.

```sql
select http_do_headers('DELETE', 'http://httpbin.org/delete');
/*
Access-Control-Allow-Credentials: true
Access-Control-Allow-Origin: *
Connection: keep-alive
Content-Length: 337
Content-Type: application/json
Date: Mon, 08 Aug 2022 17:41:56 GMT
Server: gunicorn/19.9.0
*/
```

### Request body utilities

More utility functions may be added in the future. Follow [#3](https://github.com/asg017/sqlite-http/issues/3) for more info.

<h4 name="http_post_form_urlencoded"> <code>http_post_form_urlencoded(key1, value1, ...)</code></h4>

Encodes the given keys and values into a url-encoded string, mainly for use in `application/x-www-form-urlencoded`.

```sql
select http_post_form_urlencoded(
  'name', 'Alex',
  'age', 99
);
-- 'age=99&name=Alex'


select http_post_body(
  'http://httpbin.org/post',
  http_headers('Content-Type', 'application/x-www-form-urlencoded'),
  http_post_form_urlencoded(
    'name', 'Alex',
    'age', 99
  )
);
/*
{
  "args": {},
  "data": "",
  "files": {},
  "form": {
    "age": "99",
    "name": "Alex"
  },
  "headers": {
    "Accept-Encoding": "gzip",
    "Content-Length": "16",
    "Content-Type": "application/x-www-form-urlencoded",
    "Host": "httpbin.org",
    "User-Agent": "Go-http-client/1.1",
    "X-Amzn-Trace-Id": "Root=1-62f13e78-070d2ddd6f311c14737f0d37"
  },
  "json": null,
  "origin": "XXXXXX",
  "url": "http://httpbin.org/post"
}
*/
```

### HTTP Headers

More header utilities may be added in the future. Follow [#23](https://github.com/asg017/sqlite-http/issues/23) for more info.

<h4 name="http_headers"> <code>http_headers(name1, value1, ...)</code></h4>

Generates the given names and values into HTTP headers, in "wire" format. Useful in `http_get` and other related request functions.

```sql
select http_headers(
  'User-Agent', 'PostmanRuntime/1.0.0',
  'X-API-Key', 'abcdef12345'
);
/*
/User-Agent: PostmanRuntime/1.0.0
X-Api-Key: abcdef12345
*/

select http_get_body(
  'http://httpbin.org/headers',
  http_headers(
    'X-Foo', 'bar',
    'X-Name', 'Alex'
  )
);
/*
{
  "headers": {
    "Accept-Encoding": "gzip",
    "Host": "httpbin.org",
    "User-Agent": "Go-http-client/1.1",
    "X-Amzn-Trace-Id": "Root=1-62f13fe2-103461a96946d7542c5412d3",
    "X-Foo": "bar",
    "X-Name": "Alex"
  }
}
*/
```

Note that the default `User-Agent`, if not overridden, is `"Go-http-client/1.1"`. This may change in the future.

<h4 name="http_headers_has"> <code>http_headers_has(headers, name)</code></h4>

Returns `1` if the given `name` exists as a name inside the given `headers`, `0` otherwise.

Keep in mind, HTTP headers are case-insensitive, and headers with the same name can exists multiple times.

```sql
select http_headers_has(
  http_headers('X-Foo', 'bar'),
  'x-foo'
);
-- 1

select http_headers_has(
  http_headers('X-Foo', 'bar'),
  'x-nope'
);
-- 0
/*
*/
```

<h4 name="http_headers_get"> <code>http_headers_get(headers, name)</code></h4>

Returns the string value of the first header in `headers` found with the given `name` as a name.

Keep in mind, HTTP headers are case-insensitive, and headers with the same name can exists multiple times. Use [`http_headers_each`](#http_headers_each) if you want to get all values with a given name.

```sql
select http_();
/*
*/
```

<h4 name="http_headers_each"> <code>http_headers_each(headers)</code></h4>

A table function that iterates through all header entries in the given `headers`, in wire format.

Note that `http_headers_each` may change the casing of a given header's name. Since HTTP header names are case-insensitive, it shouldn't change much.

```sql
CREATE TABLE http_headers_each(
  name TEXT, -- Name of the current HTTP header
  value TEXT -- Value of the current HTTP header
);
```

```sql
select
  name,
  value
from http_headers_each(
  http_get_headers('https://api.census.gov/data/')
);

/*
┌──────────────────────────────┬──────────────────────────────────────────────────────────────┐
│             name             │                            value                             │
├──────────────────────────────┼──────────────────────────────────────────────────────────────┤
│ Strict-Transport-Security    │ max-age=31536000                                             │
├──────────────────────────────┼──────────────────────────────────────────────────────────────┤
│ Access-Control-Allow-Headers │ Origin, X-Requested-With, Content-Type, Accept               │
├──────────────────────────────┼──────────────────────────────────────────────────────────────┤
│ Access-Control-Allow-Methods │ GET,POST,OPTIONS                                             │
├──────────────────────────────┼──────────────────────────────────────────────────────────────┤
│ Access-Control-Allow-Origin  │ *                                                            │
├──────────────────────────────┼──────────────────────────────────────────────────────────────┤
│ Cache-Control                │ private                                                      │
├──────────────────────────────┼──────────────────────────────────────────────────────────────┤
│ Content-Type                 │ application/json;charset=utf-8                               │
├──────────────────────────────┼──────────────────────────────────────────────────────────────┤
│ Date                         │ Wed, 03 Aug 2022 21:41:30 GMT                                │
├──────────────────────────────┼──────────────────────────────────────────────────────────────┤
│ Set-Cookie                   │ TS010383f0=01283c52a4a8f8b52f957f1f4dc0e76601beccd5188de846a │
│                              │ 0166ce7672edcf03f9c37e1f99f02bf385f0e39381d28a44b7da397e0; P │
│                              │ ath=/; Domain=.api.census.gov                                │
└──────────────────────────────┴──────────────────────────────────────────────────────────────┘
*/
```

<h4 name="http_headers_date"> <code>http_headers_date(value)</code></h4>

Parses timestamps in RFC5322 format (like `"Mon, 08 Aug 2022 17:12:47 GMT"`) into ISO8601, the format that other SQLite date functions use. Useful for most `Date`, `Last-Modified`, and other timestamp related headers.

Keep in mind, some servers may return [an obselete format](https://httpwg.org/specs/rfc7231.html#origination.date) (like RFC850 or asctime()), which this library does not support.

```sql
select http_headers_date('Sun, 06 Nov 1994 08:49:37 GMT');
-- '1994-11-06 08:49:37'

select http_headers_date(
  http_headers_get(
    http_get_headers('https://www.irs.gov/pub/irs-pdf/p1.pdf'),
    'Expires'
  )
);
-- '2022-08-09 17:19:46'
/*
*/
```

### HTTP Cookies

More cookie utilities may be added in the future. Follow [#24](https://github.com/asg017/sqlite-http/issues/24) for more info.

#### `http_cookies()`

### Configuring `sqlite-http` Behavior

Change the timeout and rate-limit settings for all HTTP requests made by `sqlite-http`, in the given connection. Settings don't persist after a connection is closed.

<h4 name="http_rate_limit"> <code>http_rate_limit(duration_ms)</code></h4>

Wait `duration_ms` milliseconds between all `sqlite-http` requests. This is helpful when asserting a "rate limit", to ensure you don't flood a site with requests.

Note that _all_ HTTP requests will be effected, including `http_do`, `http_do_body`, and `http_do_header` related functions.

```sql
select http_rate_limit(100);

select http_get_body('http://localhost:8080');

-- If the above request happens in < 100ms, then this will block and wait
-- Until the 100ms has passed since the start of the above request.
select http_get_body('http://localhost:8080');
```

<h4 name="http_timeout_set"> <code>http_timeout_set(duration_ms)</code></h4>

Set the timeout value for all HTTP requests to `duration_ms` milliseconds. Defaults to 5 seconds.

Note that _all_ HTTP requests will be effected, including `http_do`, `http_do_body`, and `http_do_header` related functions.

```sql
select http_timeout_set(2500); -- 2500

-- blocks for 2 seconds, succeeds
select http_get_body('http://httpbin.org/delay/2');

select http_timeout_set(500); -- 500

-- blocks for .5 seconds, fails!
select http_get_body('http://httpbin.org/delay/2');
-- "Runtime error: Get "http://httpbin.org/delay/2": context deadline exceeded (Client.Timeout exceeded while awaiting headers)"
```

### `sqlite-http` Information

<h4 name="http_version"> <code>http_version()</code></h4>

Returns the version string of the `sqlite-http` library, modeled after [`sqlite_version()`](https://www.sqlite.org/lang_corefunc.html#sqlite_version).

```sql
select http_version();
-- "v0.0.0"
```

<h4 name="http_debug"> <code>http_debug()</code></h4>

Returns debug information of the `sqlite-http` library, including the version string. Subject to change.

```sql
select http_debug();
/*
Version: "v0.0.0"
Commit: 85445c7b9d539e2626731275da6496d96d6dbb05
Runtime: go1.17 darwin/amd64
Date: 2021-11-17T16:20:06Z-0800
*/
```
