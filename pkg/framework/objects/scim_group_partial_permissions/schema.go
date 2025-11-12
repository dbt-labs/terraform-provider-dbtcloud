package scim_group_partial_permissions

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

func (r *scimGroupPartialPermissionsResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: helper.DocString(
			`Provide a partial set of permissions for an externally managed group (e.g., SCIM, manually created). 
			This resource ONLY manages a subset of permissions and never creates or deletes groups.
			
			This is designed for federated permission management where a platform team sets global permissions 
			and individual teams manage their own project-specific permissions for the same group.
			
			⚠️  **Important Differences:**
			- ~~~dbt_cloud_group~~~: Creates group and fully manages ALL permissions (single Terraform workspace)
			- ~~~dbt_cloud_group_partial_permissions~~~: Creates group and manages PARTIAL permissions (multiple Terraform workspaces)
			- ~~~dbt_cloud_scim_group_permissions~~~: Externally-managed group, fully manages ALL permissions (replaces all permissions)
			- ~~~dbt_cloud_scim_group_partial_permissions~~~: Externally-managed group, manages PARTIAL permissions (adds/removes only specified permissions)

			**Use Case:**
			- Group exists in external identity provider (e.g., Okta, Azure AD) and syncs via SCIM
			- Platform team manages base permissions (e.g., account-level access)
			- Individual teams manage their own project-specific permissions
			- Multiple Terraform workspaces can safely manage different permissions for the same group

			⚠️  Do not mix different resource types for the same group:
			- Don't use ~~~dbt_cloud_scim_group_permissions~~~ (full permissions) with ~~~dbt_cloud_scim_group_partial_permissions~~~ (partial permissions)
			- Don't use ~~~dbt_cloud_group~~~ or ~~~dbt_cloud_group_partial_permissions~~~ for externally managed groups

			The resource currently requires a Service Token with Account Admin access.

			**Behavior:**
			- When creating: Adds specified permissions to the existing group (if not already present)
			- When updating: Adds new permissions and removes old permissions from this resource
			- When deleting: Removes only the permissions managed by this resource (group and other permissions remain)
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
				Description: "The ID of the existing group to manage partial permissions for. This group must already exist and is typically from an external identity provider synced via SCIM.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"permissions": schema.SetNestedAttribute{
				Description: "Partial set of permissions to apply to the group. These permissions will be added to any existing permissions. Other permissions on the group will not be affected.",
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
