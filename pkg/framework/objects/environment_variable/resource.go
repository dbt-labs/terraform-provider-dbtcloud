package environment_variable

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &environmentVariableResource{}
	_ resource.ResourceWithConfigure   = &environmentVariableResource{}
	_ resource.ResourceWithImportState = &environmentVariableResource{}
)

// EnvironmentVariableResource is a helper function to simplify the provider implementation.
func EnvironmentVariableResource() resource.Resource {
	return &environmentVariableResource{}
}

// environmentVariableResource is the resource implementation.
type environmentVariableResource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the resource.
func (r *environmentVariableResource) Configure(
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
func (r *environmentVariableResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_environment_variable"
}

// Schema defines the schema for the resource.
func (r *environmentVariableResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resourceSchema
}

// Create creates the resource and sets the initial Terraform state.
func (r *environmentVariableResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Retrieve values from plan
	var plan EnvironmentVariableResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	name := plan.Name.ValueString()
	environmentValues := plan.EnvironmentValues.Elements()

	envValuesMap := make(map[string]string)
	for key, value := range environmentValues {
		if valueStr, ok := value.(types.String); ok {
			envValuesMap[key] = valueStr.ValueString()
		}
	}

	// Create new envVar
	envVar, err := r.client.CreateEnvironmentVariable(
		int(projectID),
		name,
		envValuesMap,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating envrionment variable",
			"Could not create environment variable, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate computed values
	plan.ID = types.StringValue(fmt.Sprintf("%d:%s", projectID, envVar.Name))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *environmentVariableResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state
	var state EnvironmentVariableResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get env var from API
	projectID := int(state.ProjectID.ValueInt64())
	name := state.Name.ValueString()

	envVar, err := r.client.GetEnvironmentVariable(projectID, name)
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
	state.ID = types.StringValue(fmt.Sprintf("%d:%s", projectID, envVar.Name))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *environmentVariableResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Retrieve values from plan
	var plan EnvironmentVariableResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state EnvironmentVariableResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	name := plan.Name.ValueString()
	// Get current environment variable from API
	currentEnvVar, err := r.client.GetEnvironmentVariable(projectID, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting the environment variable",
			"Error: "+err.Error(),
		)
		return
	}

	// Update values for previously existing environments and disregard those missing from plan
	environmentValues := plan.EnvironmentValues.Elements()
	envValuesMap := make(map[string]string)
	for key, keyValuePair := range currentEnvVar.EnvironmentNameValues {
		idStr := strconv.Itoa(keyValuePair.ID)
		// We assume the value will be deleted
		envValuesMap[idStr] = ""
		if valueStr, ok := environmentValues[key].(types.String); ok {
			envValuesMap[idStr] = valueStr.ValueString()
		}
	}

	// Add any new values
	if len(environmentValues) > len(currentEnvVar.EnvironmentNameValues) {
		for key, value := range environmentValues {
			_, exists := currentEnvVar.EnvironmentNameValues[key]
			if !exists {
				if valueStr, ok := value.(types.String); ok {
					envValuesMap[key] = valueStr.ValueString()
				}
			}
		}
	}

	envVar := dbt_cloud.AbstractedEnvironmentVariable{
		Name:              name,
		ProjectID:         projectID,
		EnvironmentValues: envValuesMap,
	}

	// Update credential
	_, err = r.client.UpdateEnvironmentVariable(
		projectID,
		envVar,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating envrionment variable",
			"Could not update envrionment variable, unexpected error: "+err.Error(),
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
func (r *environmentVariableResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Retrieve values from state
	var state EnvironmentVariableResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete env var
	projectID := state.ProjectID.ValueInt64()
	name := state.Name.ValueString()

	_, err := r.client.DeleteEnvironmentVariable(
		name,
		int(projectID),
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
func (r *environmentVariableResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Extract the resource ID
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 {
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

	name := idParts[1]
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("id"),
		fmt.Sprintf("%d:%s", projectID, name),
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("project_id"),
		projectID,
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("name"),
		name,
	)...)
}
