.load dist/http.so
.bail on
.mode csv
.headers on

create table commits as select 1 as i, * from http_get() limit 0;

select printf("Inserting...");

insert into commits
  with recursive pages as (
    select 1 as i, * from http_get('https://api.github.com/repos/asg017/dataflow/commits?page=1')
    union all
    select pages.i+1 as i, req.* 
    from http_get( printf('https://api.github.com/repos/asg017/dataflow/commits?page=%d', pages.i + 1)) as req, pages
    where json_array_length(req.response_body) > 1
  )
  select * from pages;