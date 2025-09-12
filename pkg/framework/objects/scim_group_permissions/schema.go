package scim_group_permissions

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *scimGroupPermissionsResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: helper.DocString(
			`Manage permissions for groups that are externally managed (e.g., SCIM, manually created). 
			This resource ONLY manages permissions and never creates or deletes groups.
			
			⚠️  Do not use this resource alongside ~~~dbt_cloud_group~~~ or ~~~dbt_cloud_group_partial_permissions~~~ 
			for the same group to avoid permission conflicts.
			
			This resource is ideal for SCIM-managed environments where groups exist in your identity 
			provider and are synced to dbt Cloud, but you want to manage permissions via Terraform.

			**Use Case Guidelines:**
			- Use ~~~dbt_cloud_group~~~ when Terraform creates and fully manages the group
			- Use ~~~dbt_cloud_group_partial_permissions~~~ when multiple Terraform workspaces manage the same Terraform-created group  
			- Use ~~~dbt_cloud_scim_group_permissions~~~ when the group is externally managed (e.g., SCIM, manual creation) and you only want to manage permissions

			The resource currently requires a Service Token with Account Admin access.
			`,
		),
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the group (same as group_id)",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"group_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the existing group to manage permissions for. This group must already exist.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"permissions": schema.SetNestedAttribute{
				Description: "Set of permissions to apply to the group. This will replace all existing permissions for the group.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"permission_set": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf(dbt_cloud.PermissionSets...),
							},
							Description: "Set of permissions to apply. The permissions allowed are the same as the ones for the `dbtcloud_group` resource.",
						},
						"project_id": schema.Int64Attribute{
							Optional:    true,
							Description: "Project ID to apply this permission to for this group.",
						},
						"all_projects": schema.BoolAttribute{
							Required:    true,
							Description: "Whether access should be provided for all projects or not.",
						},
						"writable_environment_categories": schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: helper.DocString(
								`What types of environments to apply Write permissions to. 
								Even if Write access is restricted to some environment types, the permission set will have Read access to all environments. 
								The values allowed are ~~~all~~~, ~~~development~~~, ~~~staging~~~, ~~~production~~~ and ~~~other~~~. 
								Not setting a value is the same as selecting ~~~all~~~. 
								Not all permission sets support environment level write settings, only ~~~analyst~~~, ~~~database_admin~~~, ~~~developer~~~, ~~~git_admin~~~ and ~~~team_admin~~~.`,
							),
						},
					},
				},
				Optional: true,
			},
		},
	}
}
