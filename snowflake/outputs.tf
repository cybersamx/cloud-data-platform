output "database_grant_roles" {
  value = {
    for k, v in snowflake_database_grant.database_grants : k => {
      id    = v.id
      roles = v.roles
    }
  }
}

output "schema_grant_roles" {
  value = {
    for k, v in snowflake_schema_grant.schema_grants : k => {
      id    = v.id
      roles = v.roles
    }
  }
}
