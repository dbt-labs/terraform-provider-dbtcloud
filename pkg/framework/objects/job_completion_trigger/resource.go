package job_completion_trigger

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &jobCompletionTriggerResource{}
	_ resource.ResourceWithConfigure   = &jobCompletionTriggerResource{}
	_ resource.ResourceWithImportState = &jobCompletionTriggerResource{}
)

func JobCompletionTriggerResource() resource.Resource {
	return &jobCompletionTriggerResource{}
}

type jobCompletionTriggerResource struct {
	client *dbt_cloud.Client
}

func (r *jobCompletionTriggerResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_job_completion_trigger"
}

func (r *jobCompletionTriggerResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resourceSchema
}

func (r *jobCompletionTriggerResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*dbt_cloud.Client)
}

func (r *jobCompletionTriggerResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan JobCompletionTriggerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := plan.JobID.ValueInt64()
	job, err := r.client.GetJob(strconv.FormatInt(jobID, 10))
	if err != nil {
		resp.Diagnostics.AddError("Error fetching job", err.Error())
		return
	}

	statuses, diags := buildStatusInts(plan.Statuses)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	job.JobCompletionTrigger = &dbt_cloud.JobCompletionTrigger{
		Condition: dbt_cloud.JobCompletionTriggerCondition{
			JobID:     int(plan.TriggerJobID.ValueInt64()),
			ProjectID: int(plan.ProjectID.ValueInt64()),
			Statuses:  statuses,
		},
	}

	updatedJob, err := r.client.UpdateJob(strconv.FormatInt(jobID, 10), *job)
	if err != nil {
		resp.Diagnostics.AddError("Error setting job completion trigger", err.Error())
		return
	}

	plan.ID = types.Int64Value(int64(*updatedJob.ID))
	plan.JobID = types.Int64Value(int64(*updatedJob.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *jobCompletionTriggerResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state JobCompletionTriggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := state.JobID.ValueInt64()
	job, err := r.client.GetJob(strconv.FormatInt(jobID, 10))
	if err != nil {
		if helper.HandleResourceNotFound(ctx, err, &resp.Diagnostics, &resp.State, "job_completion_trigger") {
			return
		}
		resp.Diagnostics.AddError("Error fetching job", err.Error())
		return
	}

	if job.JobCompletionTrigger == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	cond := job.JobCompletionTrigger.Condition
	state.TriggerJobID = types.Int64Value(int64(cond.JobID))
	state.ProjectID = types.Int64Value(int64(cond.ProjectID))

	statusStrings := make([]string, 0, len(cond.Statuses))
	for _, s := range cond.Statuses {
		name, ok := utils.JobCompletionTriggerConditionsMappingCodeHuman[s]
		if !ok {
			resp.Diagnostics.AddError("Unexpected status value", fmt.Sprintf("Unknown status int %d returned from API", s))
			return
		}
		statusStrings = append(statusStrings, name.(string))
	}
	statusSet, setDiags := types.SetValueFrom(ctx, types.StringType, statusStrings)
	resp.Diagnostics.Append(setDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Statuses = statusSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *jobCompletionTriggerResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan JobCompletionTriggerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := plan.JobID.ValueInt64()
	job, err := r.client.GetJob(strconv.FormatInt(jobID, 10))
	if err != nil {
		resp.Diagnostics.AddError("Error fetching job", err.Error())
		return
	}

	statuses, diags := buildStatusInts(plan.Statuses)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	job.JobCompletionTrigger = &dbt_cloud.JobCompletionTrigger{
		Condition: dbt_cloud.JobCompletionTriggerCondition{
			JobID:     int(plan.TriggerJobID.ValueInt64()),
			ProjectID: int(plan.ProjectID.ValueInt64()),
			Statuses:  statuses,
		},
	}

	_, err = r.client.UpdateJob(strconv.FormatInt(jobID, 10), *job)
	if err != nil {
		resp.Diagnostics.AddError("Error updating job completion trigger", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *jobCompletionTriggerResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state JobCompletionTriggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := state.JobID.ValueInt64()
	job, err := r.client.GetJob(strconv.FormatInt(jobID, 10))
	if err != nil {
		if helper.HandleResourceNotFound(ctx, err, &resp.Diagnostics, &resp.State, "job_completion_trigger") {
			return
		}
		resp.Diagnostics.AddError("Error fetching job", err.Error())
		return
	}

	job.JobCompletionTrigger = nil

	_, err = r.client.UpdateJob(strconv.FormatInt(jobID, 10), *job)
	if err != nil {
		resp.Diagnostics.AddError("Error removing job completion trigger", err.Error())
		return
	}
}

func (r *jobCompletionTriggerResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	jobID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing job ID for import", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), jobID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("job_id"), jobID)...)
}

// buildStatusInts converts the plan's statuses set to a slice of API integer codes
// using the shared utils.JobCompletionTriggerConditionsMappingHumanCode mapping.
func buildStatusInts(statusSet types.Set) ([]int, diag.Diagnostics) {
	statusStrings := helper.StringSetToStringSlice(statusSet)
	statuses := make([]int, 0, len(statusStrings))
	var diags diag.Diagnostics
	for _, s := range statusStrings {
		v, ok := utils.JobCompletionTriggerConditionsMappingHumanCode[s]
		if !ok {
			diags.AddError(
				"Invalid status value",
				fmt.Sprintf("Invalid status %q: must be one of success, error, canceled", s),
			)
			return nil, diags
		}
		statuses = append(statuses, v)
	}
	return statuses, diags
}
