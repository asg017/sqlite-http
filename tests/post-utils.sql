.load dist/http0.so
.load tests/deps/assert0

.bail on 

with urlencoded as (
  select http_post_body(
    'http://localhost:8080/post', 
    http_headers('Content-Type', 'application/x-www-form-urlencoded'),
    http_post_form_url_encoded(
      'name', 'Alex',
      'age', 32
    )
  ) as value
)
select 
  assert(json_valid(urlencoded.value), 'httpbin returned json for post'),
  assert(json_extract(urlencoded.value, '$.form.name') == 'Alex', 'form includes name'),
  assert(json_extract(urlencoded.value, '$.form.age') == '32', 'form includes age')
from urlencoded;