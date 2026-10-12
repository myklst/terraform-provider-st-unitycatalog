// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccTablesDataSource verifies reading tables at the catalog level (no
// schema filter) and at the schema level (with schema filter).
func TestAccTablesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Catalog-level: list all tables in the catalog.
			{
				Config: testAccTablesDataSourceConfig("my_catalog", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.st-unitycatalog_tables.test", "catalog_name", "my_catalog"),
					resource.TestCheckResourceAttrSet("data.st-unitycatalog_tables.test", "id"),
					resource.TestCheckResourceAttrSet("data.st-unitycatalog_tables.test", "tables.#"),
				),
			},
		},
	})
}

// TestAccTablesDataSource_SchemaFilter verifies reading tables filtered by
// schema.
func TestAccTablesDataSource_SchemaFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTablesDataSourceConfig("my_catalog", "my_schema"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.st-unitycatalog_tables.test", "catalog_name", "my_catalog"),
					resource.TestCheckResourceAttr("data.st-unitycatalog_tables.test", "schema_name", "my_schema"),
					resource.TestCheckResourceAttr("data.st-unitycatalog_tables.test", "id", "my_catalog.my_schema"),
					resource.TestCheckResourceAttrSet("data.st-unitycatalog_tables.test", "tables.#"),
				),
			},
		},
	})
}

func testAccTablesDataSourceConfig(catalogName, schemaName string) string {
	schemaBlock := ""
	if schemaName != "" {
		schemaBlock = fmt.Sprintf("  schema_name = %q\n", schemaName)
	}
	return fmt.Sprintf(`
%[1]s

data "st-unitycatalog_tables" "test" {
  catalog_name = %[2]q
%[3]s}
`, providerConfig, catalogName, schemaBlock)
}
