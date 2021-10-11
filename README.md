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

### POST/DO body utilities
  - multipart forms
  - other types of POSTs idk
  - `http_post_multiform(name1, file1, ...)`

### More cookie utility functions
  - `http_cookie_name(cookie)`, `http_cookie_expires(cookie)`

### URL utility functions
  - query parameters, `url_query_parameters`
  - host, path
  - scheme
  - username/password

### HAR (HTTP Archive) compatibility
  - Export table to HAR?
  - `http_to_har('select * from archives')`
  - `http_to_har('select * from http_get(...)')`
  - `http_get_har(...)`
  - `http_from_har(har)` (table-valued functions)
- [ ] Rate limiters?