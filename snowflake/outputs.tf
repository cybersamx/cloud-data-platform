output "database_grant_roles" {
  value = {
    for k, v in snowflake_database_grant.roles : k => {
      id    = v.id
      roles = v.roles
    }
  }
}

output "schema_grant_roles" {
  value = {
    for k, v in snowflake_schema_grant.roles : k => {
      id    = v.id
      roles = v.roles
    }
  }
}

output "tables" {
  value = [
    for v in snowflake_table.all : {
      id     = v.id
      schema = v.schema
      name   = v.name
    }
  ]
}
