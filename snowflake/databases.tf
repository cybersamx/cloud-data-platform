locals {
  schemas = {
    RAW = {
      name                = "RAW"
      comment             = "Dump of raw data extracted from data sources"
      schema_usage_roles  = ["LOADER", "TRANSFORMER"]
      schema_modify_roles = ["LOADER"]

      tables = {
        TRIPS = {
          name = "TRIPS"
          privileges = {
            SELECT = {
              name  = "SELECT"
              roles = ["LOADER", "TRANSFORMER"]
            }
            INSERT = {
              name  = "INSERT"
              roles = ["TRANSFORMER"]
            }
          }
          columns = {
            trip_duration = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            start_time = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            stop_time = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            start_station_id = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            start_station_name = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            start_station_latitude = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            start_station_longitude = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            end_station_id = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            end_station_name = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            end_station_latitude = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            end_station_longitude = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            bike_id = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            membership_type = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            usertype = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            birth_year = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            gender = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
          }
        }
      }
    }
    ANALYTICS = {
      name                = "ANALYTICS"
      comment             = "Tables and views for analytics and reporting"
      schema_usage_roles  = []
      schema_modify_roles = ["TRANSFORMER"]

      tables = {
        "TRIPS" = {
          name = "TRIPS"
          privileges = {
            SELECT = {
              name  = "SELECT"
              roles = ["TRANSFORMER"]
            }
            INSERT = {
              name  = "INSERT"
              roles = ["TRANSFORMER"]
            }
          }
          columns = {
            trip_duration = {
              // For types use capitalized, basic types like NUMBER(38,0) than lowercase, alias types like
              // integer or text. While NUMBER(38,0) is semantically equivalent to number(38,0), integer or INTEGER,
              // the current Snowflake provider does an exact case sensitive match of the the string between the
              // value in the terraform code and state. As a result, INTEGER or number(38,0) will be marked as change.
              type     = "NUMBER(38,0)"
              nullable = true
            }
            start_time = {
              type     = "TIMESTAMP_NTZ(9)"
              nullable = true
            }
            stop_time = {
              type     = "TIMESTAMP_NTZ(9)"
              nullable = true
            }
            start_station_id = {
              type     = "NUMBER(38,0)"
              nullable = true
            }
            start_station_name = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            start_station_latitude = {
              type     = "FLOAT"
              nullable = true
            }
            start_station_longitude = {
              type     = "FLOAT"
              nullable = true
            }
            end_station_id = {
              type     = "NUMBER(38,0)"
              nullable = true
            }
            end_station_name = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            end_station_latitude = {
              type     = "FLOAT"
              nullable = true
            }
            end_station_longitude = {
              type     = "FLOAT"
              nullable = true
            }
            bike_id = {
              type     = "NUMBER(38,0)"
              nullable = true
            }
            membership_type = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            usertype = {
              type     = "VARCHAR(16777216)"
              nullable = true
            }
            birth_year = {
              type     = "NUMBER(38,0)"
              nullable = true
            }
            gender = {
              type     = "NUMBER(38,0)"
              nullable = true
            }
          }
        }
      }
    }
  }
}

locals {
  # Flatten the nested data structures.
  table_privileges = distinct(flatten([
    for schema in local.schemas : [
      for table in schema.tables : [
        for privilege in table.privileges : {
          schema    = schema.name
          table     = table.name
          privilege = privilege.name
          roles     = privilege.roles
        }
      ]
    ]
  ]))
}

resource "snowflake_database" "cdp" {
  provider = snowflake.sys_admin

  name = "CDP_DEV"
}

resource "snowflake_database_grant" "usage_roles" {
  provider = snowflake.security_admin
  # Explicit depends_on because we are using string to reference schemas and role.
  depends_on = [snowflake_schema.schemas, snowflake_role.roles]
  for_each   = local.schemas

  database_name     = snowflake_database.cdp.name
  privilege         = "USAGE"
  roles             = each.value.schema_usage_roles
  with_grant_option = false
}

resource "snowflake_database_grant" "modify_roles" {
  provider = snowflake.security_admin
  # Explicit depends_on because we are using string to reference schemas and role.
  depends_on = [snowflake_schema.schemas, snowflake_role.roles]
  for_each   = local.schemas

  database_name     = snowflake_database.cdp.name
  privilege         = "MODIFY"
  roles             = each.value.schema_modify_roles
  with_grant_option = false
}

resource "snowflake_schema_grant" "usage_roles" {
  provider = snowflake.security_admin
  # Explicit depends_on because we are using string to reference schemas and role.
  depends_on = [snowflake_schema.schemas, snowflake_role.roles]
  for_each   = local.schemas

  database_name     = snowflake_database.cdp.name
  schema_name       = each.key
  privilege         = "USAGE"
  roles             = each.value.schema_usage_roles
  with_grant_option = false
}

resource "snowflake_schema_grant" "modify_roles" {
  provider = snowflake.security_admin
  # Explicit depends_on because we are using string to reference schemas and role.
  depends_on = [snowflake_schema.schemas, snowflake_role.roles]
  for_each   = local.schemas

  database_name     = snowflake_database.cdp.name
  schema_name       = each.key
  privilege         = "MODIFY"
  roles             = each.value.schema_modify_roles
  with_grant_option = false
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
  name     = "TRIPS" # TODO: Don't hardcode, get the value from locals.

  dynamic "column" {
    for_each = each.value.tables.TRIPS.columns
    content {
      name     = column.key
      type     = lookup(column.value, "type", "VARCHAR(16777216)")
      nullable = lookup(column.value, "nullable", true)
    }
  }
}

resource "snowflake_table_grant" "trips" {
  provider = snowflake.sys_admin
  # Explicit depends_on because we are using string to reference schemas and role.
  depends_on = [snowflake_schema.schemas, snowflake_role.roles, snowflake_table.trips]
  # for_each only accept map or list of string, so we need to convert the list to a map.
  for_each = { for item in local.table_privileges : "${item.schema}.${item.table}.${item.privilege}" => item }

  database_name     = snowflake_database.cdp.name
  schema_name       = each.value.schema
  table_name        = each.value.table
  privilege         = each.value.privilege
  roles             = each.value.roles
  with_grant_option = false
}
