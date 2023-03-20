select
  bike_id,
  birth_year,
  end_station_id,
  end_station_latitude,
  end_station_longitude,
  end_station_name,
  gender,
  membership_type,
  start_station_id,
  start_station_latitude,
  start_station_longitude,
  start_station_name,
  start_time,
  stop_time,
  trip_duration,
  usertype
from cdp_dev.raw.trips
