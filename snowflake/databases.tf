locals {
  schemas = {
    "RAW" = {
      comment      = "Dump of raw data extracted from data sources"
      usage_roles  = ["TRANSFORMER"]
      modify_roles = ["LOADER"]
    }
    "ANALYTICS" = {
      comment      = "Tables and views for analytics and reporting"
      usage_roles  = []
      modify_roles = ["TRANSFORMER"]
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

  column {
    name     = "trip_duration"
    type     = "text"
    nullable = true
  }

  column {
    name     = "start_time"
    type     = "text"
    nullable = true
  }

  column {
    name     = "stop_time"
    type     = "text"
    nullable = true
  }
}
