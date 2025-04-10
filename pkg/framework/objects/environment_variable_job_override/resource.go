package environment_variable_job_override

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &environmentVariableJobOverrideResource{}
	_ resource.ResourceWithConfigure   = &environmentVariableJobOverrideResource{}
	_ resource.ResourceWithImportState = &environmentVariableJobOverrideResource{}
)

// EnvironmentVariableJobOverrideResource is a helper function to simplify the provider implementation.
func EnvironmentVariableJobOverrideResource() resource.Resource {
	return &environmentVariableJobOverrideResource{}
}

// environmentVariableJobOverrideResource is the resource implementation.
type environmentVariableJobOverrideResource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the resource.
func (r *environmentVariableJobOverrideResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dbt_cloud.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *dbt_cloud.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *environmentVariableJobOverrideResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_environment_variable_job_override"
}

// Schema defines the schema for the resource.
func (r *environmentVariableJobOverrideResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resourceSchema
}

// Create creates the resource and sets the initial Terraform state.
func (r *environmentVariableJobOverrideResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Retrieve values from plan
	var plan EnvironmentVariableJobOverrideResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	name := plan.Name.ValueString()
	rawValue := plan.RawValue.ValueString()
	jobDefinitionID := int(plan.JobDefinitionID.ValueInt64())

	// Create new envVar
	environmentVariableJobOverride, err := r.client.CreateEnvironmentVariableJobOverride(
		int(projectID),
		name,
		rawValue,
		jobDefinitionID,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating envrionment variable",
			"Could not create environment variable, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate computed values
	plan.ID = types.StringValue(fmt.Sprintf(
		"%d%s%d%s%d",
		environmentVariableJobOverride.ProjectID,
		dbt_cloud.ID_DELIMITER,
		environmentVariableJobOverride.JobDefinitionID,
		dbt_cloud.ID_DELIMITER,
		*environmentVariableJobOverride.ID,
	))

	plan.EnvironmentVariableJobOverrideID = types.Int64Value(int64(*environmentVariableJobOverride.ID))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *environmentVariableJobOverrideResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state
	var state EnvironmentVariableJobOverrideResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get env var from API
	projectID := int(state.ProjectID.ValueInt64())
	jobDefinitionID := int(state.JobDefinitionID.ValueInt64())
	id := state.EnvironmentVariableJobOverrideID.ValueInt64()

	environmentVariableJobOverride, err := r.client.GetEnvironmentVariableJobOverride(projectID, jobDefinitionID, int(id))
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading envrionment variable",
			"Could not read environment variable ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Refresh state values
	//state.ID = types.StringValue(fmt.Sprintf("%d:%s", projectID, envVar.Name))

	state.ID = types.StringValue(fmt.Sprintf(
		"%d%s%d%s%d",
		environmentVariableJobOverride.ProjectID,
		dbt_cloud.ID_DELIMITER,
		environmentVariableJobOverride.JobDefinitionID,
		dbt_cloud.ID_DELIMITER,
		*environmentVariableJobOverride.ID,
	))
	state.EnvironmentVariableJobOverrideID = types.Int64Value(int64(*environmentVariableJobOverride.ID))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *environmentVariableJobOverrideResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Retrieve values from plan
	var plan EnvironmentVariableJobOverrideResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state EnvironmentVariableJobOverrideResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	id := plan.EnvironmentVariableJobOverrideID.ValueInt64()

	envVarJobOverride := dbt_cloud.EnvironmentVariableJobOverride{
		ProjectID:       projectID,
		Name:            plan.Name.ValueString(),
		ID:              helper.Int64ToIntPointer(id),
		JobDefinitionID: int(plan.JobDefinitionID.ValueInt64()),
		RawValue:        plan.RawValue.ValueString(),
		Type:            "job",
	}

	// Update credential
	_, err := r.client.UpdateEnvironmentVariableJobOverride(projectID, int(id), envVarJobOverride)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating envrionment variable job override",
			"Could not update envrionment variable jobe override, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *environmentVariableJobOverrideResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Retrieve values from state
	var state EnvironmentVariableJobOverrideResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete env var
	projectID := state.ProjectID.ValueInt64()
	id := helper.Int64ToIntPointer(state.EnvironmentVariableJobOverrideID.ValueInt64())

	_, err := r.client.DeleteEnvironmentVariableJobOverride(
		int(projectID), *id,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting envrionment variable",
			"Could not delete envrionment variable, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *environmentVariableJobOverrideResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Extract the resource ID
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 3 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf(
				"Expected import identifier with format: project_id:name. Got: %q",
				req.ID,
			),
		)
		return
	}

	projectID, err := strconv.Atoi(idParts[0])
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf(
				"Could not convert project_id to integer. Got: %q",
				idParts[0],
			),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("project_id"),
		projectID,
	)...)

	jobDefinitionID, err := strconv.Atoi(idParts[1])
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf(
				"Could not convert job_definition_id to integer. Got: %q",
				idParts[1],
			),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("job_definition_id"),
		jobDefinitionID,
	)...)

	envVarJobOverrideID := idParts[2]
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("environment_variable_job_override_id"),
		envVarJobOverrideID,
	)...)
}
