package service_token

import (
	"context"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &serviceTokenResource{}
	_ resource.ResourceWithConfigure   = &serviceTokenResource{}
	_ resource.ResourceWithImportState = &serviceTokenResource{}
)

func ServiceTokenResource() resource.Resource {
	return &serviceTokenResource{}
}

type serviceTokenResource struct {
	client *dbt_cloud.Client
}

// Metadata implements resource.Resource.
func (st *serviceTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_token"
}

// Configure implements resource.ResourceWithConfigure.
func (st *serviceTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	switch c := req.ProviderData.(type) {
	case nil: // do nothing
	case *dbt_cloud.Client:
		st.client = c
	default:
		resp.Diagnostics.AddError("Missing client", "A client is required to configure the service token resource")
	}
}

// Schema implements resource.Resource.
func (st *serviceTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the service token",
			},
			"uid": schema.StringAttribute{
				Description: "Service token UID (part of the token)",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Service token name",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token_string": schema.StringAttribute{
				Description: "Service token secret value (only accessible on creation))",
				Computed:    true,
				Sensitive:   true,
			},
			"state": schema.Int64Attribute{
				Description: "Service token state (1 is active, 2 is inactive)",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"service_token_permissions": schema.SetNestedBlock{
				Description: "Permissions set for the service token",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"permission_set": schema.StringAttribute{
							Description: "Set of permissions to apply",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(dbt_cloud.PermissionSets...),
							},
						},
						"all_projects": schema.BoolAttribute{
							Description: "Whether or not to apply this permission to all projects for this service token",
							Required:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
						},
						// TODO(cwalden): Would this be better as a Set of Int64?
						// TODO(cwalden): Add a validator to ensure that the project ID is set if `all_projects` is false
						"project_id": schema.Int64Attribute{
							Description: "Project ID to apply this permission to for this service token",
							Optional:    true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
						},
						// TODO(cwalden): Add validator to ensure that this is configurable for the given `permission_set`
						"writable_environment_categories": schema.SetAttribute{
							Description: helper.DocString(
								`What types of environments to apply Write permissions to.
								Even if Write access is restricted to some environment types, the permission set will have Read access to all environments.
								The values allowed are ~~~all~~~, ~~~development~~~, ~~~staging~~~, ~~~production~~~ and ~~~other~~~.
								Not setting a value is the same as selecting ~~~all~~~.
								Not all permission sets support environment level write settings, only ~~~analyst~~~, ~~~database_admin~~~, ~~~developer~~~, ~~~git_admin~~~ and ~~~team_admin~~~.`,
							),
							Optional: true,
							Computed: true,
							// Default:  helper.EmptySetDefault(types.StringType),
							Default: setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{
								types.StringValue("all"),
							})),
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Set{
								setvalidator.ValueStringsAre(
									stringvalidator.OneOf(dbt_cloud.EnvironmentCategories...),
								),
							},
						},
					},
				},
			},
		},
	}
}

// Read implements resource.Resource.
func (st *serviceTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state ServiceTokenResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	svcTokID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to convert the service token ID to an integer", err.Error())
		return
	}

	svcTok, err := st.client.GetServiceToken(svcTokID)

	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The service token was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting the service token", err.Error())
		return
	}

	state.ID = types.StringValue(strconv.Itoa(*svcTok.ID))
	state.UID = types.StringValue(svcTok.UID)
	state.Name = types.StringValue(svcTok.Name)
	state.State = types.Int64Value(int64(svcTok.State))

	svcTokPerms, err := st.client.GetServiceTokenPermissions(int(svcTokID))

	if err != nil {
		resp.Diagnostics.AddError("Error getting the service token permissions", err.Error())
		return
	}

	perms, diags := ConvertServiceTokenPermissionDataToModel(ctx, *svcTokPerms)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ServiceTokenPermissions = perms

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

// Create implements resource.Resource.
func (st *serviceTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan ServiceTokenResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	state := plan.State.ValueInt64()

	createdSrvTok, err := st.client.CreateServiceToken(name, int(state))
	if err != nil {
		resp.Diagnostics.AddError("Unable to create the service token", err.Error())
		return
	}

	if createdSrvTok == nil || createdSrvTok.ID == nil {
		resp.Diagnostics.AddError("Error creating the service token", "The created service token or its ID is null")
		return
	}

	srvTokPermissions, diags := ConvertServiceTokenPermissionModelToData(ctx, plan.ServiceTokenPermissions, *createdSrvTok.ID, st.client.AccountID)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	updatedSvcTokPerms, err := st.client.UpdateServiceTokenPermissions(*createdSrvTok.ID, srvTokPermissions)
	if err != nil {
		resp.Diagnostics.AddError("Unable to assign permissions to the service token", err.Error())
		return
	}

	perms, diags := ConvertServiceTokenPermissionDataToModel(ctx, *updatedSvcTokPerms)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(*createdSrvTok.ID))
	plan.UID = types.StringValue(createdSrvTok.UID)
	plan.Name = types.StringValue(createdSrvTok.Name)
	plan.State = types.Int64Value(int64(createdSrvTok.State))
	plan.TokenString = types.StringValue(*createdSrvTok.TokenString)
	plan.ServiceTokenPermissions = perms

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

}

// Update implements resource.Resource.
func (st *serviceTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Operation not supported",
		"Service tokens cannot be updated after creation. To modify a service token, you must delete and recreate it.",
	)
}

// Delete implements resource.Resource.
func (st *serviceTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServiceTokenResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	svcTokID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to convert the service token ID to an integer", err.Error())
		return
	}

	if _, err := st.client.DeleteServiceToken(svcTokID); err != nil {
		resp.Diagnostics.AddError("Unable to delete the service token", err.Error())
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
func (st *serviceTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
