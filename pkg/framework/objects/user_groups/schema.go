package user_groups

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var resourceSchema = resource_schema.Schema{
	Description: helper.DocString(
		`Assigns a set of dbt Cloud groups to a given User ID. 

~> If additional groups were assigned manually in dbt Cloud, they will be removed. The full list of groups need to be provided as config.
~> This resource does not currently support deletion (e.g. a deleted resource will stay as-is in dbt Cloud).
This is intentional in order to prevent accidental deletion of all users groups assigned to a user.
If you would like a different behavior, please open an issue on GitHub. To remove all groups for a user, set "group_ids" to the empty set "[]".`,
	),
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource. It is the same as the user_id.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"user_id": resource_schema.Int64Attribute{
			Description: "The internal ID of a dbt Cloud user.",
			Required:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"group_ids": schema.SetAttribute{
			Description: "IDs of the groups to assign to the user. If additional groups were assigned manually in dbt Cloud, they will be removed.",
			ElementType: types.Int64Type,
			Required:    true,
		},
	},
}

var datasourceSchema = datasource_schema.Schema{
	Description: helper.DocString(
		`Gets information about a specific dbt Cloud user's groups.`,
	),
	Attributes: map[string]datasource_schema.Attribute{
		"id": datasource_schema.StringAttribute{
			Description: "The ID of this resource. It is the same as the user_id.",
			Computed:    true,
		},
		"user_id": datasource_schema.Int64Attribute{
			Description: "The internal ID of a dbt Cloud user.",
			Required:    true,
		},
		"group_ids": datasource_schema.SetAttribute{
			Description: "IDs of the groups assigned to the user.",
			ElementType: types.Int64Type,
			Computed:    true,
		},
	},
}
