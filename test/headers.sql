.load dist/http.so

.bail on 
.headers on
.mode csv

.param init

insert into sqlite_parameters values(':headers_dups',  http_headers(
  "user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36",
  "dup", "a",
  "dup", "b",
  "dup", "c"
));

select printf("RESULTS");

create temporary table testcases(
  category, 
  description, 
  result integer,
  status GENERATED ALWAYS AS ( iif(result, '✅', '❌') ) VIRTUAL
  check (result == 1)
);

-- headers_has
insert into testcases(category, description, result)
  select 'headers_has' as category,
    printf('lookup is case-insensitive (%s)', tests.value) as description,
    http_headers_has(
      :headers_dups, 
      tests.value) as result
  from json_each('["user-agent", "USER-AGENT", "USer-agENt"]') as tests
  union all
  select 'headers_has' as category,
    printf('lookup fails on not-exists (%s)', tests.value) as description,
    not http_headers_has(
      :headers_dups, 
      tests.value) as result
  from json_each('["DNE", "not here", ""]') as tests;

-- headers_get

insert into testcases(category, description, result)
  values
    ( 
      'headers_get', 
      'lookup returns first',
      http_headers_get(:headers_dups, 'dup') == "a"
    ),
    ( 
      'headers_all', 
      'lookup returns null if not exist',
      http_headers_all(:headers_dups, 'dup') == '["a","b","c"]'
    );


select * from testcases;

.exit 1