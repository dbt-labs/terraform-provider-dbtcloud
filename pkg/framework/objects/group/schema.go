package group

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *groupResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resource_schema.Schema{
		Description: helper.DocString(
			`Provide a complete set of permissions for a group. This is different from ~~~dbt_cloud_partial_group_permissions~~~.

			With this resource type only one resource can be used to manage the permissions for a given group.
			`,
		),
		Attributes: map[string]resource_schema.Attribute{
			"id": resource_schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the group",
				// this is used so that we don't show that ID is going to change
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": resource_schema.StringAttribute{
				Required:    true,
				Description: "The name of the group. This is used to identify an existing group",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"assign_by_default": resource_schema.BoolAttribute{
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the group will be assigned by default to users. The value needs to be the same for all partial permissions for the same group.",
			},
			"sso_mapping_groups": resource_schema.SetAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
				Default:     helper.EmptySetDefault(types.StringType),
				Description: "Mapping groups from the IdP. At the moment the complete list needs to be provided in each partial permission for the same group.",
			},
			// "group_permissions": resource_schema.SetNestedAttribute{
			// 	Description: "Partial permissions for the group. Those permissions will be added/removed when config is added/removed.",
			// 	NestedObject: resource_schema.NestedAttributeObject{
			// 		Attributes: map[string]resource_schema.Attribute{
			// 			"permission_set": resource_schema.StringAttribute{
			// 				Required: true,
			// 				Validators: []validator.String{
			// 					stringvalidator.OneOf(dbt_cloud.PermissionSets...),
			// 				},
			// 				Description: "Set of permissions to apply. The permissions allowed are the same as the ones for the `dbtcloud_group` resource.",
			// 			},
			// 			"project_id": resource_schema.Int64Attribute{
			// 				Optional:    true,
			// 				Description: "Project ID to apply this permission to for this group.",
			// 			},
			// 			"all_projects": resource_schema.BoolAttribute{
			// 				Required:    true,
			// 				Description: "Whether access should be provided for all projects or not.",
			// 			},
			// 			"writable_environment_categories": resource_schema.SetAttribute{
			// 				ElementType: types.StringType,
			// 				Optional:    true,
			// 				Description: helper.DocString(
			// 					`What types of environments to apply Write permissions to.
			// 					Even if Write access is restricted to some environment types, the permission set will have Read access to all environments.
			// 					The values allowed are ~~~all~~~, ~~~development~~~, ~~~staging~~~, ~~~production~~~ and ~~~other~~~.
			// 					Not setting a value is the same as selecting ~~~all~~~.
			// 					Not all permission sets support environment level write settings, only ~~~analyst~~~, ~~~database_admin~~~, ~~~developer~~~, ~~~git_admin~~~ and ~~~team_admin~~~.`,
			// 				),
			// 			},
			// 		},
			// 	},
			// 	Optional: true,
			// },
		},
		// For now we use a Block to move from SDKv2 to PLugin Framework, but we might change to a SetAttribute in the future, using the code from above
		Blocks: map[string]resource_schema.Block{
			"group_permissions": resource_schema.SetNestedBlock{
				Description: "Partial permissions for the group. Those permissions will be added/removed when config is added/removed.",
				NestedObject: resource_schema.NestedBlockObject{
					Attributes: map[string]resource_schema.Attribute{
						"permission_set": resource_schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf(dbt_cloud.PermissionSets...),
							},
							Description: "Set of permissions to apply. The permissions allowed are the same as the ones for the `dbtcloud_group` resource.",
						},
						"project_id": resource_schema.Int64Attribute{
							Optional:    true,
							Description: "Project ID to apply this permission to for this group.",
						},
						"all_projects": resource_schema.BoolAttribute{
							Required:    true,
							Description: "Whether access should be provided for all projects or not.",
						},
						"writable_environment_categories": resource_schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
							Default:     helper.EmptySetDefault(types.StringType),
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
			},
		},
	}
}

func (d *groupDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = datasource_schema.Schema{
		Description: "Retrieve group details",
		Attributes: map[string]datasource_schema.Attribute{
			"group_id": datasource_schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the group",
			},
			"id": datasource_schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of this resource",
			},
			"name": datasource_schema.StringAttribute{
				Computed:    true,
				Description: "Group name",
			},
			"assign_by_default": datasource_schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the group will be assigned by default to users. The value needs to be the same for all partial permissions for the same group.",
			},
			"sso_mapping_groups": datasource_schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "SSO mapping group names for this group",
			},
			"group_permissions": datasource_schema.SetNestedAttribute{
				Computed:    true,
				Description: "Partial permissions for the group. Those permissions will be added/removed when config is added/removed.",
				NestedObject: datasource_schema.NestedAttributeObject{
					Attributes: map[string]datasource_schema.Attribute{
						"permission_set": datasource_schema.StringAttribute{
							Computed:    true,
							Description: "Set of permissions to apply. The permissions allowed are the same as the ones for the `dbtcloud_group` resource.",
						},
						"project_id": datasource_schema.Int64Attribute{
							Computed:    true,
							Description: "Project ID to apply this permission to for this group.",
						},
						"all_projects": datasource_schema.BoolAttribute{
							Computed:    true,
							Description: "Whether access should be provided for all projects or not.",
						},
						"writable_environment_categories": datasource_schema.SetAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "What types of environments to apply Write permissions to.",
						},
					},
				},
			},
		},
	}
}

var groupsDataSourceSchema = datasource_schema.Schema{
	Description: "Retrieve all groups in the account with optional filtering",
	Attributes: map[string]datasource_schema.Attribute{
		"name": datasource_schema.StringAttribute{
			Optional:    true,
			Description: "Filter groups by exact name match",
		},
		"name_contains": datasource_schema.StringAttribute{
			Optional:    true,
			Description: "Filter groups by partial name match (case insensitive)",
		},
		"state": datasource_schema.StringAttribute{
			Optional:    true,
			Description: "Filter groups by state. Accepts both string and integer formats: 'active'/'1' for active resources, 'deleted'/'2' for deleted resources, 'all' for all resources. Defaults to active groups only if not specified.",
			Validators: []validator.String{
				stringvalidator.OneOf("active", "1", "deleted", "2", "all"),
			},
		},
		"groups": datasource_schema.SetNestedAttribute{
			Computed:    true,
			Description: "Set of groups in the account",
			NestedObject: datasource_schema.NestedAttributeObject{
				Attributes: map[string]datasource_schema.Attribute{
					"id": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The ID of the group",
					},
					"name": datasource_schema.StringAttribute{
						Computed:    true,
						Description: "Group name",
					},
					"state": datasource_schema.Int64Attribute{
						Computed:    true,
						Description: "The state of the group (1=active, 2=deleted)",
					},
					"assign_by_default": datasource_schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the group will be assigned by default to users",
					},
					"scim_managed": datasource_schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the group is managed by SCIM",
					},
					"sso_mapping_groups": datasource_schema.SetAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "SSO mapping group names for this group",
					},
				},
			},
		},
	},
}
