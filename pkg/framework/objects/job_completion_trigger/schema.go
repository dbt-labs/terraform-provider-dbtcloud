package job_completion_trigger

import (
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var resourceSchema = resource_schema.Schema{
	Description: "Manages a job completion trigger, which fires a downstream job when an upstream job reaches certain statuses. Use this resource instead of the `job_completion_trigger` block on `dbtcloud_job` to break circular dependency chains.",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of the downstream job (same as `job_id`).",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"job_id": resource_schema.Int64Attribute{
			Required:    true,
			Description: "The ID of the downstream job that will be triggered.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"trigger_job_id": resource_schema.Int64Attribute{
			Required:    true,
			Description: "The ID of the upstream job whose completion fires this trigger.",
		},
		"project_id": resource_schema.Int64Attribute{
			Required:    true,
			Description: "The dbt Cloud project ID.",
		},
		"statuses": resource_schema.SetAttribute{
			Required:    true,
			ElementType: types.StringType,
			Description: "The set of job completion statuses that trigger the downstream job. Valid values: `success`, `error`, `canceled`.",
		},
	},
}
