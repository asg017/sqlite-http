.load dist/http0
.mode box
.timer on
.bail on

create table if not exists snapshots as
  select * from http_get() limit 0;

insert into snapshots
  select *
  from http_get('http://www.ridembl.com:8080/MBLBusTracker/BusTracker/getRouteDetailByRouteID?RouteId=10');


create table vehicle_readings as
  select 
    snapshots.rowid as snapshot,
    snapshots.timings ->> '$.body_end' as response_time,
    vehicles.value ->> '$.BlockFareboxId'              as block_farebox_id,
    vehicles.value ->> '$.CommStatus'                  as comm_status,
    vehicles.value ->> '$.Destination'                 as destination,
    vehicles.value ->> '$.Deviation'                   as deviation,
    vehicles.value ->> '$.Direction'                   as direction,
    vehicles.value ->> '$.DirectionLong'               as direction_long,
    vehicles.value ->> '$.DisplayStatus'               as display_status,
    vehicles.value ->> '$.StopId'                      as stop_id,
    vehicles.value ->> '$.CurrentStatus'               as current_status,
    vehicles.value ->> '$.DriverName'                  as driver_name,
    vehicles.value ->> '$.GPSStatus'                   as gps_status,
    vehicles.value ->> '$.Heading'                     as heading,
    vehicles.value ->> '$.LastStop'                    as last_stop,
    vehicles.value ->> '$.LastUpdated'                 as last_updated,
    vehicles.value ->> '$.Latitude'                    as latitude,
    vehicles.value ->> '$.Longitude'                   as longitude,
    vehicles.value ->> '$.Name'                        as name,
    vehicles.value ->> '$.OccupancyStatus'             as occupancy_status,
    vehicles.value ->> '$.OnBoard'                     as on_board,
    vehicles.value ->> '$.OpStatus'                    as op_status,
    vehicles.value ->> '$.RouteId'                     as route_id,
    vehicles.value ->> '$.RunId'                       as run_id,
    vehicles.value ->> '$.Speed'                       as speed,
    vehicles.value ->> '$.TripId'                      as trip_id,
    vehicles.value ->> '$.VehicleId'                   as vehicle_id,
    vehicles.value ->> '$.SeatingCapacity'             as seating_capacity,
    vehicles.value ->> '$.TotalCapacity'               as total_capacity,
    vehicles.value ->> '$.PropertyName'                as property_name,
    vehicles.value ->> '$.OccupancyStatusReportLabel'  as occupancy_status_report_label
  from snapshots
  join json_each(snapshots.response_body, '$.Vehicles') as vehicles
