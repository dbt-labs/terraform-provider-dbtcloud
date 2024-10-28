package lineage_integration

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &lineageIntegrationResource{}
	_ resource.ResourceWithConfigure   = &lineageIntegrationResource{}
	_ resource.ResourceWithImportState = &lineageIntegrationResource{}
)

func LineageIntegrationResource() resource.Resource {
	return &lineageIntegrationResource{}
}

type lineageIntegrationResource struct {
	client *dbt_cloud.Client
}

func (r *lineageIntegrationResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_lineage_integration"
}

func (r *lineageIntegrationResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data LineageIntegrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	projectID := data.ProjectID.ValueInt64()
	lineageIntegrationID := data.LineageIntegrationID.ValueInt64()
	lineageIntegration, err := r.client.GetLineageIntegration(projectID, lineageIntegrationID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The lineage_integration resource was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting the lineage", err.Error())
		return
	}

	data.ID = types.StringValue(
		fmt.Sprintf(
			"%d%s%d",
			lineageIntegration.ProjectID,
			dbt_cloud.ID_DELIMITER,
			*lineageIntegration.ID,
		),
	)

	data.LineageIntegrationID = types.Int64PointerValue(lineageIntegration.ID)
	data.ProjectID = types.Int64Value(int64(lineageIntegration.ProjectID))
	data.Name = types.StringValue(lineageIntegration.Name)
	data.Host = types.StringValue(lineageIntegration.Config.Host)
	data.SiteID = types.StringValue(lineageIntegration.Config.SiteID)
	data.TokenName = types.StringValue(lineageIntegration.Config.TokenName)

	// we only set the token if it is null as it is sensitive
	// this means that we are importing the data
	if data.Token.IsNull() {
		data.Token = types.StringValue("********")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *lineageIntegrationResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var data LineageIntegrationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lineageIntegration, err := r.client.CreateLineageIntegration(
		data.ProjectID.ValueInt64(),
		data.Name.ValueString(),
		data.Host.ValueString(),
		data.SiteID.ValueString(),
		data.TokenName.ValueString(),
		data.Token.ValueString(),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create lineage integration",
			"Error: "+err.Error(),
		)
		return
	}

	data.ID = types.StringValue(
		fmt.Sprintf(
			"%d%s%d",
			lineageIntegration.ProjectID,
			dbt_cloud.ID_DELIMITER,
			*lineageIntegration.ID,
		),
	)
	data.LineageIntegrationID = types.Int64PointerValue(lineageIntegration.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *lineageIntegrationResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data LineageIntegrationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lineageID := data.LineageIntegrationID.ValueInt64()
	projectID := data.ProjectID.ValueInt64()

	err := r.client.DeleteLineageIntegration(projectID, lineageID)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting lineage", err.Error())
		return
	}
}

func (r *lineageIntegrationResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state LineageIntegrationResourceModel

	// Read plan and state values into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	patchPayload := dbt_cloud.LineageIntegration{}

	if plan.Host != state.Host {
		patchPayload.Config.Host = plan.Host.ValueString()
	}
	if plan.SiteID != state.SiteID {
		patchPayload.Config.SiteID = plan.SiteID.ValueString()
	}
	if plan.TokenName != state.TokenName {
		patchPayload.Config.TokenName = plan.TokenName.ValueString()
	}
	if plan.Token != state.Token {
		patchPayload.Config.Token = plan.Token.ValueString()
	}

	projectID := state.ProjectID.ValueInt64()
	lineageID := state.LineageIntegrationID.ValueInt64()

	// Update the lineage
	_, err := r.client.UpdateLineageIntegration(projectID, lineageID, patchPayload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating lineage", err.Error())
		return
	}

	// Set the updated state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *lineageIntegrationResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {

	projectID, lineageID, err := helper.SplitIDToInts(req.ID, "lineage")
	if err != nil {
		resp.Diagnostics.AddError("Error splitting the ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("lineage_integration_id"), lineageID,
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("project_id"), projectID,
	)...)
}

func (r *lineageIntegrationResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}
