resource "snowflake_database" "all" {
  provider = snowflake.sys_admin
  for_each = toset(var.databases)

  name = each.value
}

resource "snowflake_database_grant" "roles" {
  provider   = snowflake.security_admin
  depends_on = [snowflake_role.roles, snowflake_database.all]
  # for_each only accept map or list of string, so we need to convert the list to a map.
  for_each = { for item in local.schema_privileges : "${item.database}.${item.schema}.${item.privilege}" => item }

  database_name     = each.value.database
  privilege         = each.value.privilege
  roles             = each.value.roles
  with_grant_option = false
}

resource "snowflake_schema_grant" "roles" {
  provider   = snowflake.security_admin
  depends_on = [snowflake_role.roles, snowflake_schema.all]
  # for_each only accept map or list of string, so we need to convert the list to a map.
  for_each = { for item in local.schema_privileges : "${item.database}.${item.schema}.${item.privilege}" => item }

  database_name     = each.value.database
  schema_name       = each.value.schema
  privilege         = each.value.privilege
  roles             = each.value.roles
  with_grant_option = false
}

resource "snowflake_schema" "all" {
  provider   = snowflake.sys_admin
  depends_on = [snowflake_database.all]
  # for_each only accept map or list of string, so we need to convert the list to a map.
  for_each = { for item in local.schemas : "${item.database}.${item.name}" => item }

  database = each.value.database
  name     = each.value.name
  comment  = each.value.comment
}

resource "snowflake_table" "all" {
  provider   = snowflake.sys_admin
  depends_on = [snowflake_schema.all]
  # for_each only accept map or list of string, so we need to convert the list to a map.
  for_each = { for item in local.tables : "${item.database}.${item.schema}.${item.name}" => item }

  database = each.value.database
  schema   = each.value.schema
  name     = each.value.name

  dynamic "column" {
    for_each = each.value.columns
    content {
      name     = column.key
      type     = lookup(column.value, "type", "VARCHAR(16777216)")
      nullable = lookup(column.value, "nullable", true)
    }
  }
}

resource "snowflake_table_grant" "trips" {
  provider   = snowflake.sys_admin
  depends_on = [snowflake_role.roles, snowflake_table.all]
  # for_each only accept map or list of string, so we need to convert the list to a map.
  for_each = { for item in local.table_privileges : "${item.database}.${item.schema}.${item.table}.${item.privilege}" => item }

  database_name     = each.value.database
  schema_name       = each.value.schema
  table_name        = each.value.table
  privilege         = each.value.privilege
  roles             = each.value.roles
  with_grant_option = false
}
