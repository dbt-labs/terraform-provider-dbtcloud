package platform_metadata_credentials

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

var (
	_ resource.Resource                = &databricksPlatformMetadataCredentialResource{}
	_ resource.ResourceWithConfigure   = &databricksPlatformMetadataCredentialResource{}
	_ resource.ResourceWithImportState = &databricksPlatformMetadataCredentialResource{}
)

// DatabricksPlatformMetadataCredentialResource returns a new resource instance
func DatabricksPlatformMetadataCredentialResource() resource.Resource {
	return &databricksPlatformMetadataCredentialResource{}
}

type databricksPlatformMetadataCredentialResource struct {
	client *dbt_cloud.Client
}

func (r *databricksPlatformMetadataCredentialResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_databricks_platform_metadata_credential"
}

func (r *databricksPlatformMetadataCredentialResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = DatabricksPlatformMetadataCredentialSchema
}

func (r *databricksPlatformMetadataCredentialResource) Configure(
	_ context.Context,
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

func (r *databricksPlatformMetadataCredentialResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan DatabricksPlatformMetadataCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the API request
	credential := dbt_cloud.PlatformMetadataCredential{
		ConnectionID:            plan.ConnectionID.ValueInt64(),
		CatalogIngestionEnabled: plan.CatalogIngestionEnabled.ValueBool(),
		CostOptimizationEnabled: plan.CostOptimizationEnabled.ValueBool(),
		CostInsightsEnabled:     plan.CostInsightsEnabled.ValueBool(),
		Config: dbt_cloud.PlatformMetadataCredentialConfig{
			Token:   plan.Token.ValueString(),
			Catalog: plan.Catalog.ValueString(),
		},
	}

	// Create the credential
	created, err := r.client.CreatePlatformMetadataCredential(credential)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Databricks platform metadata credential",
			"Could not create Databricks platform metadata credential: "+err.Error(),
		)
		return
	}

	// Update plan with computed values
	plan.ID = types.StringValue(fmt.Sprintf("%d:%d", r.client.AccountID, *created.ID))
	plan.CredentialID = types.Int64Value(*created.ID)
	plan.AdapterVersion = types.StringValue(created.AdapterVersion)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *databricksPlatformMetadataCredentialResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state DatabricksPlatformMetadataCredentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialID := state.CredentialID.ValueInt64()

	credential, err := r.client.GetPlatformMetadataCredential(credentialID)
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading Databricks platform metadata credential",
			"Could not read Databricks platform metadata credential ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state with values from API
	state.ConnectionID = types.Int64Value(credential.ConnectionID)
	state.CatalogIngestionEnabled = types.BoolValue(credential.CatalogIngestionEnabled)
	state.CostOptimizationEnabled = types.BoolValue(credential.CostOptimizationEnabled)
	state.CostInsightsEnabled = types.BoolValue(credential.CostInsightsEnabled)
	state.AdapterVersion = types.StringValue(credential.AdapterVersion)

	// Update non-sensitive config fields from API
	// Note: Token is returned masked, so we preserve the plan-time value
	if credential.Config.Catalog != "" {
		state.Catalog = types.StringValue(credential.Config.Catalog)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *databricksPlatformMetadataCredentialResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan DatabricksPlatformMetadataCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DatabricksPlatformMetadataCredentialResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialID := state.CredentialID.ValueInt64()

	// Build the update request
	// Note: connection_id is immutable and should not be included in update requests
	credential := dbt_cloud.PlatformMetadataCredential{
		CatalogIngestionEnabled: plan.CatalogIngestionEnabled.ValueBool(),
		CostOptimizationEnabled: plan.CostOptimizationEnabled.ValueBool(),
		CostInsightsEnabled:     plan.CostInsightsEnabled.ValueBool(),
		Config: dbt_cloud.PlatformMetadataCredentialConfig{
			Token:   plan.Token.ValueString(),
			Catalog: plan.Catalog.ValueString(),
		},
	}

	// Update the credential
	updated, err := r.client.UpdatePlatformMetadataCredential(credentialID, credential)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Databricks platform metadata credential",
			"Could not update Databricks platform metadata credential: "+err.Error(),
		)
		return
	}

	// Update plan with computed values
	plan.AdapterVersion = types.StringValue(updated.AdapterVersion)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *databricksPlatformMetadataCredentialResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state DatabricksPlatformMetadataCredentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialID := state.CredentialID.ValueInt64()

	err := r.client.DeletePlatformMetadataCredential(credentialID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Databricks platform metadata credential",
			"Could not delete Databricks platform metadata credential: "+err.Error(),
		)
		return
	}
}

func (r *databricksPlatformMetadataCredentialResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Support both "credential_id" and "account_id:credential_id" formats
	idParts := strings.Split(req.ID, ":")

	var credentialID int
	var err error

	if len(idParts) == 1 {
		// Just credential_id
		credentialID, err = strconv.Atoi(idParts[0])
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				fmt.Sprintf("Could not parse credential_id as integer: %s", idParts[0]),
			)
			return
		}
	} else if len(idParts) == 2 {
		// account_id:credential_id format
		credentialID, err = strconv.Atoi(idParts[1])
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				fmt.Sprintf("Could not parse credential_id as integer: %s", idParts[1]),
			)
			return
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: credential_id or account_id:credential_id. Got: %s", req.ID),
		)
		return
	}

	// Set the ID and credential_id for the Read operation
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("id"),
		fmt.Sprintf("%d:%d", r.client.AccountID, credentialID),
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("credential_id"),
		credentialID,
	)...)
}
