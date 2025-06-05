package semantic_layer_configuration

import (
	"context"
	"strings"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &semanticLayerConfigurationResource{}
	_ resource.ResourceWithConfigure   = &semanticLayerConfigurationResource{}
)

func SemanticLayerConfigurationResource() resource.Resource {
	return &semanticLayerConfigurationResource{}
}

type semanticLayerConfigurationResource struct {
	client *dbt_cloud.Client
}

func (r *semanticLayerConfigurationResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_semantic_layer_configuration"
}

func (r *semanticLayerConfigurationResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state SemanticLayerConfigurationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	configID := state.ID.ValueInt64()
	projectID := state.ProjectID.ValueInt64()
	retrievedConfig, err := r.client.GetSemanticLayerConfiguration(projectID, configID)

	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The Semantic Layer configuration was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting the Semantic Layer configuration", err.Error())
		return
	}

	state.ID = types.Int64Value(retrievedConfig.ID)
	state.ProjectID = types.Int64Value(retrievedConfig.ProjectID)
	state.EnvironmentID = types.Int64Value(retrievedConfig.EnvironmentID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *semanticLayerConfigurationResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SemanticLayerConfigurationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueInt64()
	environmentID := plan.EnvironmentID.ValueInt64()

	// Check if there is at least one successful run in the environment
	filter := dbt_cloud.RunFilter{
		EnvironmentID: int(environmentID),
		Status:        10,
	}

	runs, err := r.client.GetRuns(&filter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Issue when retrieving runs",
			err.Error(),
		)
		return
	}

	if len(*runs) == 0 {
		resp.Diagnostics.AddError(
			"No successful runs found",
			"Please run a job in the environment before creating a Semantic Layer configuration.",
		)
		return
	}

	createdConfig, err := r.client.CreateSemanticLayerConfiguration(
		projectID,
		environmentID,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Semantic Layer configuration",
			"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.Int64Value(createdConfig.ID)

	// Fetch associated project
	project, err := r.client.GetProject(strconv.FormatInt(projectID, 10))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to retrieve project",
			"Error: "+err.Error(),
		)
		return
	}

	// Update Project with Semantic Layer configuration ID
	project.SemanticLayerConfigID = &createdConfig.ID
	_, err = r.client.UpdateProject(strconv.FormatInt(projectID, 10), *project)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update project with Semantic Layer configuration ID",
			"Error: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *semanticLayerConfigurationResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SemanticLayerConfigurationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configID := state.ID
	projectID := state.ProjectID.ValueInt64()

	err := r.client.DeleteSemanticLayerConfiguration(projectID, configID.ValueInt64())

	if err != nil {
		resp.Diagnostics.AddError(
			"Issue deleting Semantic Layer Configuration",
			"Error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *semanticLayerConfigurationResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state SemanticLayerConfigurationModel

	// Read plan and state values into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	configID := state.ID.ValueInt64()
	projectID := state.ProjectID.ValueInt64()

	retrievedConfigID, err := r.client.GetSemanticLayerConfiguration(projectID, configID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Issue getting Semantic Layer configuration",
			"Error: "+err.Error(),
		)
		return
	}

	retrievedConfigID.EnvironmentID = plan.EnvironmentID.ValueInt64()

	_, err = r.client.UpdateSemanticLayerConfiguration(
		projectID,
		configID,
		*retrievedConfigID,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update Semantic Layer configuration",
			"Error: "+err.Error(),
		)
		return
	}

	state.ID = types.Int64Value(retrievedConfigID.ID)
	state.EnvironmentID = types.Int64Value(retrievedConfigID.EnvironmentID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *semanticLayerConfigurationResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
