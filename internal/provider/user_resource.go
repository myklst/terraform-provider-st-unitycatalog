package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/myklst/terraform-provider-st-unitycatalog/client"
)

var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

// NewUserResource creates a new Unity Catalog user resource.
func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *client.Client
}

// emailModel maps a single SCIM2 email entry to/from Terraform state.
type emailModel struct {
	Value   types.String `tfsdk:"value"`
	Primary types.Bool   `tfsdk:"primary"`
}

// userResourceModel maps the SCIM2 user attributes managed by this provider to
// Terraform state. The id is server-assigned and not user configurable, but it
// is echoed back in PUT requests (see Update).
type userResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Schemas     types.List   `tfsdk:"schemas"`
	DisplayName types.String `tfsdk:"display_name"`
	Emails      types.List   `tfsdk:"emails"`
	Active      types.Bool   `tfsdk:"active"`
}

// emailAttrTypes is the attr.Type mapping for the email nested object.
var emailAttrTypes = map[string]attr.Type{
	"value":   types.StringType,
	"primary": types.BoolType,
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a user in Unity Catalog via the SCIM2 Users API " +
			"(/api/1.0/unity-control/scim2/Users).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The SCIM2 identifier of the user, assigned by the server. " +
					"This attribute is not user configurable but is sent back to the server " +
					"in PUT requests as required by the SCIM2 API.",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schemas": schema.ListAttribute{
				Description: "The SCIM2 schema URNs for the user resource, e.g. " +
					"[\"urn:ietf:params:scim:schemas:core:2.0:User\"].",
				ElementType: types.StringType,
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the user.",
				Required:    true,
			},
			"emails": schema.ListNestedAttribute{
				Description: "The list of user emails. Changing this value forces replacement " +
					"of the resource, because emails are immutable in Unity Catalog.",
				Required: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"value": schema.StringAttribute{
							Description: "The email address.",
							Required:    true,
						},
						"primary": schema.BoolAttribute{
							Description: "Whether this is the primary email. Defaults to false.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
					},
				},
			},
			"active": schema.BoolAttribute{
				Description: "Whether the user is active.",
				Required:    true,
			},
		},
	}
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

// modelToUser converts the Terraform model into a SCIM2 user payload. When
// includeID is true the server-assigned id is carried over so it can be sent in
// a PUT request body.
func (r *userResource) modelToUser(ctx context.Context, model userResourceModel, includeID bool) (*client.User, diag.Diagnostics) {
	var diags diag.Diagnostics

	var schemas []string
	d := model.Schemas.ElementsAs(ctx, &schemas, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	var emails []emailModel
	d = model.Emails.ElementsAs(ctx, &emails, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	scimEmails := make([]client.Email, 0, len(emails))
	for _, e := range emails {
		scimEmails = append(scimEmails, client.Email{
			Value:   e.Value.ValueString(),
			Primary: e.Primary.ValueBool(),
		})
	}

	u := &client.User{
		Schemas:     schemas,
		DisplayName: model.DisplayName.ValueString(),
		Emails:      scimEmails,
		Active:      model.Active.ValueBool(),
	}
	if includeID {
		u.ID = model.Id.ValueString()
	}
	return u, diags
}

// userToModel converts a SCIM2 user returned by the server into the Terraform
// model used for state.
func (r *userResource) userToModel(ctx context.Context, u *client.User) (userResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := userResourceModel{
		Id:          types.StringValue(u.ID),
		DisplayName: types.StringValue(u.DisplayName),
		Active:      types.BoolValue(u.Active),
	}

	schemasList, d := types.ListValueFrom(ctx, types.StringType, u.Schemas)
	diags.Append(d...)
	if diags.HasError() {
		return model, diags
	}
	model.Schemas = schemasList

	emails := make([]emailModel, 0, len(u.Emails))
	for _, e := range u.Emails {
		emails = append(emails, emailModel{
			Value:   types.StringValue(e.Value),
			Primary: types.BoolValue(e.Primary),
		})
	}

	emailsList, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: emailAttrTypes}, emails)
	diags.Append(d...)
	if diags.HasError() {
		return model, diags
	}
	model.Emails = emailsList

	return model, diags
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, diags := r.modelToUser(ctx, plan, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateUser(ctx, u)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Unity Catalog user", err.Error())
		return
	}

	state, diags := r.userToModel(ctx, created)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, err := r.client.GetUser(ctx, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Unity Catalog user", err.Error())
		return
	}

	updated, diags := r.userToModel(ctx, u)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &updated)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The id is computed and not user-configurable, but the SCIM2 PUT endpoint
	// requires it to be present in the request body. Carry it over from the
	// current state.
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Id = state.Id

	u, diags := r.modelToUser(ctx, plan, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateUser(ctx, u)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Unity Catalog user", err.Error())
		return
	}

	result, diags := r.userToModel(ctx, updated)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteUser(ctx, state.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete Unity Catalog user", err.Error())
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
