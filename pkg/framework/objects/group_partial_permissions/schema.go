package group_partial_permissions

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *groupPartialPermissionsResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: helper.DocString(
			`Provide a partial set of permissions for a group. This is different from ~~~dbt_cloud_group~~~ as it allows to have multiple resources updating the same dbt Cloud group and is useful for companies managing a single dbt Cloud Account configuration from different Terraform projects/workspaces.

			If a company uses only one Terraform project/workspace to manage all their dbt Cloud Account config, it is recommended to use ~~~dbt_cloud_group~~~ instead of ~~~dbt_cloud_group_partial_permissions~~~.

			~> This is currently an experimental resource and any feedback is welcome in the GitHub repository.

			The resource currently requires a Service Token with Account Admin access.

			The current behavior of the resource is the following:

			- when using ~~~dbt_cloud_group_partial_permissions~~~, don't use ~~~dbt_cloud_group~~~ for the same group in any other project/workspace. Otherwise, the behavior is undefined and partial permissions might be removed.
			- when defining a new ~~~dbt_cloud_group_partial_permissions~~~
			  - if the group doesn't exist with the given ~~~name~~~, it will be created
			  - if a group exists with the given ~~~name~~~, permissions will be added in the dbt Cloud group if they are not present yet
			- in a given Terraform project/workspace, avoid having different ~~~dbt_cloud_group_partial_permissions~~~ for the same group name to prevent sync issues. Add all the permissions in the same resource. 
			- all resources for the same group name need to have the same values for ~~~assign_by_default~~~ and ~~~sso_mapping_groups~~~. Those fields are not considered "partial". (Please raise feedback in GitHub if you think that ~~~sso_mapping_groups~~~ should be "partial" as well)
			- when a resource is updated, the dbt Cloud group will be updated accordingly, removing and adding permissions
			- when the resource is deleted/destroyed, if the resulting permission sets is empty, the group will be deleted ; otherwise, the group will be updated, removing the permissions from the deleted resource
			`,
		),
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the group",
				// this is used so that we don't show that ID is going to change
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the group. This is used to identify an existing group",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"assign_by_default": schema.BoolAttribute{
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the group will be assigned by default to users. The value needs to be the same for all partial permissions for the same group.",
			},
			"sso_mapping_groups": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Mapping groups from the IdP. At the moment the complete list needs to be provided in each partial permission for the same group.",
			},
			"group_permissions": schema.SetNestedAttribute{
				Description: "Partial permissions for the group. Those permissions will be added/removed when config is added/removed.",
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
					},
				},
				Optional: true,
			},
		},
	}
}
