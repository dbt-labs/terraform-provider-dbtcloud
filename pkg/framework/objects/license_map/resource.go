package license_map

import (
	"context"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
	"strconv"
	"strings"
)

var (
	_ resource.Resource              = &licenseMapResource{}
	_ resource.ResourceWithConfigure = &licenseMapResource{}

	licenseTypes = []string{
		"developer",
		"read_only",
		"it",
	}
)

func LicenseMapResource() resource.Resource {
	return &licenseMapResource{}
}

type licenseMapResource struct {
	client *dbt_cloud.Client
}

func (r *licenseMapResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_license_map"
}

func (r *licenseMapResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state LicenseMapResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	licenseMapID := state.ID.ValueInt64()
	licenseMap, err := r.client.GetLicenseMap(int(licenseMapID))
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The license map resource was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting the license map", err.Error())
		return
	}

	state.ID = types.Int64Value(int64(*licenseMap.ID))
	state.LicenseType = types.StringValue(licenseMap.LicenseType)
	state.SSOLicenseMappingGroups, _ = types.SetValueFrom(
		context.Background(),
		types.StringType,
		licenseMap.SSOLicenseMappingGroups,
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *licenseMapResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan LicenseMapResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var configSsoMapping []string
	diags := plan.SSOLicenseMappingGroups.ElementsAs(
		context.Background(),
		&configSsoMapping,
		false,
	)
	if diags.HasError() {
		resp.Diagnostics.AddError("Error extracting the list of SSO groups", "")
		return
	}

	licenseMap, err := r.client.CreateLicenseMap(
		plan.LicenseType.ValueString(),
		configSsoMapping,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create license map",
			"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.Int64PointerValue(helper.IntPointerToInt64Pointer(licenseMap.ID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *licenseMapResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state LicenseMapResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	licenseMapID := int(state.ID.ValueInt64())

	err := r.client.DestroyLicenseMap(licenseMapID)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting the license map", err.Error())
		return
	}
}

func (r *licenseMapResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state LicenseMapResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	licenseMapID := int(state.ID.ValueInt64())
	licenseMap, err := r.client.GetLicenseMap(licenseMapID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting the license map",
			"Error: "+err.Error(),
		)
		return
	}

	var planSsoMapping []string
	diags := plan.SSOLicenseMappingGroups.ElementsAs(
		context.Background(),
		&planSsoMapping,
		false,
	)
	if diags.HasError() {
		resp.Diagnostics.AddError("Error extracting the list of SSO groups from the plan", "")
		return
	}

	var stateSsoMapping []string
	diags = state.SSOLicenseMappingGroups.ElementsAs(
		context.Background(),
		&stateSsoMapping,
		false,
	)
	if diags.HasError() {
		resp.Diagnostics.AddError("Error extracting the list of SSO groups from the state", "")
		return
	}

	deletedSsoMapping, newSsoMapping := lo.Difference(stateSsoMapping, planSsoMapping)
	hasMappingChanges := len(deletedSsoMapping) > 0 || len(newSsoMapping) > 0

	if state.LicenseType != plan.LicenseType || hasMappingChanges {
		if state.LicenseType != plan.LicenseType {
			licenseMap.LicenseType = plan.LicenseType.ValueString()
		}

		if hasMappingChanges {
			licenseMap.SSOLicenseMappingGroups = planSsoMapping
		}

		_, err = r.client.UpdateLicenseMap(licenseMapID, *licenseMap)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update the existing license map",
				"Error: "+err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *licenseMapResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	licenseMapID, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing license map ID for import", err.Error())
		return
	}
	ssoLicenseMappingGroups, _ := types.SetValue(types.StringType, nil)
	state := LicenseMapResourceModel{
		ID:                      types.Int64Value(int64(licenseMapID)),
		SSOLicenseMappingGroups: ssoLicenseMappingGroups,
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *licenseMapResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
