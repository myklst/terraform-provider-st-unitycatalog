// Copyright IBM Corp. 2026

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/myklst/terraform-provider-st-unitycatalog/client"
)

var (
	_ datasource.DataSource = &tablesDataSource{}
)

// NewTablesDataSource creates a new Unity Catalog tables data source.
func NewTablesDataSource() datasource.DataSource {
	return &tablesDataSource{}
}

type tablesDataSource struct {
	client *client.Client
}

// tableModel maps a single table from the API response to Terraform state.
type tableModel struct {
	CatalogName types.String `tfsdk:"catalog_name"`
	SchemaName  types.String `tfsdk:"schema_name"`
	Name        types.String `tfsdk:"name"`
	TableType   types.String `tfsdk:"table_type"`
	Comment     types.String `tfsdk:"comment"`
}

// tablesDataSourceModel maps the data source read result to Terraform state.
type tablesDataSourceModel struct {
	CatalogName types.String `tfsdk:"catalog_name"`
	SchemaName  types.String `tfsdk:"schema_name"`
	Tables      types.List   `tfsdk:"tables"`
	Id          types.String `tfsdk:"id"`
}

var tableAttrTypes = map[string]attr.Type{
	"catalog_name": types.StringType,
	"schema_name":  types.StringType,
	"name":         types.StringType,
	"table_type":   types.StringType,
	"comment":      types.StringType,
}

func (d *tablesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tables"
}

func (d *tablesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists tables in a Unity Catalog catalog, optionally filtered by schema, " +
			"via GET /api/2.1/unity-catalog/tables?catalog_name=...&schema_name=...",
		Attributes: map[string]schema.Attribute{
			"catalog_name": schema.StringAttribute{
				Description: "The catalog name to list tables from. Required.",
				Required:    true,
			},
			"schema_name": schema.StringAttribute{
				Description: "The schema name to filter tables by. When omitted, all tables " +
					"in the catalog are returned.",
				Optional: true,
			},
			"tables": schema.ListNestedAttribute{
				Description: "The list of tables in the catalog (or schema).",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"catalog_name": schema.StringAttribute{
							Description: "The catalog name.",
							Computed:    true,
						},
						"schema_name": schema.StringAttribute{
							Description: "The schema name.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The table name.",
							Computed:    true,
						},
						"table_type": schema.StringAttribute{
							Description: "The table type (e.g. MANAGED, EXTERNAL, VIEW).",
							Computed:    true,
						},
						"comment": schema.StringAttribute{
							Description: "An optional comment on the table.",
							Computed:    true,
						},
					},
				},
			},
			"id": schema.StringAttribute{
				Description: "The data source identifier: catalog_name[.schema_name].",
				Computed:    true,
			},
		},
	}
}

func (d *tablesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *tablesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config tablesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogName := config.CatalogName.ValueString()
	schemaName := ""
	if !config.SchemaName.IsNull() {
		schemaName = config.SchemaName.ValueString()
	}

	result, err := d.client.ListTables(ctx, catalogName, schemaName)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list Unity Catalog tables", err.Error())
		return
	}

	tables := make([]tableModel, 0, len(result.Tables))
	for _, t := range result.Tables {
		tables = append(tables, tableModel{
			CatalogName: types.StringValue(t.CatalogName),
			SchemaName:  types.StringValue(t.SchemaName),
			Name:        types.StringValue(t.Name),
			TableType:   types.StringValue(t.TableType),
			Comment:     types.StringValue(t.Comment),
		})
	}

	tablesList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: tableAttrTypes}, tables)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Tables = tablesList

	id := catalogName
	if schemaName != "" {
		id = fmt.Sprintf("%s.%s", catalogName, schemaName)
	}
	config.Id = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
