resource "st-unitycatalog_permissions" "catalog_perms" {
  catalog    = "my_catalog"
  principal  = "myucuser"
  privileges = ["USE CATALOG"]
}

resource "st-unitycatalog_permissions" "schema_perms" {
  catalog    = "my_catalog"
  schema     = "my_schema"
  principal  = "myucuser"
  privileges = ["USE SCHEMA", "SELECT"]
}

resource "st-unitycatalog_permissions" "table_perms" {
  catalog    = "my_catalog"
  schema     = "my_schema"
  table      = "my_table"
  principal  = "myucuser"
  privileges = ["SELECT"]
}
