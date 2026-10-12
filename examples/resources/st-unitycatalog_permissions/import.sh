# Import by composite ID: catalog[.schema[.table]]#principal
terraform import 'st-unitycatalog_permissions.catalog_perms' 'my_catalog#myucuser'
terraform import 'st-unitycatalog_permissions.schema_perms' 'my_catalog.my_schema#myucuser'
terraform import 'st-unitycatalog_permissions.table_perms' 'my_catalog.my_schema.my_table#myucuser'
