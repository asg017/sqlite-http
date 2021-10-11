.load ../../dist/http.so

create table if not exists snapshots as select * from http_get() limit 0;

insert into snapshots
  select * from http_get('https://oag.ca.gov/privacy/databreach/list');
