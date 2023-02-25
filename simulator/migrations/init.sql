-- Create tables and indexes.

CREATE TABLE IF NOT EXISTS trips
(
  trip_duration           INTEGER,
  start_time              TIMESTAMP,
  stop_time               TIMESTAMP,
  start_station_id        INTEGER,
  start_station_name      VARCHAR(16),
  start_station_latitude  REAL,
  start_station_longitude REAL,
  end_station_id          INTEGER,
  end_station_name        VARCHAR(256),
  end_station_latitude    REAL,
  end_station_longitude   REAL,
  bike_id                 INTEGER,
  membership_type         VARCHAR(64),
  usertype                VARCHAR(64),
  birth_year              INTEGER,
  gender                  INTEGER
);

CREATE INDEX IF NOT EXISTS idx_bike_id ON trips (bike_id);
CREATE INDEX IF NOT EXISTS idx_start_station_id ON trips (lower(start_station_id));
CREATE INDEX IF NOT EXISTS idx_end_station_id ON trips (lower(end_station_id));
