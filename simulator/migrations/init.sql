-- Trips

CREATE TABLE IF NOT EXISTS trips
(
  trip_duration           VARCHAR(64)  NULL,
  start_time              VARCHAR(32)  NULL,
  stop_time               VARCHAR(32)  NULL,
  start_station_id        VARCHAR(16)  NULL,
  start_station_name      VARCHAR(256) NULL,
  start_station_latitude  VARCHAR(16)  NULL,
  start_station_longitude VARCHAR(16)  NULL,
  end_station_id          VARCHAR(16)  NULL,
  end_station_name        VARCHAR(256) NULL,
  end_station_latitude    VARCHAR(16)  NULL,
  end_station_longitude   VARCHAR(16)  NULL,
  bike_id                 VARCHAR(32)  NULL,
  membership_type         VARCHAR(32)  NULL,
  usertype                VARCHAR(32)  NULL,
  birth_year              VARCHAR(16)  NULL,
  gender                  VARCHAR(16)  NULL
);

-- Riders

-- We capture and persist the data as json in Postgres.

CREATE TABLE IF NOT EXISTS riders
(
  json TEXT
);
