.load dist/http.so
--.load /Users/alex/projects/sqlite-stdlib/fetchx/fetch.so

.mode csv
.headers on
.bail on


create temporary table testcases(
  category, 
  description, 
  result integer,
  status GENERATED ALWAYS AS ( iif(result, '✅', '❌') ) VIRTUAL
);

create TEMPORARY view get as select *
    from http_get("http://localhost:8080/get", 
      http_headers(
        "a", "b",
        "label", "o no"
      )
    )
  ;

insert into testcases(category, description, result)
  --with 
  select 'http_get',
    'timings',
    json_valid(get.timings)
  from get
  union all select 'http_get',
    'request_url',
    get.request_url == 'http://localhost:8080/get'
  from get
  union all select 'http_get',
    'request_method',
    get.request_method == 'GET'
  from get
  union all select 'http_get',
    'request_headers a==b',
    http_headers_get(get.request_headers, 'A') == 'b'
  from get
  -- seems like Headers.Write doesn't include user-agent... guess it happens higher up
  union all select 'http_get',
    'request_headers has default user-agent',
    http_headers_get(get.request_headers, 'user-agent') == 'Go-http-client/1.1'
  from get
  union all select 'http_get',
    'response_status',
    get.response_status == "200 OK"
  from get
  union all select 'http_get',
    'response_status_code 200s',
    get.response_status_code == 200
  from get
  union all select 'http_get',
    'response_headers is json',
    http_headers_get(get.response_headers, "Content-Type") == 'application/json'
  from get
  union all select 'http_get',
    'response_body is json',
    json_valid(get.response_body)
  from get
  union all select 'http_get',
    'response_body app returned header A',
    json_extract(get.response_body, '$.headers.A') == 'b'
  from get
  union all select 'http_get',
    'response_body app returned default user-agent',
    json_extract(get.response_body, '$.headers.User-Agent') == 'Go-http-client/1.1'
  from get
  ;


insert into testcases(category, description, result)
  with cookies as (select *
    from http_get("http://localhost:8080/cookies", 
      http_headers(),
      http_cookies("donald", "duck")
    )
  )
  select 'http_get',
    'response_body app returned cookie "donald"',
    json_extract(cookies.response_body, '$.cookies.donald') == 'duck'
  from cookies;

select * from testcases;

--select json_extract(get.response_body, '$.headers') from get;

--select request_headers from get;
.exit 1