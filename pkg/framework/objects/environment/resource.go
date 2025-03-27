package environment

import (
	"context"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &environmentResource{}
	_ resource.ResourceWithConfigure   = &environmentResource{}
	_ resource.ResourceWithImportState = &environmentResource{}
)

func EnvironmentResource() resource.Resource {
	return &environmentResource{}
}

type environmentResource struct {
	client *dbt_cloud.Client
}

func (r *environmentResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

// getEnvironmentDetails retrieves and maps environment details from the API
func (r *environmentResource) getEnvironmentDetails(projectID, environmentID int64) (*EnvironmentDataSourceModel, error) {
	environment, err := r.client.GetEnvironment(int(projectID), int(environmentID))
	if err != nil {
		return nil, err
	}

	state := &EnvironmentDataSourceModel{
		CredentialsID: types.Int64PointerValue(
			helper.IntPointerToInt64Pointer(environment.Credential_Id),
		),
		Name:            types.StringValue(environment.Name),
		DbtVersion:      types.StringValue(environment.Dbt_Version),
		Type:            types.StringValue(environment.Type),
		UseCustomBranch: types.BoolValue(environment.Use_Custom_Branch),
		CustomBranch:    types.StringPointerValue(environment.Custom_Branch),
		DeploymentType:  types.StringPointerValue(environment.DeploymentType),
		ExtendedAttributesID: types.Int64PointerValue(
			helper.IntPointerToInt64Pointer(environment.ExtendedAttributesID),
		),
		EnableModelQueryHistory: types.BoolValue(environment.EnableModelQueryHistory),
	}

	return state, nil
}

func (r *environmentResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state EnvironmentDataSourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	updatedState, err := r.getEnvironmentDetails(state.ProjectID.ValueInt64(), state.EnvironmentID.ValueInt64())
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The environment was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting the environment", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, updatedState)...)
}

func (r *environmentResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan EnvironmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	environment, err := r.client.CreateEnvironment(
		plan.IsActive.ValueBool(),
		int(plan.ProjectID.ValueInt64()),
		plan.Name.ValueString(),
		plan.DbtVersion.ValueString(),
		plan.Type.ValueString(),
		plan.UseCustomBranch.ValueBool(),
		plan.CustomBranch.ValueString(),
		int(plan.CredentialsID.ValueInt64()),
		plan.DeploymentType.ValueString(),
		int(plan.ExtendedAttributesID.ValueInt64()),
		int(plan.ConnectionID.ValueInt64()),
		plan.EnableModelQueryHistory.ValueBool(),
	)

	if err != nil {
		resp.Diagnostics.AddError("Error creating the environment", err.Error())
		return
	}

	plan.EnvironmentID = types.Int64Value(int64(*environment.Environment_Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *environmentResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state EnvironmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var envToUpdate dbt_cloud.Environment

	if plan.Name.ValueString() != state.Name.ValueString() {
		envToUpdate.Name = plan.Name.ValueString()
	} else {
		envToUpdate.Name = state.Name.ValueString()
	}

	if plan.DbtVersion.ValueString() != state.DbtVersion.ValueString() {
		envToUpdate.Dbt_Version = plan.DbtVersion.ValueString()
	} else {
		envToUpdate.Dbt_Version = state.DbtVersion.ValueString()
	}

	if plan.Type.ValueString() != state.Type.ValueString() {
		envToUpdate.Type = plan.Type.ValueString()
	} else {
		envToUpdate.Type = state.Type.ValueString()
	}

	if plan.UseCustomBranch.ValueBool() != state.UseCustomBranch.ValueBool() {
		envToUpdate.Use_Custom_Branch = plan.UseCustomBranch.ValueBool()
	} else {
		envToUpdate.Use_Custom_Branch = state.UseCustomBranch.ValueBool()
	}

	if plan.CustomBranch.ValueString() != state.CustomBranch.ValueString() {
		customBranch := plan.CustomBranch.ValueString()
		envToUpdate.Custom_Branch = &customBranch
	} else {
		customBranch := state.CustomBranch.ValueString()
		envToUpdate.Custom_Branch = &customBranch
	}

	if plan.DeploymentType.ValueString() != state.DeploymentType.ValueString() {
		deploymentType := plan.DeploymentType.ValueString()
		envToUpdate.DeploymentType = &deploymentType
	} else {
		deploymentType := state.DeploymentType.ValueString()
		envToUpdate.DeploymentType = &deploymentType
	}

	if plan.ExtendedAttributesID.ValueInt64() != state.ExtendedAttributesID.ValueInt64() {
		extendedAttrID := int(plan.ExtendedAttributesID.ValueInt64())
		envToUpdate.ExtendedAttributesID = &extendedAttrID
	} else {
		extendedAttrID := int(state.ExtendedAttributesID.ValueInt64())
		envToUpdate.ExtendedAttributesID = &extendedAttrID
	}

	if plan.ConnectionID.ValueInt64() != state.ConnectionID.ValueInt64() {
		connID := int(plan.ConnectionID.ValueInt64())
		envToUpdate.ConnectionID = &connID
	} else {
		connID := int(state.ConnectionID.ValueInt64())
		envToUpdate.ConnectionID = &connID
	}

	if plan.EnableModelQueryHistory.ValueBool() != state.EnableModelQueryHistory.ValueBool() {
		envToUpdate.EnableModelQueryHistory = plan.EnableModelQueryHistory.ValueBool()
	} else {
		envToUpdate.EnableModelQueryHistory = state.EnableModelQueryHistory.ValueBool()
	}

	_, err := r.client.UpdateEnvironment(
		envToUpdate.Project_Id,
		envToUpdate.Environment_Id,
		envToUpdate,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating environment", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *environmentResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state EnvironmentDataSourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteEnvironment(
		int(state.ProjectID.ValueInt64()),
		int(state.EnvironmentID.ValueInt64()),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting environment", err.Error())
		return
	}
}

func (r *environmentResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *environmentResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	r.client = req.ProviderData.(*dbt_cloud.Client)
}
