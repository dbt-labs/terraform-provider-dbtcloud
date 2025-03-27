package environment

import (
	"context"

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

func (r *environmentResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state EnvironmentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	environment, err := r.client.GetEnvironment(int(state.ProjectID.ValueInt64()), int(state.EnvironmentID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error getting the environment", err.Error())
		return
	}

	state.EnvironmentID = types.Int64Value(int64(*environment.Environment_Id))
	state.ID = state.EnvironmentID
	state.Name = types.StringValue(environment.Name)
	state.ProjectID = types.Int64Value(int64(environment.Project_Id))
	state.IsActive = types.BoolValue(environment.State == dbt_cloud.STATE_ACTIVE)
	state.DbtVersion = types.StringValue(environment.Dbt_Version)
	state.Type = types.StringValue(environment.Type)
	state.UseCustomBranch = types.BoolValue(environment.Use_Custom_Branch)
	state.CustomBranch = types.StringPointerValue(environment.Custom_Branch)
	state.DeploymentType = types.StringPointerValue(environment.DeploymentType)
	if environment.ExtendedAttributesID != nil {
		state.ExtendedAttributesID = types.Int64Value(int64(*environment.ExtendedAttributesID))
	} else {
		state.ExtendedAttributesID = types.Int64Value(0)
	}
	state.EnableModelQueryHistory = types.BoolValue(environment.EnableModelQueryHistory)
	state.ConnectionID = types.Int64PointerValue(
		helper.IntPointerToInt64Pointer(environment.ConnectionID),
	)
	state.CredentialID = types.Int64PointerValue(
		helper.IntPointerToInt64Pointer(environment.Credential_Id),
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
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
		int(plan.CredentialID.ValueInt64()),
		plan.DeploymentType.ValueString(),
		int(plan.ExtendedAttributesID.ValueInt64()),
		int(plan.ConnectionID.ValueInt64()),
		plan.EnableModelQueryHistory.ValueBool(),
	)

	if err != nil {
		resp.Diagnostics.AddError("Error creating the environment", err.Error())
		return
	}

	plan.ID = types.Int64Value(int64(*environment.ID))
	plan.Name = types.StringValue(environment.Name)
	plan.EnvironmentID = types.Int64Value(int64(*environment.Environment_Id))
	plan.ProjectID = types.Int64Value(int64(environment.Project_Id))
	plan.IsActive = types.BoolValue(environment.State == dbt_cloud.STATE_ACTIVE)
	plan.DbtVersion = types.StringValue(environment.Dbt_Version)
	plan.Type = types.StringValue(environment.Type)
	plan.UseCustomBranch = types.BoolValue(environment.Use_Custom_Branch)
	plan.CustomBranch = types.StringPointerValue(environment.Custom_Branch)
	plan.DeploymentType = types.StringPointerValue(environment.DeploymentType)
	if environment.ExtendedAttributesID != nil {
		plan.ExtendedAttributesID = types.Int64Value(int64(*environment.ExtendedAttributesID))
	} else {
		plan.ExtendedAttributesID = types.Int64Value(0)
	}
	plan.EnableModelQueryHistory = types.BoolValue(environment.EnableModelQueryHistory)
	plan.ConnectionID = types.Int64PointerValue(
		helper.IntPointerToInt64Pointer(environment.ConnectionID),
	)
	plan.CredentialID = types.Int64PointerValue(
		helper.IntPointerToInt64Pointer(environment.Credential_Id),
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *environmentResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state EnvironmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	envToUpdate, err := r.client.GetEnvironment(int(plan.ProjectID.ValueInt64()), int(plan.EnvironmentID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error getting the environment", err.Error())
		return
	}

	if plan.Name.ValueString() != state.Name.ValueString() {
		envToUpdate.Name = plan.Name.ValueString()
	}

	if plan.CredentialID.ValueInt64() != state.CredentialID.ValueInt64() {
		envToUpdate.Credential_Id = helper.Int64ToIntPointer(plan.CredentialID.ValueInt64())
	}

	if plan.DbtVersion.ValueString() != state.DbtVersion.ValueString() {
		envToUpdate.Dbt_Version = plan.DbtVersion.ValueString()
	}

	if plan.Type.ValueString() != state.Type.ValueString() {
		envToUpdate.Type = plan.Type.ValueString()
	}

	if plan.UseCustomBranch.ValueBool() != state.UseCustomBranch.ValueBool() {
		envToUpdate.Use_Custom_Branch = plan.UseCustomBranch.ValueBool()
	}

	if plan.CustomBranch.ValueString() != state.CustomBranch.ValueString() {
		customBranch := plan.CustomBranch.ValueString()
		envToUpdate.Custom_Branch = &customBranch
	}

	if plan.DeploymentType.ValueString() != state.DeploymentType.ValueString() {
		deploymentType := plan.DeploymentType.ValueString()
		envToUpdate.DeploymentType = &deploymentType
	}

	if plan.ExtendedAttributesID.ValueInt64() != state.ExtendedAttributesID.ValueInt64() {
		extendedAttrID := int(plan.ExtendedAttributesID.ValueInt64())
		envToUpdate.ExtendedAttributesID = &extendedAttrID
	}

	if plan.ConnectionID.ValueInt64() != state.ConnectionID.ValueInt64() {
		connID := int(plan.ConnectionID.ValueInt64())
		envToUpdate.ConnectionID = &connID
	}

	if plan.EnableModelQueryHistory.ValueBool() != state.EnableModelQueryHistory.ValueBool() {
		envToUpdate.EnableModelQueryHistory = plan.EnableModelQueryHistory.ValueBool()
	}

	_, err = r.client.UpdateEnvironment(
		int(plan.ProjectID.ValueInt64()),
		int(plan.EnvironmentID.ValueInt64()),
		*envToUpdate,
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
	var state EnvironmentResourceModel
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
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
