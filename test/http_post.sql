.load dist/http.so

.mode csv
.headers on
.bail on


create temporary table testcases(
  category, 
  description, 
  result integer,
  status GENERATED ALWAYS AS ( iif(result, '✅', '❌') ) VIRTUAL);

create TEMPORARY view post_json as select *
    from http_post("http://localhost:8080/post", 
      http_headers("content-type", "application/json"),
      json_object("name", "Alex")
    );

/*TODO
- other types of posts  
- 
*/

insert into testcases(category, description, result)
  select 'json',
    'sent json body',
    cast(request_body as text) --== '{"name":"Alex"}'
  from post_json
  union all
  select 'json',
    'app received application/json header',
    json_extract(response_body, '$.headers.Content-Type') == 'application/json'
  from post_json
  union all
  select 'json',
    'app got json body name=Alex',
    json_extract(response_body, '$.json.name') == 'Alex'
  from post_json
  ;

select * from testcases;

.exit 1;