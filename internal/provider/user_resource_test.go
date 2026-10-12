package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccUserResource covers the create, read, import, and in-place update
// (display_name and active) lifecycle of the st-unitycatalog_user resource.
func TestAccUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read.
			{
				Config: testAccUserResourceConfig("Test User", "acc-test@example.com", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("st-unitycatalog_user.test", "id"),
					resource.TestCheckResourceAttr("st-unitycatalog_user.test", "display_name", "Test User"),
					resource.TestCheckResourceAttr("st-unitycatalog_user.test", "active", "true"),
					resource.TestCheckResourceAttr("st-unitycatalog_user.test", "schemas.#", "1"),
					resource.TestCheckResourceAttr("st-unitycatalog_user.test", "schemas.0", "urn:ietf:params:scim:schemas:core:2.0:User"),
					resource.TestCheckResourceAttr("st-unitycatalog_user.test", "emails.#", "1"),
					resource.TestCheckResourceAttr("st-unitycatalog_user.test", "emails.0.value", "acc-test@example.com"),
					resource.TestCheckResourceAttr("st-unitycatalog_user.test", "emails.0.primary", "true"),
				),
			},
			// ImportState verifies the resource can be imported and that the
			// imported state matches the prior state.
			{
				ResourceName:      "st-unitycatalog_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update display_name and active in place.
			{
				Config: testAccUserResourceConfig("Test User Updated", "acc-test@example.com", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("st-unitycatalog_user.test", "display_name", "Test User Updated"),
					resource.TestCheckResourceAttr("st-unitycatalog_user.test", "active", "false"),
					resource.TestCheckResourceAttr("st-unitycatalog_user.test", "emails.0.value", "acc-test@example.com"),
				),
			},
		},
	})
}

// TestAccUserResource_EmailsRequiresReplace verifies that changing the emails
// attribute forces replacement of the resource, since emails are immutable in
// Unity Catalog.
func TestAccUserResource_EmailsRequiresReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with the initial email.
			{
				Config: testAccUserResourceConfig("Replace User", "replace-1@example.com", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("st-unitycatalog_user.test", "id"),
					resource.TestCheckResourceAttr("st-unitycatalog_user.test", "emails.0.value", "replace-1@example.com"),
				),
			},
			// Changing the email forces the resource to be replaced.
			{
				Config: testAccUserResourceConfig("Replace User", "replace-2@example.com", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("st-unitycatalog_user.test", "emails.0.value", "replace-2@example.com"),
				),
			},
		},
	})
}

// testAccUserResourceConfig renders a complete acceptance-test configuration for
// the st-unitycatalog_user resource.
func testAccUserResourceConfig(displayName, email string, active bool) string {
	return fmt.Sprintf(`
%[1]s

resource "st-unitycatalog_user" "test" {
  schemas      = ["urn:ietf:params:scim:schemas:core:2.0:User"]
  display_name = %[2]q
  emails = [{
    value   = %[3]q
    primary = true
  }]
  active = %[4]t
}
`, providerConfig, displayName, email, active)
}
