.load dist/http0.so
.load tests/deps/assert0

select 
  assert(
    json_extract(http_cookies('name', 'alex'), '$.name') == "alex",
    "http_cookies returns json"
  );