with data as (
  select parse_json(val) as json from {{ ref('riders') }}
)

select
  j.value:bikeid::text as bike_id,
  j.value:bike_type::text as bike_type,
  j.value:end_station_id::int as end_station_id

from data as d, lateral flatten(input => d.json) j
