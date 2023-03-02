locals {
  schemas = {
    RAW = {
      comment      = "Dump of raw data extracted from data sources"
      usage_roles  = ["TRANSFORMER"]
      modify_roles = ["LOADER"]

      tables = {
        trips = {
          columns = {
            trip_duration = {
              type     = "text"
              nullable = true
            }
            start_time = {
              type     = "text"
              nullable = true
            }
            stop_time = {
              type     = "text"
              nullable = true
            }
            start_station_id = {
              type     = "text"
              nullable = true
            }
            start_station_name = {
              type     = "text"
              nullable = true
            }
            start_station_latitude = {
              type     = "text"
              nullable = true
            }
            start_station_longitude = {
              type     = "text"
              nullable = true
            }
            end_station_id = {
              type     = "text"
              nullable = true
            }
            end_station_name = {
              type     = "text"
              nullable = true
            }
            end_station_latitude = {
              type     = "text"
              nullable = true
            }
            end_station_longitude = {
              type     = "text"
              nullable = true
            }
            bike_id = {
              type     = "text"
              nullable = true
            }
            membership_type = {
              type     = "text"
              nullable = true
            }
            usertype = {
              type     = "text"
              nullable = true
            }
            birth_year = {
              type     = "text"
              nullable = true
            }
            gender = {
              type     = "text"
              nullable = true
            }
          }
        }
      }
    }
    ANALYTICS = {
      comment      = "Tables and views for analytics and reporting"
      usage_roles  = []
      modify_roles = ["TRANSFORMER"]

      tables = {
        trips = {
          columns = {
            trip_duration = {
              type     = "integer"
              nullable = true
            }
            start_time = {
              type     = "timestamp_ntz(9)"
              nullable = true
            }
            stop_time = {
              type     = "timestamp_ntz(9)"
              nullable = true
            }
            start_station_id = {
              type     = "integer"
              nullable = true
            }
            start_station_name = {
              type     = "text"
              nullable = true
            }
            start_station_latitude = {
              type     = "float"
              nullable = true
            }
            start_station_longitude = {
              type     = "float"
              nullable = true
            }
            end_station_id = {
              type     = "integer"
              nullable = true
            }
            end_station_name = {
              type     = "text"
              nullable = true
            }
            end_station_latitude = {
              type     = "float"
              nullable = true
            }
            end_station_longitude = {
              type     = "float"
              nullable = true
            }
            bike_id = {
              type     = "integer"
              nullable = true
            }
            membership_type = {
              type     = "text"
              nullable = true
            }
            usertype = {
              type     = "text"
              nullable = true
            }
            birth_year = {
              type     = "integer"
              nullable = true
            }
            gender = {
              type     = "integer"
              nullable = true
            }
          }
        }
      }
    }
  }
}

resource "snowflake_database" "cdp" {
  provider = snowflake.sys_admin

  name = "CDP-DEV"
}

resource "snowflake_schema" "schemas" {
  provider = snowflake.sys_admin
  for_each = local.schemas

  name     = each.key
  database = snowflake_database.cdp.name
  comment  = each.value.comment
}

resource "snowflake_table" "trips" {
  provider = snowflake.sys_admin
  for_each = local.schemas

  database = snowflake_database.cdp.name
  schema   = each.key
  name     = "trips"

  dynamic "column" {
    for_each = each.value.tables.trips.columns
    content {
      name     = column.key
      type     = lookup(column.value, "type", "text")
      nullable = lookup(column.value, "nullable", true)
    }
  }
}
