resource "snowflake_user" "users" {
  provider = snowflake.security_admin
  for_each = var.users

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
  for_each = var.roles

  name    = each.key
  comment = each.value.comment
}

resource "snowflake_role_grants" "new_roles" {
  provider = snowflake.security_admin
  # Explicit depends_on because we are using string to reference users.
  depends_on = [snowflake_user.users]
  for_each   = var.roles

  role_name = each.key
  users     = each.value.users
}
