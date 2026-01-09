package job_completion_trigger

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ResourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Refers to a job completion trigger condition for a job. This is separate from the dbtcloud_job resource to allow for job chaining without circular dependencies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the resource (same as the job_id)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"job_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the job to add the trigger to (downstream job)",
				PlanModifiers: []planmodifier.Int64{
					// If job_id changes, we must replace the resource (it's a different job)
					// Actually, no, we can just delete the old one and create new one.
				},
			},
			"trigger_job_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the job that triggers this job (upstream job)",
			},
			"project_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the project where the trigger job is running in",
			},
			"statuses": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of statuses to trigger the job on. Possible values are `success`, `error` and `canceled`.",
			},
		},
	}
}
