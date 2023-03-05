locals {
  # Flatten the nested data structure to table privileges.
  table_privileges = distinct(flatten([
    for schema in var.schemas : [
      for table in schema.tables : [
        for privilege in table.privileges : {
          schema    = schema.name
          table     = table.name
          privilege = privilege.name
          roles     = privilege.roles
        }
      ]
    ]
  ]))

  # Flatten the nested data structure to schema_privileges.
  schema_privileges = distinct(flatten([
    for schema in var.schemas : [
      for privilege in schema.privileges : {
        schema    = schema.name
        privilege = privilege.name
        roles     = privilege.roles
      }
    ]
  ]))

  # Flatten the nested data structure to tables.
  tables = distinct(flatten([
    for schema in var.schemas : [
      for table in schema.tables : {
        schema  = schema.name
        name    = table.name
        columns = table.columns
      }
    ]
  ]))
}


