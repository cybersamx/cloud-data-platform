resource "snowflake_user" "users" {
  provider = snowflake.security_admin
  # for_each only accept map or list of string, so we need to convert the list to a map.
  for_each = { for v in local.users : "${v.database}.${v.name}" => v }

  name                 = each.value.name
  login_name           = each.value.name
  comment              = each.value.comment
  password             = each.value.password
  default_role         = each.value.default_role
  default_namespace    = each.value.default_namespace
  default_warehouse    = each.value.default_warehouse
  must_change_password = false
}

# Create new roles.
resource "snowflake_role" "roles" {
  provider = snowflake.security_admin
  for_each = { for v in var.roles : v.name => v }

  name    = each.key
  comment = each.value.comment
}

resource "snowflake_role_grants" "role_grants" {
  provider   = snowflake.security_admin
  depends_on = [snowflake_user.users, snowflake_role.roles]
  for_each   = { for v in var.role_grants : v.role => v }

  role_name = each.key
  users     = each.value.users
}
