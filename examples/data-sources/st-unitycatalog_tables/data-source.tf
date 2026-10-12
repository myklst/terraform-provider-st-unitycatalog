data "st-unitycatalog_tables" "all" {
  catalog_name = "my-catalog"
}

data "st-unitycatalog_tables" "by_schema" {
  catalog_name = "my-catalog"
  schema_name  = "my-catalog-schema"
}
