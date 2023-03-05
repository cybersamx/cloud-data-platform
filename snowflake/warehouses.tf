resource "snowflake_warehouse" "cdp" {
  provider = snowflake.sys_admin

  name           = "CDP_DEV"
  warehouse_size = "X-Small"
  # Auto suspend after seconds of inactivity, since this is a prototype, keep the value low to save $
  auto_suspend = 15
}
