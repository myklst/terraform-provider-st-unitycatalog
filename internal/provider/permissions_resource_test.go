// Copyright IBM Corp. 2026

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccPermissionsResource covers the create, read, import, and in-place
// update (privileges) lifecycle of the st-unitycatalog_permissions resource
// at the catalog level.
func TestAccPermissionsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read.
			{
				Config: testAccPermissionsResourceConfig("my_catalog", "", "", "testuser", `"USE CATALOG"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("st-unitycatalog_permissions.test", "id"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "catalog", "my_catalog"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "principal", "testuser"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "privileges.#", "1"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "privileges.0", "USE CATALOG"),
				),
			},
			// ImportState verifies the resource can be imported and that the
			// imported state matches the prior state.
			{
				ResourceName:      "st-unitycatalog_permissions.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update privileges in place (add SELECT).
			{
				Config: testAccPermissionsResourceConfig("my_catalog", "", "", "testuser", `"USE CATALOG", "SELECT"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "privileges.#", "2"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "privileges.0", "USE CATALOG"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "privileges.1", "SELECT"),
				),
			},
		},
	})
}

// TestAccPermissionsResource_SchemaLevel verifies permissions at the schema
// level (catalog.schema).
func TestAccPermissionsResource_SchemaLevel(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsResourceConfig("my_catalog", "my_schema", "", "testuser", `"USE SCHEMA", "SELECT"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("st-unitycatalog_permissions.test", "id"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "catalog", "my_catalog"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "schema", "my_schema"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "principal", "testuser"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "privileges.#", "2"),
				),
			},
			// Import
			{
				ResourceName:      "st-unitycatalog_permissions.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccPermissionsResource_TableLevel verifies permissions at the table
// level (catalog.schema.table).
func TestAccPermissionsResource_TableLevel(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionsResourceConfig("my_catalog", "my_schema", "my_table", "testuser", `"SELECT"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("st-unitycatalog_permissions.test", "id"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "catalog", "my_catalog"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "schema", "my_schema"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "table", "my_table"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "principal", "testuser"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "privileges.#", "1"),
					resource.TestCheckResourceAttr("st-unitycatalog_permissions.test", "privileges.0", "SELECT"),
				),
			},
			// Import
			{
				ResourceName:      "st-unitycatalog_permissions.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// testAccPermissionsResourceConfig renders a complete acceptance-test
// configuration for the st-unitycatalog_permissions resource. The privileges
// argument is a pre-formatted HCL list body (e.g. `"USE CATALOG", "SELECT"`).
func testAccPermissionsResourceConfig(catalog, schema, table, principal, privileges string) string {
	schemaBlock := ""
	if schema != "" {
		schemaBlock = fmt.Sprintf("  schema     = %q\n", schema)
	}
	tableBlock := ""
	if table != "" {
		tableBlock = fmt.Sprintf("  table      = %q\n", table)
	}
	return fmt.Sprintf(`
%[1]s

resource "st-unitycatalog_permissions" "test" {
  catalog    = %[2]q
%[3]s%[4]s  principal  = %[5]q
  privileges = [%[6]s]
}
`, providerConfig, catalog, schemaBlock, tableBlock, principal, privileges)
}
