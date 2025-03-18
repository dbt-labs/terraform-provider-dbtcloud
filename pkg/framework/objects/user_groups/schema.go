package user_groups

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var resourceSchema = resource_schema.Schema {
	Description: helper.DocString(
		`Assigns a set of dbt Cloud groups to a given User ID. 

~> If additional groups were assigned manually in dbt Cloud, they will be removed. The full list of groups need to be provided as config.
		
~> This resource does not currently support deletion (e.g. a deleted resource will stay as-is in dbt Cloud).
This is intentional in order to prevent accidental deletion of all users groups assigned to a user.
If you would like a different behavior, please open an issue on GitHub. To remove all groups for a user, set "group_ids" to the empty set "[]".`,
	),
	Attributes: map[string] resource_schema.Attribute{
		"id": resource_schema.Int64Attribute{
			Computed: true,
			Description: "The ID of the user group resource.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"user_id": resource_schema.Int64Attribute{
			Description: "The internal ID of a dbt Cloud user.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"group_ids": schema.SetAttribute{
			ElementType: types.Int64Type,
			Required:    true,
			Description: "IDs of the groups to assign to the user. If additional groups were assigned manually in dbt Cloud, they will be removed.",
		},
	},
}
