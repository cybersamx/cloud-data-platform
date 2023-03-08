resource "snowflake_user" "users" {
  provider = snowflake.security_admin
  # for_each only accept map or list of string, so we need to convert the list to a map.
  for_each = { for item in local.users : "${item.database}.${item.name}" => item }

  name                 = each.value.name
  login_name           = each.value.name
  comment              = each.value.comment
  password             = each.value.password
  default_role         = each.value.default_role
  default_namespace    = each.value.default_namespace
  default_warehouse    = each.value.default_warehouse
  must_change_password = false
}

resource "snowflake_role" "roles" {
  provider = snowflake.security_admin
  for_each = var.roles

  name    = each.key
  comment = each.value.comment
}

resource "snowflake_role_grants" "role_grants" {
  provider = snowflake.security_admin
  # Explicit depends_on because we are using string to reference users.
  depends_on = [snowflake_user.users]
  for_each   = var.roles

  role_name = each.key
  users     = each.value.users
}
