locals {
  # Flatten the nested data structures.
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
}
