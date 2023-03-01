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
}

resource "snowflake_warehouse" "cdp" {
  provider = snowflake.sys_admin

  name           = "CDP-DEV"
  warehouse_size = "X-Small"
  # Auto suspend after seconds of inactivity, since this is a prototype, keep the value low to save $
  auto_suspend = 15
}
