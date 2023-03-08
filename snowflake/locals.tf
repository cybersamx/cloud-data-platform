locals {
  # Flatten the nested data structure to users.
  users = distinct(flatten([
    for database in var.databases : [
      for user in var.users : {
        database          = database
        name              = user.name
        comment           = user.comment
        password          = var.user_password
        default_role      = user.default_role
        default_namespace = user.default_namespace
        default_warehouse = user.default_warehouse
      }
    ]
  ]))

  # Flatten the nested data structure to schemas.
  schemas = distinct(flatten([
    for database in var.databases : [
      for schema in var.schemas : {
        database = database
        name     = schema.name
        comment  = schema.comment
      }
    ]
  ]))

  # Flatten the nested data structure to tables.
  tables = distinct(flatten([
    for database in var.databases : [
      for schema in var.schemas : [
        for table in schema.tables : {
          database = database
          schema   = schema.name
          name     = table.name
          columns  = table.columns
        }
      ]
    ]
  ]))

  # Flatten the nested data structure to table privileges.
  table_privileges = distinct(flatten([
    for database in var.databases : [
      for schema in var.schemas : [
        for table in schema.tables : [
          for privilege in table.privileges : {
            database  = database
            schema    = schema.name
            table     = table.name
            privilege = privilege.name
            roles     = privilege.roles
          }
        ]
      ]
    ]
  ]))

  # Flatten the nested data structure to schema_privileges.
  schema_privileges = distinct(flatten([
    for database in var.databases : [
      for schema in var.schemas : [
        for privilege in schema.privileges : {
          database  = database
          schema    = schema.name
          privilege = privilege.name
          roles     = privilege.roles
        }
      ]
    ]
  ]))
}


