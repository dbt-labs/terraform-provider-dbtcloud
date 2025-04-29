package partial_environment_variable

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/environment_variable"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

var (
	_ resource.Resource                = &partialEnvironmentVariableResource{}
	_ resource.ResourceWithConfigure   = &partialEnvironmentVariableResource{}
	_ resource.ResourceWithImportState = &partialEnvironmentVariableResource{}
)

func PartialEnvironmentVariableResource() resource.Resource {
	return &partialEnvironmentVariableResource{}
}

type partialEnvironmentVariableResource struct {
	client *dbt_cloud.Client
}

func (r *partialEnvironmentVariableResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_partial_environment_variable"
}

func (r *partialEnvironmentVariableResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state environment_variable.EnvironmentVariableResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// Extract project_id and name from state
	projectID := int(state.ProjectID.ValueInt64())
	name := state.Name.ValueString()

	// Get environment variable from API
	envVar, err := r.client.GetEnvironmentVariable(projectID, name)
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The environment variable resource was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting the environment variable", err.Error())
		return
	}

	abstractedEnvVar := convertFullToAbstractedEnvironmentVariable(envVar)

	// Set the global values
	state.ID = types.StringValue(fmt.Sprintf("%d:%s", projectID, name))
	state.ProjectID = types.Int64Value(int64(projectID))
	state.Name = types.StringValue(name)

	// Get configured environment values
	var configuredEnvValues map[string]string
	diags := state.EnvironmentValues.ElementsAs(ctx, &configuredEnvValues, false)
	if diags.HasError() {
		resp.Diagnostics.AddError("Error extracting environment values", "")
		return
	}

	// Keep only the environment values that are in both the config and remote
	// We only manage the environment values that are specified in the config
	managedEnvValues := make(map[string]string)
	for envName := range configuredEnvValues {
		if remoteValue, exists := abstractedEnvVar.EnvironmentValues[envName]; exists {
			managedEnvValues[envName] = remoteValue
		}
	}

	// Update the environment values in the state with the remote values
	updatedEnvValuesMap := make(map[string]attr.Value)
	for key, value := range managedEnvValues {
		updatedEnvValuesMap[key] = types.StringValue(value)
	}
	state.EnvironmentValues, _ = types.MapValue(types.StringType, updatedEnvValuesMap)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *partialEnvironmentVariableResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan environment_variable.EnvironmentVariableResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract values from plan
	projectID := int(plan.ProjectID.ValueInt64())
	name := plan.Name.ValueString()

	// Extract environment values from the plan
	var configEnvValues map[string]string
	diags := plan.EnvironmentValues.ElementsAs(ctx, &configEnvValues, false)
	if diags.HasError() {
		resp.Diagnostics.AddError("Error extracting environment values", "")
		return
	}

	// Check if environment variable already exists and fetch it
	existingEnvVar, err := r.client.GetEnvironmentVariable(projectID, name)
	if err != nil && !strings.Contains(err.Error(), "resource-not-found") {
		resp.Diagnostics.AddError(
			"Error checking for existing environment variable",
			"Error: "+err.Error(),
		)
		return
	}

	if existingEnvVar != nil {
		/// THOUGHT PROCESS DOC
		/// As updating with only the new value would delete the other envs' values, we need to "full update"
		/// Therefore the update request needs to contain all existing envs and values of the Env Var

		// Figure out what env vars are missing from the remote
		abstractedEnvVar := convertFullToAbstractedEnvironmentVariable(existingEnvVar)
		remoteKeys := make([]string, 0, len(abstractedEnvVar.EnvironmentValues))
		for k := range abstractedEnvVar.EnvironmentValues {
			remoteKeys = append(remoteKeys, k)
		}
		configKeys := make([]string, 0, len(configEnvValues))
		for k := range configEnvValues {
			configKeys = append(configKeys, k)
		}
		missingRemoteKeys, missingConfigKeys := lo.Difference(configKeys, remoteKeys)

		// Merge the missing values into the remote values to form a complete map of all the values
		mergedEnvVars := make(map[string]string)
		for _, key := range missingRemoteKeys {
			mergedEnvVars[key] = configEnvValues[key]
		}

		for _, key := range missingConfigKeys {
			// In the case where the plan value at create is different than the value in the remote
			// We need to specify the ID instead of Env Name as the key, otherwise the update does not go through
			idStr := strconv.Itoa(existingEnvVar.EnvironmentNameValues[key].ID)
			mergedEnvVars[idStr] = abstractedEnvVar.EnvironmentValues[key]
		}

		// Update the remote with the new values
		envVar := dbt_cloud.AbstractedEnvironmentVariable{
			Name:              name,
			ProjectID:         projectID,
			EnvironmentValues: mergedEnvVars,
		}
		_, err := r.client.UpdateEnvironmentVariable(projectID, envVar)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update existing environment variable",
				"Error: "+err.Error(),
			)
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	} else {
		// It doesn't exist, so we create it with our values
		_, err := r.client.CreateEnvironmentVariable(projectID, name, configEnvValues)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to create environment variable",
				"Error: "+err.Error(),
			)
			return
		}
	}

	// Set the ID and update the state
	plan.ID = types.StringValue(fmt.Sprintf("%d:%s", projectID, name))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *partialEnvironmentVariableResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state environment_variable.EnvironmentVariableResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	name := state.Name.ValueString()

	// Get the current environment variable
	envVar, err := r.client.GetEnvironmentVariable(projectID, name)
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") {
			// Already gone, nothing to do
			return
		}
		resp.Diagnostics.AddError("Error getting the environment variable", err.Error())
		return
	}

	abstractedEnvVar := convertFullToAbstractedEnvironmentVariable(envVar)

	// Get the environment values we're managing
	var managedEnvValues map[string]string
	diags := state.EnvironmentValues.ElementsAs(ctx, &managedEnvValues, false)
	if diags.HasError() {
		resp.Diagnostics.AddError("Error extracting environment values", "")
		return
	}

	// Get environment values that are not managed by this resource
	var remainingEnvValues = make(map[string]string)
	for envName, envValue := range abstractedEnvVar.EnvironmentValues {
		if _, exists := managedEnvValues[envName]; !exists {
			remainingEnvValues[envName] = envValue
		}
	}

	if len(remainingEnvValues) > 0 {
		// Create new map with ID as key to perform update and set their value to empty string
		removableValues := make(map[string]string)
		for key := range managedEnvValues {
			idStr := strconv.Itoa(envVar.EnvironmentNameValues[key].ID)
			removableValues[idStr] = ""
		}

		// Update the environment variable to remove only our managed keys
		updatedEnvVar := dbt_cloud.AbstractedEnvironmentVariable{
			Name:              name,
			ProjectID:         projectID,
			EnvironmentValues: removableValues,
		}

		_, err = r.client.UpdateEnvironmentVariable(projectID, updatedEnvVar)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update the environment variable",
				"Error: "+err.Error(),
			)
			return
		}
	} else {
		// If no environment values remain, delete the entire environment variable
		_, err = r.client.DeleteEnvironmentVariable(name, projectID)
		if err != nil {
			resp.Diagnostics.AddError("Error deleting the environment variable", err.Error())
			return
		}
	}
}

