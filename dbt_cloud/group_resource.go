package dbt_cloud

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &groupResource{}
	_ resource.ResourceWithConfigure = &groupResource{}
)

func NewGroupResource() resource.Resource {
	return &groupResource{}
}

type groupResource struct {
	client *Client
}

type groupPermissionModel struct {
	PermissionSet types.String `tfsdk:"permission_set"`
	ProjectID     types.Int64  `tfsdk:"project_id"`
	AllProjects   types.Bool   `tfsdk:"all_projects"`
}

type groupResourceModel struct {
	ID               types.Int64            `tfsdk:"id"`
	Name             types.String           `tfsdk:"name"`
	IsActive         types.Bool             `tfsdk:"is_active"`
	AssignByDefault  types.Bool             `tfsdk:"assign_by_default"`
	SSOMappingGroups []types.String         `tfsdk:"sso_mapping_groups"`
	GroupPermissions []groupPermissionModel `tfsdk:"group_permissions"`
	LastUpdated      types.String           `tfsdk:"last_updated"`
}

func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Group name",
			},
			"is_active": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the group is active. (Default is true)",
			},
			"assign_by_default": schema.BoolAttribute{
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether or not to assign this group to users by default. (Default is false)",
			},
			"sso_mapping_groups": schema.ListAttribute{
				Optional:    true,
				Description: "SSO mapping group names for this group",
				ElementType: types.StringType,
			},
			"group_permissions": schema.SetNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{

						"permission_set": schema.StringAttribute{
							Required:    true,
							Description: "Set of permissions to apply",
							Validators: []validator.String{
								stringvalidator.OneOf([]string{
									"owner",
									"member",
									"account_admin",
									"admin",
									"database_admin",
									"git_admin",
									"team_admin",
									"job_admin",
									"job_viewer",
									"analyst",
									"developer",
									"stakeholder",
									"readonly",
									"project_creator",
									"account_viewer",
									"metadata_only",
									"webhooks_only",
								}...),
							},
						},
						"project_id": schema.Int64Attribute{
							Optional:    true,
							Description: "Project ID to apply this permission to for this group",
						},
						"all_projects": schema.BoolAttribute{
							Required:    true,
							Description: "Whether or not to apply this permission to all projects for this group",
						},
					},
				},
			},
		},
	}
}

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ssoMappingGroups := make([]string, len(plan.SSOMappingGroups))
	for i, _ := range plan.SSOMappingGroups {
		ssoMappingGroups[i] = string(plan.SSOMappingGroups[i].ValueString())
	}

	group, err := r.client.CreateGroup(string(plan.Name.ValueString()), bool(plan.AssignByDefault.ValueBool()), ssoMappingGroups)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating group",
			"Could not create group, unexpected error: "+err.Error(),
		)
		return
	}

	groupPermissions := make([]GroupPermission, len(plan.GroupPermissions))
	for i, permission := range plan.GroupPermissions {
		groupPermission := GroupPermission{
			GroupID:     *group.ID,
			AccountID:   group.AccountID,
			Set:         string(permission.PermissionSet.ValueString()),
			ProjectID:   int(permission.ProjectID.ValueInt64()),
			AllProjects: bool(permission.AllProjects.ValueBool()),
		}
		groupPermissions[i] = groupPermission
	}

	_, err = r.client.UpdateGroupPermissions(*group.ID, groupPermissions)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating group permissions",
			"Could not update group permissions, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.Int64Value(int64(*group.ID))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}
