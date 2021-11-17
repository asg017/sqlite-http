.bail on
.load dist/http0.so

create table if not exists snapshots as 
  select 0 as agency, * from http_get limit 0;

create table if not exists agencies(
  id integer,
  website,
  name,
  acronym,
  value
);

insert into agencies(id, website, name, acronym, value)
  with api_agencies as (
    select http_post_body(
      'https://api.stateentityprofile.ca.gov/api/Agencies/Get?page=0&pageSize=0&lang=en'
    ) as response
  )
  select 
    json_extract(items.value, '$.AgencyId') as id,
    json_extract(items.value, '$.WebsiteURL') as website,
    json_extract(items.value, '$.AgencyName') as name,
    json_extract(items.value, '$.Acronym') as acronym,
    items.value as value
  from api_agencies, json_each(api_agencies.response, '$.Data') as items;

select http_rate_limit(500);

select printf("%d agencies", count(*)) from agencies;

insert into snapshots
  select agencies.rowid, requests.* 
  from agencies, http_get(agencies.website) as requests;