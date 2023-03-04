locals {
  new_roles = {
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
}

resource "snowflake_user" "users" {
  provider = snowflake.security_admin
  for_each = local.users

  name                 = each.key
  login_name           = each.key
  comment              = each.value.comment
  default_role         = each.value.role
  default_namespace    = "${snowflake_database.cdp.name}.PUBLIC"
  default_warehouse    = each.value.warehouse
  must_change_password = false
}

resource "snowflake_role" "roles" {
  provider = snowflake.security_admin
  for_each = local.new_roles

  name    = each.key
  comment = each.value.comment
}

resource "snowflake_role_grants" "new_roles" {
  provider = snowflake.security_admin
  # Explicit depends_on because we are using string to reference users.
  depends_on = [snowflake_user.users]
  for_each   = local.new_roles

  role_name = each.key
  users     = each.value.users
}