func (r *partialEnvironmentVariableResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state environment_variable.EnvironmentVariableResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	name := state.Name.ValueString()

	// Get current environment variable from API
	envVar, err := r.client.GetEnvironmentVariable(projectID, name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting the environment variable",
			"Error: "+err.Error(),
		)
		return
	}

	// Extract environment values from plan and state
	var planEnvValues map[string]string
	diags := plan.EnvironmentValues.ElementsAs(ctx, &planEnvValues, false)
	if diags.HasError() {
		resp.Diagnostics.AddError("Error extracting environment values from plan", "")
		return
	}

	var stateEnvValues map[string]string
	diags = state.EnvironmentValues.ElementsAs(ctx, &stateEnvValues, false)
	if diags.HasError() {
		resp.Diagnostics.AddError("Error extracting environment values from state", "")
		return
	}

	// Check for new environments or value changes and identify which values change from the state
	requiredEnvValues := make(map[string]string)
	for envName, planValue := range planEnvValues {
		stateValue, exists := stateEnvValues[envName]
		if !exists || planValue != stateValue {
			requiredEnvValues[envName] = planValue
		}
	}

	if len(requiredEnvValues) != 0 {
		updateEnvValues := make(map[string]string)
		// Add or update values from plan
		for key, value := range requiredEnvValues {
			id := envVar.EnvironmentNameValues[key].ID
			if id != 0 {
				idStr := strconv.Itoa(id)
				updateEnvValues[idStr] = value
			} else {
				updateEnvValues[key] = value
			}
		}

		// Update the environment variable with all values (managed and unmanaged)
		updatedEnvVar := dbt_cloud.AbstractedEnvironmentVariable{
			Name:              name,
			ProjectID:         projectID,
			EnvironmentValues: updateEnvValues,
		}

		_, err = r.client.UpdateEnvironmentVariable(projectID, updatedEnvVar)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update the environment variable",
				"Error: "+err.Error(),
			)
			return
		}
	}

	// Set the updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *partialEnvironmentVariableResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}

// ImportState imports the resource into Terraform state
func (r *partialEnvironmentVariableResource) ImportState(
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

func convertFullToAbstractedEnvironmentVariable(
	fullEnvVar *dbt_cloud.FullEnvironmentVariable,
) *dbt_cloud.AbstractedEnvironmentVariable {
	envValuesMap := make(map[string]string)
	for key, keyValuePair := range fullEnvVar.EnvironmentNameValues {
		envValuesMap[key] = keyValuePair.Value
	}

	return &dbt_cloud.AbstractedEnvironmentVariable{
		Name:              fullEnvVar.Name,
		ProjectID:         fullEnvVar.ProjectID,
		EnvironmentValues: envValuesMap,
	}
}
