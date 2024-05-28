package partial_license_map

import (
	"context"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

var (
	_ resource.Resource              = &partialLicenseMapResource{}
	_ resource.ResourceWithConfigure = &partialLicenseMapResource{}
)

func PartialLicenseMapResource() resource.Resource {
	return &partialLicenseMapResource{}
}

type partialLicenseMapResource struct {
	client *dbt_cloud.Client
}

func (r *partialLicenseMapResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_partial_license_map"
}

func (r *partialLicenseMapResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state LicenseMapResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// check if the ID exists
	licenseMapID := int(state.ID.ValueInt64())
	licenseMap, err := r.client.GetLicenseMap(licenseMapID)
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

	// if the ID exists, make sure that it is the one we are looking for
	if !matchPartial(state, *licenseMap) {
		// read all the objects and check if one exists
		allLicenseMaps, err := r.client.GetAllLicenseMaps()
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get all license maps",
				"Error: "+err.Error(),
			)
			return
		}

		var fullLicenseMap *dbt_cloud.LicenseMap
		for _, licenseMap := range allLicenseMaps {
			if matchPartial(state, licenseMap) {
				// it exists, we stop here
				fullLicenseMap = &licenseMap
				break
			}
		}

		// if it was not found, it means that the object was deleted
		if fullLicenseMap == nil {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The license map resource was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}

		// if it is found, we set it correctly
		licenseMapID = *fullLicenseMap.ID
		licenseMap = fullLicenseMap
	}

	// we set the "global" values
	state.ID = types.Int64Value(int64(licenseMapID))
	state.LicenseType = types.StringValue(licenseMap.LicenseType)

	// we set the "partial" values by intersecting the config with the remote
	var ssoMappingConfigured []string
	diags := state.SSOLicenseMappingGroups.ElementsAs(
		context.Background(),
		&ssoMappingConfigured,
		false,
	)
	if diags.HasError() {
		resp.Diagnostics.AddError("Error extracting the list of SSO groups", "")
		return
	}

	state.SSOLicenseMappingGroups, _ = types.SetValueFrom(
		context.Background(),
		types.StringType,
		lo.Intersect(ssoMappingConfigured, licenseMap.SSOLicenseMappingGroups),
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *partialLicenseMapResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan LicenseMapResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// we read the values from the config
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

	// check if it exists
	// we don't need to check uniqueness and can just return the first as the API only allows one license type
	allLicenseMaps, err := r.client.GetAllLicenseMaps()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get all license maps",
			"Error: "+err.Error(),
		)
		return
	}

	var fullLicenseMap *dbt_cloud.LicenseMap
	for _, licenseMap := range allLicenseMaps {
		if matchPartial(plan, licenseMap) {
			// it exists, we stop here
			fullLicenseMap = &licenseMap
			break
		}
	}

	if fullLicenseMap != nil {
		// if it exists, we get the ID
		licenseMapID := fullLicenseMap.ID
		plan.ID = types.Int64PointerValue(helper.IntPointerToInt64Pointer(licenseMapID))

		// and we calculate all the partial fields
		// the global ones are already set in the plan
		remoteSsoMapping := fullLicenseMap.SSOLicenseMappingGroups
		missingSsoMapping := lo.Without(configSsoMapping, remoteSsoMapping...)

		// we only update if something global, but not part of the ID is different or if something partial needs to be added
		if len(missingSsoMapping) == 0 {
			// nothing to do if they are all the same
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		} else {
			// if one of them is different, we get the new values for all
			// and we update the object
			allSsoMapping := append(remoteSsoMapping, missingSsoMapping...)
			fullLicenseMap.SSOLicenseMappingGroups = allSsoMapping

			_, err := r.client.UpdateLicenseMap(*licenseMapID, *fullLicenseMap)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to update the existing license map",
					"Error: "+err.Error(),
				)
				return
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		}

	} else {
		// it doesn't exist so we create it
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
}

func (r *partialLicenseMapResource) Delete(
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
	licenseMap, err := r.client.GetLicenseMap(licenseMapID)
	if err != nil {
		resp.Diagnostics.AddError("Error getting the license map", err.Error())
		return
	}

	// we read the values from the config
	var configSsoMapping []string
	diags := state.SSOLicenseMappingGroups.ElementsAs(
		context.Background(),
		&configSsoMapping,
		false,
	)
	if diags.HasError() {
		resp.Diagnostics.AddError("Error extracting the list of SSO groups", "")
		return
	}

	remoteSsoMapping := licenseMap.SSOLicenseMappingGroups
	requiredSsoMapping := lo.Without(remoteSsoMapping, configSsoMapping...)

	if len(requiredSsoMapping) > 0 {
		// we update the object if there are some partial values left
		// but we leave the object existing, without deleting it entirely
		licenseMap.SSOLicenseMappingGroups = requiredSsoMapping
		_, err = r.client.UpdateLicenseMap(licenseMapID, *licenseMap)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update the existing license map",
				"Error: "+err.Error(),
			)
			return
		}
	} else {
		// we delete the object if there is no config left at all
		err = r.client.DestroyLicenseMap(licenseMapID)
		if err != nil {
			resp.Diagnostics.AddError("Error deleting the license map", err.Error())
			return
		}
	}
}

func (r *partialLicenseMapResource) Update(
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

	// we compare the partial objects and update them if needed
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

	remoteSsoMapping := licenseMap.SSOLicenseMappingGroups
	deletedSsoMapping := lo.Without(stateSsoMapping, planSsoMapping...)
	newSsoMapping := lo.Without(planSsoMapping, stateSsoMapping...)
	requiredSsoMapping := lo.Without(
		lo.Union(remoteSsoMapping, newSsoMapping),
		deletedSsoMapping...)

	// we check if there are changes to be sent, both global and local
	if len(deletedSsoMapping) > 0 ||
		len(newSsoMapping) > 0 {

		// we update the values to be the plan ones for global
		// and the calculated ones for the local ones
		licenseMap.SSOLicenseMappingGroups = requiredSsoMapping
		_, err = r.client.UpdateLicenseMap(licenseMapID, *licenseMap)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update the existing license map",
				"Error: "+err.Error(),
			)
			return
		}
	}

	// Set the updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *partialLicenseMapResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
