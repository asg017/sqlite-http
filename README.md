# sqlite-http

## Overview

Scalar functions:

- [http_get_body](#)(_url, headers, cookies_)
- [http_get_headers](#)(_url, headers, cookies_)
- [http_post_body](#)(_url, headers, body, cookies_)
- [http_post_headers](#)(_url, headers, body, cookies_)
- [http_do_body](#)(_method, url, headers, body, cookies_)
- [http_do_headers](#)(_method, url, headers, body, cookies_)
- [http_headers](#)(_label1, value1_)
- [http_headers_has](#)(_headers, key_)
- [http_headers_get](#)(_headers, get_)
- [http_headers_all](#)(_headers, key_)
- [http_cookies](#)(_label1, value1_)

Table-valued functions:

- [http_get](#)(_url, headers, cookies_)
- [http_post](#)(_url, headers, body, cookies_)
- [http_do](#)(_method, url, headers, body, cookies_)

```SQL

-- Also http_get and http_get
CREATE TABLE http_do(

);
```

## Interface Overview

- making requests
- HEADERS arguments
- COOKIES arguments

## Function Detauls

### The `http_get_body()` and `http_get_headers()` functions

X

Examples:

- `http_()` ➡ `''`

### The `http_post_body()` and `http_post_headers()` functions

X

Examples:

- `http_()` ➡ `''`

### The `http_do_body()` and `http_do_headers()` functions

X

Examples:

- `http_()` ➡ `''`

### The `http_headers()` functions

X

Examples:

- `http_()` ➡ `''`

#### The `http_headers_has()` function

X

Examples:

- `http_()` ➡ `''`

#### The `http_headers_get()` function

X

Examples:

- `http_()` ➡ `''`

#### The `http_headers_all()` function

X

Examples:

- `http_()` ➡ `''`

### The `http_cookies()` function

X

Examples:

- `http_()` ➡ `''`

## TODO

### Immediate

- [ ] Code cleanup
  - http_do -> get.go, post.go, do.go
  - New internal package?
  - cookies package?
  - url package?
- [ ] TVF for headers, ex `select name, value from http_headers_each('...')`
  - how handle case-insensitive lookup?
- [ ] Timings
  - move to end
  - add marker for body start, body end? might have to add to cursor
  - document caveats, e.g. sqlite .Next()/.Column, need to access fields, etc.
- [ ] `select http_settings('fail_on', )`
  - if request fails, what do
  - ex ssl error, site not exist, timeout
  - for now, lets do udf for changing failure behavior and other settings
  - `select http_timeout_set(5000)`

### POST/DO body utilities

- multipart forms
- other types of POSTs idk
- `http_post_multiform(name1, file1, ...)`
- `application/x-www-form-urlencoded`

### More cookie utility functions

- `http_cookie_name(cookie)`, `http_cookie_expires(cookie)`

### URL utility functions

- query parameters, `url_query_parameters`
- host, path
- scheme
- username/password

```
http_url("https", "acs.ca.gov", "/path/to/file",
  http_url_query_parameters(
    "q", "books",
    _format", "json",
    "timeout", 1500
  )
)
```

```
http_url_extract_domain('https://www.dir.ca.gov/cac/cac.html') -- 'www.dir.ca.gov'
http_url_extract_tld('https://www.dir.ca.gov/cac/cac.html') -- 'ca.gov'
http_url_extract_path('https://www.dir.ca.gov/cac/cac.html') -- '/cac/cac.html'
http_url_extract_fragment('') -- ''
http_url_query_params('') -- ''


```

### HAR (HTTP Archive) compatibility

- Export do table to HAR
  - `http_to_har('select * from archives')`
  - `http_to_har('select * from http_get(...)')`
- Import from har
  - `select * from http_har(readfile('file.har'), 'creator_name')`
- Create HAR
- `http_get_har(...), http_post_har(...), http_do_har(...)`

### Rate Limiter?

```
select http_rate_limiter_set(100);
```
