.bail on
.mode box
.header on
.load dist/http0

create table cities as 
  select
      value ->> '$.city'       as name,
      value ->> '$.latitude'   as latitude,
      value ->> '$.longitude'  as longitude,
      value ->> '$.population' as population,
      value ->> '$.state'      as state
  from json_each(
    http_get_body('https://gist.githubusercontent.com/Miserlou/c5cd8364bf9b2420bb29/raw/2bf258763cdddd704f8ffd3ea9a3e81d25e2c6f6/cities.json')
  );

create table christmas_suntimes as 
  with suntimes as (
    select rowid as city,
    http_post_body(
      'http://localhost:3001/suncalc', 
      null, 
      json_object(
        'longitude', longitude,
        'latitude', latitude,
        -- ho ho ho 
        'date', '2022-12-25'
      )
    ) as suncalc_times
    from cities
  ),
  final as (
    select
      cities.rowid as city,
      suncalc_times ->> '$.nadir'          as nadir,
      suncalc_times ->> '$.nightEnd'       as night_end,
      suncalc_times ->> '$.nauticalDawn'   as nautical_dawn,
      suncalc_times ->> '$.dawn'           as dawn,
      suncalc_times ->> '$.sunrise'        as sunrise,
      suncalc_times ->> '$.sunriseEnd'     as sunrise_end,
      suncalc_times ->> '$.goldenHourEnd'  as golden_hour_end,
      suncalc_times ->> '$.solarNoon'      as solar_noon,
      suncalc_times ->> '$.goldenHour'     as golden_hour,
      suncalc_times ->> '$.sunsetStart'    as sunset_start,
      suncalc_times ->> '$.sunset'         as sunset,
      suncalc_times ->> '$.dusk'           as dusk,
      suncalc_times ->> '$.nauticalDusk'   as nautical_dusk,
      suncalc_times ->> '$.night'          as night
    from suntimes
    left join cities on cities.rowid = suntimes.city
  )
  select * from final;

select 
  cities.name,
  cities.state,
  (unixepoch(sunset) - unixepoch(sunrise)) as daylight_seconds
from christmas_suntimes
left join cities on cities.rowid = christmas_suntimes.city
order by 3 desc
limit 10;

select 
  cities.name,
  cities.state,
  (unixepoch(sunset) - unixepoch(sunrise)) as daylight_seconds
from christmas_suntimes
left join cities on cities.rowid = christmas_suntimes.city
order by 3 asc
limit 10;