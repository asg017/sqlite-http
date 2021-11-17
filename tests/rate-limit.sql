.load dist/http0.so
.load tests/deps/assert0

.bail on
.headers on

select http_rate_limit(200);

with should_100 as (
  select count.value,
    json_extract(get.timings, '$.start') as start
  from generate_series(1,1000) as count,
    http_get(
      printf('http://localhost:8080/get?i=%d', count.value)
    ) as get
),
 testcases as (
  select value,
    start as curr,
    lag(start) over (order by value) as prev
  from should_100
)
select 
  assert(
      (
        (strftime('%f', curr) - strftime('%f', prev)) * 1000 % 1000
      ) 
      between 190 and 220, 
    'requests start 200ms after previous', curr, prev, (strftime('%f', curr) - strftime('%f', prev)) * 1000 % 1000
  )
from testcases
where prev is not null;

-- (strftime('%f', '2021-10-30 16:30:06.026') - strftime('%f', '2021-10-30 16:30:05.829')) * 1000 % 1000
.exit 0

select http_rate_limit(250);

select count.value,
  json_extract(get.timings, '$.start')
from generate_series(1,10) as count,
  http_get(
    printf('http://localhost:8080/get?i=%d', count.value)
  ) as get;


select http_rate_limit(1000);

select count.value,
  json_extract(get.timings, '$.start')
from generate_series(1,3) as count,
  http_get(
    printf('http://localhost:8080/get?i=%d', count.value)
  ) as get;