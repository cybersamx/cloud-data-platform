resource "snowflake_database" "databases" {
  provider = snowflake.sys_admin
  for_each = toset(var.databases)

  name = each.value
}

resource "snowflake_schema" "schemas" {
  provider   = snowflake.sys_admin
  depends_on = [snowflake_database.databases]
  for_each   = { for v in local.schemas : "${v.database}.${v.name}" => v }

  database = each.value.database
  name     = each.value.name
  comment  = each.value.comment
}

resource "snowflake_table" "tables" {
  provider   = snowflake.sys_admin
  depends_on = [snowflake_schema.schemas]
  for_each   = { for v in local.tables : "${v.database}.${v.schema}.${v.name}" => v }

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

resource "snowflake_database_grant" "database_grants" {
  provider   = snowflake.security_admin
  depends_on = [snowflake_role.roles, snowflake_database.databases]
  for_each   = { for v in local.schema_privileges : "${v.database}.${v.schema}.${v.privilege}" => v }

  database_name     = each.value.database
  privilege         = each.value.privilege
  roles             = each.value.roles
  with_grant_option = false
}

resource "snowflake_schema_grant" "schema_grants" {
  provider   = snowflake.security_admin
  depends_on = [snowflake_role.roles, snowflake_schema.schemas]
  for_each   = { for v in local.schema_privileges : "${v.database}.${v.schema}.${v.privilege}" => v }

  database_name     = each.value.database
  schema_name       = each.value.schema
  privilege         = each.value.privilege
  roles             = each.value.roles
  with_grant_option = false
}

resource "snowflake_table_grant" "table_grants" {
  provider   = snowflake.sys_admin
  depends_on = [snowflake_role.roles, snowflake_table.tables]
  for_each   = { for v in local.table_privileges : "${v.database}.${v.schema}.${v.table}.${v.privilege}" => v }

  database_name     = each.value.database
  schema_name       = each.value.schema
  table_name        = each.value.table
  privilege         = each.value.privilege
  roles             = each.value.roles
  with_grant_option = false
}
