select
    usertype,
    count(bike_id) as count
from
    {{ source('citibike', 'trips') }}
group by
    usertype
