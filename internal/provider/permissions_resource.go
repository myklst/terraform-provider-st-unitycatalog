// Copyright IBM Corp. 2026

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/myklst/terraform-provider-st-unitycatalog/client"
)

var (
	_ resource.Resource                = &permissionsResource{}
	_ resource.ResourceWithConfigure   = &permissionsResource{}
	_ resource.ResourceWithImportState = &permissionsResource{}
)

// NewPermissionsResource creates a new Unity Catalog permissions resource.
func NewPermissionsResource() resource.Resource {
	return &permissionsResource{}
}

type permissionsResource struct {
	client *client.Client
}

// permissionsResourceModel maps the Unity Catalog permissions attributes to
// Terraform state. The securable is identified by the combination of catalog,
// schema (optional), and table (optional). Changing any of these or the
// principal forces replacement; only privileges can be updated in place.
type permissionsResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Catalog    types.String `tfsdk:"catalog"`
	Schema     types.String `tfsdk:"schema"`
	Table      types.String `tfsdk:"table"`
	Principal  types.String `tfsdk:"principal"`
	Privileges types.List   `tfsdk:"privileges"`
}

func (r *permissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions"
}

func (r *permissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages privileges for a principal on a Unity Catalog securable " +
			"(catalog, schema, or table) via the Permissions API " +
			"(/api/2.1/unity-catalog/permissions/{securable_type}/{full_name}).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier: {catalog}[.{schema}[.{table}]]#{principal}.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"catalog": schema.StringAttribute{
				Description: "The catalog name. Required; identifies the securable at the " +
					"catalog level when schema and table are unset.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schema": schema.StringAttribute{
				Description: "The schema name. When set, the securable is at the schema level " +
					"(full name: catalog.schema). Changing this forces replacement.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"table": schema.StringAttribute{
				Description: "The table name. When set (requires schema), the securable is at " +
					"the table level (full name: catalog.schema.table). Changing this forces " +
					"replacement.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"principal": schema.StringAttribute{
				Description: "The Unity Catalog user or group to grant privileges to. " +
					"Changing this forces replacement.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"privileges": schema.ListAttribute{
				Description: "The list of privileges to grant, e.g. " +
					"[\"USE CATALOG\", \"SELECT\"]. Can be updated in place; the provider " +
					"computes the add/remove diff against the prior state.",
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}
}

