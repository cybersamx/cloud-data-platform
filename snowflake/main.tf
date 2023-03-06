terraform {
  required_version = ">= 1.3"

  required_providers {
    snowflake = {
      source  = "snowflake-labs/snowflake"
      version = "~> 0.56"
    }
  }
}

# Allows us to create Snowflake resources with a specific owner.

provider "snowflake" {
  alias = "sys_admin"
  role  = "SYSADMIN"
}

provider "snowflake" {
  alias = "security_admin"
  role  = "SECURITYADMIN"
}
