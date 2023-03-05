resource "snowflake_database" "cdp" {
  provider = snowflake.sys_admin

  name = "CDP_DEV"
}

resource "snowflake_database_grant" "roles" {
  provider = snowflake.security_admin
  # Explicit depends_on because we are using string to reference schemas and role.
  depends_on = [snowflake_schema.schemas, snowflake_role.roles]
  # for_each only accept map or list of string, so we need to convert the list to a map.
  for_each = { for item in local.schema_privileges : "${item.schema}.${item.privilege}" => item }

  database_name     = snowflake_database.cdp.name
  privilege         = each.value.privilege
  roles             = each.value.roles
  with_grant_option = false
}

resource "snowflake_schema_grant" "roles" {
  provider = snowflake.security_admin
  # Explicit depends_on because we are using string to reference schemas and role.
  depends_on = [snowflake_schema.schemas, snowflake_role.roles]
  # for_each only accept map or list of string, so we need to convert the list to a map.
  for_each = { for item in local.schema_privileges : "${item.schema}.${item.privilege}" => item }

  database_name     = snowflake_database.cdp.name
  schema_name       = each.value.schema
  privilege         = each.value.privilege
  roles             = each.value.roles
  with_grant_option = false
}

resource "snowflake_schema" "schemas" {
  provider = snowflake.sys_admin
  for_each = var.schemas

  name     = each.key
  database = snowflake_database.cdp.name
  comment  = each.value.comment
}

resource "snowflake_table" "trips" {
  provider = snowflake.sys_admin
  for_each = var.schemas

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
