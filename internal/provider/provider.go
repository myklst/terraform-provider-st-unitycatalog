package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/myklst/terraform-provider-st-unitycatalog/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &unitycatalogProvider{}
)

// New is a helper function to simplify provider server instantiation.
func New() provider.Provider {
	return &unitycatalogProvider{}
}

type unitycatalogProvider struct{}

type unitycatalogProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Token    types.String `tfsdk:"token"`
}

// Metadata returns the provider type name.
func (p *unitycatalogProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "st-unitycatalog"
}

// Schema defines the provider-level schema for configuration data.
func (p *unitycatalogProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The Unity Catalog provider is used to interact with Unity Catalog resources " +
			"through the SCIM2 API. The provider needs to be configured with the proper credentials " +
			"before it can be used.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The base URL of the Unity Catalog control plane (e.g. https://unity.example.com). " +
					"May also be provided via the UNITYCATALOG_HOST environment variable.",
				Optional: true,
			},
			"username": schema.StringAttribute{
				Description: "Username for Unity Catalog SCIM2 API authentication, used for HTTP Basic auth. " +
					"Ignored when a token is configured. May also be provided via the UNITYCATALOG_USERNAME " +
					"environment variable. Optional: leave unset when the server has authentication disabled " +
					"or when authenticating with a token.",
				Optional: true,
			},
			"password": schema.StringAttribute{
				Description: "Password for Unity Catalog SCIM2 API authentication, used for HTTP Basic auth. " +
					"Ignored when a token is configured. May also be provided via the UNITYCATALOG_PASSWORD " +
					"environment variable. Optional: leave unset when the server has authentication disabled " +
					"or when authenticating with a token.",
				Optional:  true,
				Sensitive: true,
			},
			"token": schema.StringAttribute{
				Description: "Bearer token for Unity Catalog SCIM2 API authentication, sent as " +
					"'Authorization: Bearer <token>'. When set, it takes precedence over username/password. " +
					"May also be provided via the UNITYCATALOG_TOKEN environment variable. Optional: leave " +
					"unset when the server has authentication disabled or when authenticating with " +
					"username/password.",
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares a Unity Catalog SCIM2 API client for data sources and
// resources.
func (p *unitycatalogProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config unitycatalogProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Unity Catalog host",
			"The provider cannot create the Unity Catalog API client as there is an unknown configuration value for the "+
				"Unity Catalog host. Set the value statically in the configuration, or use the UNITYCATALOG_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Unity Catalog username",
			"The provider cannot create the Unity Catalog API client as there is an unknown configuration value for the "+
				"Unity Catalog username. Set the value statically in the configuration, or use the UNITYCATALOG_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Unity Catalog password",
			"The provider cannot create the Unity Catalog API client as there is an unknown configuration value for the "+
				"Unity Catalog password. Set the value statically in the configuration, or use the UNITYCATALOG_PASSWORD environment variable.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Unity Catalog token",
			"The provider cannot create the Unity Catalog API client as there is an unknown configuration value for the "+
				"Unity Catalog token. Set the value statically in the configuration, or use the UNITYCATALOG_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override with Terraform
	// configuration value if set.
	var host, username, password, token string
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	} else {
		host = os.Getenv("UNITYCATALOG_HOST")
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	} else {
		username = os.Getenv("UNITYCATALOG_USERNAME")
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	} else {
		password = os.Getenv("UNITYCATALOG_PASSWORD")
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	} else {
		token = os.Getenv("UNITYCATALOG_TOKEN")
	}

	// The host is always required to know which server to call. Authentication
	// (token or username/password) is optional: Unity Catalog may be configured
	// with authentication disabled.
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Unity Catalog host",
			"The provider cannot create the Unity Catalog API client as there is a "+
				"missing or empty value for the Unity Catalog host. Set the "+
				"host value in the configuration or use the UNITYCATALOG_HOST "+
				"environment variable. If either is already set, ensure the value "+
				"is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	scimClient := client.New(host, username, password, token)

	resp.DataSourceData = scimClient
	resp.ResourceData = scimClient
}

func (p *unitycatalogProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewTablesDataSource,
	}
}

func (p *unitycatalogProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewPermissionsResource,
	}
}