func (r *permissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

// securableInfo derives the Unity Catalog securable type and full name from the
// catalog/schema/table dimensions.
func securableInfo(catalog, schema, table string) (securableType, fullName string) {
	switch {
	case table != "":
		return "table", catalog + "." + schema + "." + table
	case schema != "":
		return "schema", catalog + "." + schema
	default:
		return "catalog", catalog
	}
}

// permissionsID builds the composite ID: catalog[.schema[.table]]#principal.
func permissionsID(catalog, schema, table, principal string) string {
	name := catalog
	if schema != "" {
		name += "." + schema
	}
	if table != "" {
		name += "." + table
	}
	return name + "#" + principal
}

// parsePermissionsID splits the composite ID back into its components.
func parsePermissionsID(id string) (catalog, schema, table, principal string, err error) {
	hashParts := strings.SplitN(id, "#", 2)
	if len(hashParts) != 2 || hashParts[0] == "" || hashParts[1] == "" {
		return "", "", "", "", fmt.Errorf("invalid permissions ID %q: expected format catalog[.schema[.table]]#principal", id)
	}
	nameParts := strings.Split(hashParts[0], ".")
	catalog = nameParts[0]
	if len(nameParts) >= 2 {
		schema = nameParts[1]
	}
	if len(nameParts) >= 3 {
		table = nameParts[2]
	}
	principal = hashParts[1]
	return
}

// modelToStrings extracts the catalog/schema/table/principal/privileges strings
// from the Terraform model. Optional fields (schema, table) return "" when null.
func modelToStrings(ctx context.Context, model permissionsResourceModel) (catalog, schema, table, principal string, privileges []string) {
	catalog = model.Catalog.ValueString()
	if !model.Schema.IsNull() {
		schema = model.Schema.ValueString()
	}
	if !model.Table.IsNull() {
		table = model.Table.ValueString()
	}
	principal = model.Principal.ValueString()
	model.Privileges.ElementsAs(ctx, &privileges, false)
	return
}

// setPrivileges sets the privileges list on the model from a Go slice.
func setPrivileges(ctx context.Context, model *permissionsResourceModel, privileges []string) error {
	list, diags := types.ListValueFrom(ctx, types.StringType, privileges)
	if diags.HasError() {
		return fmt.Errorf("failed to build privileges list: %v", diags.Errors())
	}
	model.Privileges = list
	return nil
}

func (r *permissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan permissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalog, schema, table, principal, privileges := modelToStrings(ctx, plan)
	securableType, fullName := securableInfo(catalog, schema, table)

	updateReq := &client.UpdatePermissionsRequest{
		Changes: []client.PrivilegeChange{
			{
				Principal: principal,
				Add:       privileges,
			},
		},
	}

	if err := r.client.UpdatePermissions(ctx, securableType, fullName, updateReq); err != nil {
		resp.Diagnostics.AddError("Failed to create Unity Catalog permissions", err.Error())
		return
	}

	plan.Id = types.StringValue(permissionsID(catalog, schema, table, principal))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *permissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state permissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalog, schema, table, principal, _ := modelToStrings(ctx, state)
	securableType, fullName := securableInfo(catalog, schema, table)

	result, err := r.client.GetPermissions(ctx, securableType, fullName)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Unity Catalog permissions", err.Error())
		return
	}

	var currentPrivileges []string
	for _, pa := range result.PrivilegeAssignments {
		if pa.Principal == principal {
			currentPrivileges = pa.Privileges
			break
		}
	}

	if currentPrivileges == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if err := setPrivileges(ctx, &state, currentPrivileges); err != nil {
		resp.Diagnostics.AddError("Failed to set privileges in state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *permissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan permissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state permissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, _, _, _, oldPrivs := modelToStrings(ctx, state)
	catalog, schema, table, principal, newPrivs := modelToStrings(ctx, plan)
	securableType, fullName := securableInfo(catalog, schema, table)

	toAdd := stringDifference(newPrivs, oldPrivs)
	toRemove := stringDifference(oldPrivs, newPrivs)

	if len(toAdd) > 0 || len(toRemove) > 0 {
		updateReq := &client.UpdatePermissionsRequest{
			Changes: []client.PrivilegeChange{
				{
					Principal: principal,
					Add:       toAdd,
					Remove:    toRemove,
				},
			},
		}
		if err := r.client.UpdatePermissions(ctx, securableType, fullName, updateReq); err != nil {
			resp.Diagnostics.AddError("Failed to update Unity Catalog permissions", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *permissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state permissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalog, schema, table, principal, privileges := modelToStrings(ctx, state)
	securableType, fullName := securableInfo(catalog, schema, table)

	if len(privileges) == 0 {
		return
	}

	updateReq := &client.UpdatePermissionsRequest{
		Changes: []client.PrivilegeChange{
			{
				Principal: principal,
				Remove:    privileges,
			},
		},
	}

	if err := r.client.UpdatePermissions(ctx, securableType, fullName, updateReq); err != nil {
		resp.Diagnostics.AddError("Failed to delete Unity Catalog permissions", err.Error())
		return
	}
}

func (r *permissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	catalog, schema, table, principal, err := parsePermissionsID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Import Identifier", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("catalog"), catalog)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("principal"), principal)...)
	if schema != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("schema"), schema)...)
	}
	if table != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("table"), table)...)
	}
}

// stringDifference returns the elements of a that are not in b.
func stringDifference(a, b []string) []string {
	bset := make(map[string]struct{}, len(b))
	for _, s := range b {
		bset[s] = struct{}{}
	}
	var diff []string
	for _, s := range a {
		if _, ok := bset[s]; !ok {
			diff = append(diff, s)
		}
	}
	return diff
}
