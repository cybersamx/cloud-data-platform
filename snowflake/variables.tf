# Pass the value thru environment variable TF_VAR_user_password.
variable "user_password" {
  description = "New users' password - they all share the same password."
  type        = string
}

variable "roles" {
  type = list(object({
    name    = string
    comment = optional(string)
  }))

  default = [
    {
      name    = "LOADER"
      comment = "Owns the raw database and connects to the loading warehouse"
    },
    {
      name    = "TRANSFORMER"
      comment = "Owns the analytics database and has query permission on the raw database"
    },
    {
      name    = "REPORTER"
      comment = "Has query permission on the analytics database"
    }
  ]
}

variable "role_grants" {
  type = list(object({
    role  = string
    users = list(string)
  }))

  default = [
    {
      role  = "SYSADMIN"
      users = ["AIRBYTE", "DBT"]
    },
    {
      role  = "LOADER"
      users = ["AIRBYTE"]
    },
    {
      role  = "TRANSFORMER"
      users = ["DBT"]
    },
    {
      role  = "REPORTER"
      users = ["PYTHON"]
    }
  ]
}

variable "users" {
  type = map(object({
    name              = string
    comment           = optional(string)
    default_role      = string
    default_namespace = string
    default_warehouse = optional(string)
  }))

  default = {
    AIRBYTE = {
      name              = "AIRBYTE"
      comment           = "The service account for airbyte"
      default_role      = "LOADER"
      default_namespace = "CDP_DEV.RAW"
      default_warehouse = "LOADING"
    }
    DBT = {
      name              = "DBT"
      comment           = "The service account for dbt"
      default_role      = "TRANSFORMER"
      default_namespace = "CDP_DEV.RAW"
      default_warehouse = "TRANSFORMING"
    }
    PYTHON = {
      name              = "PYTHON"
      comment           = "Ad-hoc python script for analytics"
      default_role      = "REPORTER"
      default_namespace = "CDP_DEV.ANALYTICS"
      default_warehouse = "REPORTING"
    }
  }
}

variable "warehouses" {
  type = map(object({
    name         = string
    comment      = optional(string)
    size         = optional(string, "X-Small")
    auto_suspend = optional(number, 15)
  }))

  default = {
    LOADING = {
      name    = "LOADING"
      comment = "Runs the loading of raw data in the raw schema."
    }
    TRANSFORMING = {
      name    = "TRANSFORMING"
      comment = "Runs the transformation of data."
    }
    REPORTING = {
      name    = "REPORTING"
      comment = "Runs the queries of analytical and reporting tools."
    }
  }
}

variable "databases" {
  type    = list(string)
  default = ["CDP_DEV"]
}

variable "schemas" {
  type = map(object({
    name    = string
    comment = string
    privileges = map(object({
      name  = string
      roles = list(string)
    }))
    table_privileges = map(object({
      name  = string
      roles = list(string)
    }))
  }))

  default = {
    RAW = {
      name    = "RAW"
      comment = "Dump of raw data extracted from data sources"

      privileges = {
        USAGE = {
          name  = "USAGE"
          roles = ["LOADER", "TRANSFORMER"]
        }
        MODIFY = {
          name  = "MODIFY"
          roles = ["LOADER"]
        }
      }

      table_privileges = {
        SELECT = {
          name  = "SELECT"
          roles = ["LOADER", "TRANSFORMER"]
        }
        INSERT = {
          name  = "INSERT"
          roles = ["TRANSFORMER"]
        }
      }
    }
    ANALYTICS = {
      name    = "ANALYTICS"
      comment = "Tables and views for analytics and reporting"

      privileges = {
        MODIFY = {
          name  = "MODIFY"
          roles = ["TRANSFORMER"]
        }
      }

      table_privileges = {
        SELECT = {
          name  = "SELECT"
          roles = ["TRANSFORMER"]
        }
        INSERT = {
          name  = "INSERT"
          roles = ["TRANSFORMER"]
        }
      }
    }
  }
}
