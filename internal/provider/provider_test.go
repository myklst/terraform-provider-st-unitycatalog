package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// providerConfig is a shared configuration to combine with the actual test
// configuration so the Unity Catalog client is properly configured. It is also
// possible to use the UNITYCATALOG_ environment variables instead, such as
// updating the Makefile and running the testing through that tool.
const providerConfig = `
provider "st-unitycatalog" {
  host     = "http://localhost:8080"
  username = "admin"
  password = "password"
}
`

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"st-unitycatalog": providerserver.NewProtocol6WithError(New()),
}
