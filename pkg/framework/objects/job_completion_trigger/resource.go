package job_completion_trigger

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
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

func (r *jobCompletionTriggerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job_completion_trigger"
}

func (r *jobCompletionTriggerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema()
}

func (r *jobCompletionTriggerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dbt_cloud.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *dbt_cloud.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *jobCompletionTriggerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan JobCompletionTriggerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := plan.JobID.ValueInt64()
	jobIDStr := strconv.FormatInt(jobID, 10)

	// Fetch current job
	job, err := r.client.GetJob(jobIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching job", err.Error())
		return
	}

	// Prepare trigger payload
	triggerJobID := int(plan.TriggerJobID.ValueInt64())
	projectID := int(plan.ProjectID.ValueInt64())
	var statuses []string
	plan.Statuses.ElementsAs(ctx, &statuses, false)

	statusInts := make([]int, len(statuses))
	for i, s := range statuses {
		if code, ok := utils.JobCompletionTriggerConditionsMappingHumanCode[s]; ok {
			statusInts[i] = code
		} else {
			resp.Diagnostics.AddError("Invalid status", s)
			return
		}
	}

	trigger := dbt_cloud.JobCompletionTrigger{
		Condition: dbt_cloud.JobCompletionTriggerCondition{
			JobID:     triggerJobID,
			ProjectID: projectID,
			Statuses:  statusInts,
		},
	}
	job.JobCompletionTrigger = &trigger

	// Update Job
	_, err = r.client.UpdateJob(jobIDStr, *job)
	if err != nil {
		resp.Diagnostics.AddError("Error updating job with trigger", err.Error())
		return
	}

	plan.ID = types.StringValue(jobIDStr)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *jobCompletionTriggerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state JobCompletionTriggerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobIDStr := state.ID.ValueString()
	job, err := r.client.GetJob(jobIDStr)
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading job", err.Error())
		return
	}

	if job.JobCompletionTrigger == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state
	state.TriggerJobID = types.Int64Value(int64(job.JobCompletionTrigger.Condition.JobID))
	state.ProjectID = types.Int64Value(int64(job.JobCompletionTrigger.Condition.ProjectID))
	state.JobID = types.Int64Value(int64(*job.ID))

	statusStrings := make([]string, len(job.JobCompletionTrigger.Condition.Statuses))
	for i, code := range job.JobCompletionTrigger.Condition.Statuses {
		if s, ok := utils.JobCompletionTriggerConditionsMappingCodeHuman[code]; ok {
			statusStrings[i] = s.(string)
		} else {
			statusStrings[i] = "unknown"
		}
	}
	// Use lo to dedupe just in case, though API returns set-like behavior?
	state.Statuses, _ = types.SetValueFrom(ctx, types.StringType, lo.Uniq(statusStrings))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *jobCompletionTriggerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan JobCompletionTriggerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := plan.JobID.ValueInt64()
	jobIDStr := strconv.FormatInt(jobID, 10)

	// Fetch current job
	job, err := r.client.GetJob(jobIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching job", err.Error())
		return
	}

	// Prepare trigger payload
	triggerJobID := int(plan.TriggerJobID.ValueInt64())
	projectID := int(plan.ProjectID.ValueInt64())
	var statuses []string
	plan.Statuses.ElementsAs(ctx, &statuses, false)

	statusInts := make([]int, len(statuses))
	for i, s := range statuses {
		if code, ok := utils.JobCompletionTriggerConditionsMappingHumanCode[s]; ok {
			statusInts[i] = code
		} else {
			resp.Diagnostics.AddError("Invalid status", s)
			return
		}
	}

	trigger := dbt_cloud.JobCompletionTrigger{
		Condition: dbt_cloud.JobCompletionTriggerCondition{
			JobID:     triggerJobID,
			ProjectID: projectID,
			Statuses:  statusInts,
		},
	}
	job.JobCompletionTrigger = &trigger

	// Update Job
	_, err = r.client.UpdateJob(jobIDStr, *job)
	if err != nil {
		resp.Diagnostics.AddError("Error updating job with trigger", err.Error())
		return
	}

	plan.ID = types.StringValue(jobIDStr)
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *jobCompletionTriggerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state JobCompletionTriggerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobIDStr := state.ID.ValueString()
	job, err := r.client.GetJob(jobIDStr)
	if err != nil {
		// If job not found, trigger is gone
		return
	}

	job.JobCompletionTrigger = nil
	_, err = r.client.UpdateJob(jobIDStr, *job)
	if err != nil {
		resp.Diagnostics.AddError("Error removing job trigger", err.Error())
	}
}

func (r *jobCompletionTriggerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

