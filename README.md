# sqlite-http

A SQLite extension for making HTTP requests purely in SQL.

## 🚧🚧 Work In Progress! 🚧🚧

This library is extremely experimental and subject to change. I plan to make a stable beta release and subsequent v0+v1 in the near future, so use with caution.

When v0 is ready (with a mostly stable API), I will make a release (so watch this repo for that) and will make a blog post, feel free to [follow me on twitter](https://twitter.com/agarcia_me) to get notified of that.

## Overview

Scalar functions:

- Request the body contents from a URL
  - [http_get_body](#)(_url, headers, cookies_)
    cookies\_)
  - [http_post_body](#)(_url, headers, body, cookies_)
  - [http_do_body](#)(_method, url, headers, body, cookies_)
- Request the header contents from a URL
- [http_get_headers](#)(_url, headers, - [http_post_headers](#)(\_url, headers, body, cookies_)
- [http_do_headers](#)(_method, url, headers, body, cookies_)
- [http_headers](#)(_label1, value1_)
- Create, query, and manipulate HTTP headers in wire format
  - [http_headers_has](#)(_headers, key_)
  - [http_headers_get](#)(_headers, get_)
  - [http_headers_all](#)(_headers, key_)
- Create, query, and manipulate HTTP queries
  - [http_cookies](#)(_label1, value1_)

Table-valued functions:

- [http_get](#)(_url, headers, cookies_)
- [http_post](#)(_url, headers, body, cookies_)
- [http_do](#)(_method, url, headers, body, cookies_)

- [http_rate_limit](#)(_duration_)
- [http_timeout](#)(_duration_)

```SQL

-- Also http_get and http_get
CREATE TABLE http_do(

);
```

## Interface Overview

- making requests
- HEADERS arguments
- COOKIES arguments

## Function Details

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
