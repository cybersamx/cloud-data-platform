variable "database_mame" {
  type    = string
  default = "CDP_DEV"
}

variable "roles" {
  type = map(object({
    name    = string
    comment = optional(string)
    users   = list(string)
  }))

  default = {
    "LOADER" = {
      name    = "LOADER"
      comment = "Owns the raw database and connects to the loading warehouse"
      users   = ["AIRBYTE"]
    }
    "TRANSFORMER" = {
      name    = "TRANSFORMER"
      comment = "Owns the analytics database and has query permission on the raw database"
      users   = ["DBT"]
    }
    "REPORTER" = {
      name    = "REPORTER"
      comment = "Has query permission on the analytics database"
      users   = ["PYTHON"]
    }
  }
}

variable "users" {
  type = map(object({
    name      = string
    comment   = optional(string)
    role      = string
    warehouse = optional(string)
  }))

  default = {
    "AIRBYTE" = {
      name      = "AIRBYTE"
      comment   = "The service account for airbyte"
      role      = "LOADER"
      warehouse = "LOADING"
    }
    "DBT" = {
      name      = "DBT"
      comment   = "The service account for dbt"
      role      = "LOADER"
      warehouse = "TRANSFORMING"
    }
    "PYTHON" = {
      name      = "PYTHON"
      comment   = "Ad-hoc python script for analytics"
      role      = "REPORTER"
      warehouse = "REPORTING"
    }
  }
}

variable "warehouses" {
  type = map(object({
    name         = string
    comment      = optional(string)
    size         = optional(string, "X-Small")
    auto_suspend = optional(number, 30)
  }))

  default = {
    "LOADING" = {
      name    = "LOADING"
      comment = "Runs the loading of raw data in the raw schema."
    }
    "TRANSFORMING" = {
      name    = "TRANSFORMING"
      comment = "Runs the transformation of data."
    }
    "REPORTING" = {
      name    = "REPORTING"
      comment = "Runs the queries of analytical and reporting tools."
    }
  }
}

variable "databases" {
  type    = list(string)
  default = ["CDP_DEV", "CDP_PROD"]
}

variable "schemas" {
  type = map(object({
    name    = string
    comment = string
    privileges = map(object({
      name  = string
      roles = list(string)
    }))
    tables = map(object({
      name = string
      privileges = map(object({
        name  = string
        roles = list(string)
      }))
      columns = map(object({
        type     = string
        nullable = optional(bool, true)
      }))
    }))
  }))

  default = {
    "RAW" = {
      name    = "RAW"
      comment = "Dump of raw data extracted from data sources"

      privileges = {
        "USAGE" = {
          name  = "USAGE"
          roles = ["LOADER", "TRANSFORMER"]
        }
        "MODIFY" = {
          name  = "MODIFY"
          roles = ["LOADER"]
        }
      }

      tables = {
        "TRIPS" = {
          name = "TRIPS"
          privileges = {
            "SELECT" = {
              name  = "SELECT"
              roles = ["LOADER", "TRANSFORMER"]
            }
            "INSERT" = {
              name  = "INSERT"
              roles = ["TRANSFORMER"]
            }
          }
          columns = {
            "trip_duration" = {
              type = "VARCHAR(16777216)"
            }
            "start_time" = {
              type = "VARCHAR(16777216)"
            }
            "stop_time" = {
              type = "VARCHAR(16777216)"
            }
            "start_station_id" = {
              type = "VARCHAR(16777216)"
            }
            "start_station_name" = {
              type = "VARCHAR(16777216)"
            }
            "start_station_latitude" = {
              type = "VARCHAR(16777216)"
            }
            "start_station_longitude" = {
              type = "VARCHAR(16777216)"
            }
            "end_station_id" = {
              type = "VARCHAR(16777216)"
            }
            "end_station_name" = {
              type = "VARCHAR(16777216)"
            }
            "end_station_latitude" = {
              type = "VARCHAR(16777216)"
            }
            "end_station_longitude" = {
              type = "VARCHAR(16777216)"
            }
            "bike_id" = {
              type = "VARCHAR(16777216)"
            }
            "membership_type" = {
              type = "VARCHAR(16777216)"
            }
            "usertype" = {
              type = "VARCHAR(16777216)"
            }
            "birth_year" = {
              type = "VARCHAR(16777216)"
            }
            "gender" = {
              type = "VARCHAR(16777216)"
            }
          }
        }
      }
    }
    ANALYTICS = {
      name    = "ANALYTICS"
      comment = "Tables and views for analytics and reporting"

      privileges = {
        "MODIFY" = {
          name  = "MODIFY"
          roles = ["TRANSFORMER"]
        }
      }

      tables = {
        "TRIPS" = {
          name = "TRIPS"
          privileges = {
            "SELECT" = {
              name  = "SELECT"
              roles = ["TRANSFORMER"]
            }
            "INSERT" = {
              name  = "INSERT"
              roles = ["TRANSFORMER"]
            }
          }
          columns = {
            "trip_duration" = {
              // For types use capitalized, basic types like NUMBER(38,0) than lowercase, alias types like
              // integer or text. While NUMBER(38,0) is semantically equivalent to number(38,0), integer or INTEGER,
              // the current Snowflake provider does an exact case sensitive match of the the string between the
              // value in the terraform code and state. As a result, INTEGER or number(38,0) will be marked as change.
              type = "NUMBER(38,0)"
            }
            "start_time" = {
              type = "TIMESTAMP_NTZ(9)"
            }
            "stop_time" = {
              type = "TIMESTAMP_NTZ(9)"
            }
            "start_station_id" = {
              type = "NUMBER(38,0)"
            }
            "start_station_name" = {
              type = "VARCHAR(16777216)"
            }
            "start_station_latitude" = {
              type = "FLOAT"
            }
            "start_station_longitude" = {
              type = "FLOAT"
            }
            "end_station_id" = {
              type = "NUMBER(38,0)"
            }
            "end_station_name" = {
              type = "VARCHAR(16777216)"
            }
            "end_station_latitude" = {
              type = "FLOAT"
            }
            "end_station_longitude" = {
              type = "FLOAT"
            }
            "bike_id" = {
              type = "NUMBER(38,0)"
            }
            "membership_type" = {
              type = "VARCHAR(16777216)"
            }
            "usertype" = {
              type = "VARCHAR(16777216)"
            }
            "birth_year" = {
              type = "NUMBER(38,0)"
            }
            "gender" = {
              type = "NUMBER(38,0)"
            }
          }
        }
      }
    }
  }
}
