resource "snowflake_warehouse" "cdp" {
  provider = snowflake.sys_admin
  for_each = var.warehouses

  name                = each.value.name
  comment             = each.value.comment
  initially_suspended = true
  warehouse_size      = each.value.size
  auto_suspend        = each.value.auto_suspend
}
