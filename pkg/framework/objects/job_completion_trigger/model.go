package job_completion_trigger

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type JobCompletionTriggerResourceModel struct {
	ID           types.String `tfsdk:"id"`
	JobID        types.Int64  `tfsdk:"job_id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	TriggerJobID types.Int64  `tfsdk:"trigger_job_id"`
	Statuses     types.Set    `tfsdk:"statuses"`
}

