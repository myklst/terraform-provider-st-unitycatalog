package main

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/myklst/terraform-provider-st-unitycatalog/internal/provider"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// Format Terraform code for use in documentation.
// If you do not have Terraform installed, you can remove the formatting command, but it is suggested
// to ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name st-unitycatalog

func main() {
	providerAddress := os.Getenv("PROVIDER_LOCAL_PATH")
	if providerAddress == "" {
		providerAddress = "registry.terraform.io/myklst/st-unitycatalog"
	}
	providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: providerAddress,
	})
}
