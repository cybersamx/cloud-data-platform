terraform {
  required_providers {
    snowflake = {
      source  = "snowflake-labs/snowflake"
      version = "~> 0.56"
    }
  }
}

locals {
  warehouses = {
    "LOADING" = {
      comment = "Runs the loading of raw data in the raw schema"
    }
    "TRANSFORMING" = {
      comment = "Runs the transformation of data"
    }
    "REPORTING" = {
      comment = "Runs the queries of analytical and reporting tools"
    }
  }

  roles = {
    "LOADER" = {
      comment = "Owns the raw database and connects to the loading warehouse"
      users   = ["AIRBYTE"]
    }
    "TRANSFORMER" = {
      comment = "Owns the analytics database and has query permission on the raw database"
      users   = ["DBT"]
    }
    "REPORTER" = {
      comment = "Has query permission on the analytics database"
      users   = ["PYTHON"]
    }
  }

  users = {
    "AIRBYTE" = {
      comment   = "The service account for airbyte"
      role      = "LOADER"
      warehouse = "LOADING"
    }
    "DBT" = {
      comment   = "The service account for dbt"
      role      = "LOADER"
      warehouse = "TRANSFORMING"
    }
    "PYTHON" = {
      comment   = "Ad-hoc python script for analytics"
      role      = "REPORTER"
      warehouse = "REPORTING"
    }
  }

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

provider "snowflake" {
  role = "SECURITYADMIN"
}

resource "snowflake_warehouse" "cdp" {
  name           = "CDP-DEV"
  warehouse_size = "X-Small"
  # Auto suspend after seconds of inactivity, since this is a prototype, keep the value low to save $
  auto_suspend = 15
}

resource "snowflake_database" "cdp" {
  name = "CDP-DEV"
}

resource "snowflake_schema" "schemas" {
  for_each = local.schemas

  name     = each.key
  database = snowflake_database.cdp.name
  comment  = each.value.comment
}

resource "snowflake_role" "role" {
  for_each = local.roles

  name    = each.key
  comment = each.value.comment
}

resource "snowflake_role_grants" "roles" {
  for_each = local.roles

  role_name = each.key
  users     = each.value.users
  roles     = []
}

resource "snowflake_schema_grant" "usage_roles" {
  for_each = local.schemas

  schema_name   = each.key
  database_name = snowflake_database.cdp.name
  privilege     = "USAGE"
  roles         = each.value.usage_roles
  shares        = []
}

resource "snowflake_schema_grant" "modify_roles" {
  for_each = local.schemas

  schema_name   = each.key
  database_name = snowflake_database.cdp.name
  privilege     = "MODIFY"
  roles         = each.value.modify_roles
  shares        = []
}

resource "snowflake_user" "user" {
  for_each = local.users

  name                 = each.key
  login_name           = each.key
  comment              = each.value.comment
  default_role         = each.value.role
  default_namespace    = "${snowflake_database.cdp.name}.PUBLIC"
  default_warehouse    = each.value.warehouse
  must_change_password = false
}
