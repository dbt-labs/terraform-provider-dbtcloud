package salesforce_credential

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
	_ resource.Resource                = &salesforceCredentialResource{}
	_ resource.ResourceWithConfigure   = &salesforceCredentialResource{}
	_ resource.ResourceWithImportState = &salesforceCredentialResource{}
)

func SalesforceCredentialResource() resource.Resource {
	return &salesforceCredentialResource{}
}

type salesforceCredentialResource struct {
	client *dbt_cloud.Client
}

func (r *salesforceCredentialResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_salesforce_credential"
}

func (r *salesforceCredentialResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = SalesforceResourceSchema
}

func (r *salesforceCredentialResource) Configure(
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

func (r *salesforceCredentialResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SalesforceCredentialResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	username := plan.Username.ValueString()
	clientID := plan.ClientID.ValueString()
	privateKey := plan.PrivateKey.ValueString()
	targetName := plan.TargetName.ValueString()
	numThreads := int(plan.NumThreads.ValueInt64())

	credential, err := r.client.CreateSalesforceCredential(
		projectID,
		username,
		clientID,
		privateKey,
		targetName,
		numThreads,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Salesforce credential",
			"Could not create Salesforce credential: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d:%d", projectID, *credential.ID))
	plan.CredentialID = types.Int64Value(int64(*credential.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *salesforceCredentialResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state SalesforceCredentialResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := r.client.GetSalesforceCredential(projectID, credentialID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The Salesforce credential was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting Salesforce credential", err.Error())
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%d:%d", projectID, *credential.ID))
	state.CredentialID = types.Int64Value(int64(*credential.ID))
	state.Username = types.StringValue(credential.UnencryptedCredentialDetails.Username)
	state.TargetName = types.StringValue(credential.UnencryptedCredentialDetails.TargetName)
	state.NumThreads = types.Int64Value(int64(credential.UnencryptedCredentialDetails.Threads))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *salesforceCredentialResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state SalesforceCredentialResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credentialDetails, err := dbt_cloud.GenerateSalesforceCredentialDetails(
		plan.Username.ValueString(),
		plan.ClientID.ValueString(),
		plan.PrivateKey.ValueString(),
		plan.TargetName.ValueString(),
		int(plan.NumThreads.ValueInt64()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error generating Salesforce credential details",
			err.Error(),
		)
		return
	}

	salesforceCredential := dbt_cloud.SalesforceCredentialGlobConnPatch{
		ID:                credentialID,
		CredentialDetails: credentialDetails,
	}

	_, err = r.client.UpdateSalesforceCredential(projectID, credentialID, salesforceCredential)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Salesforce credential",
			err.Error(),
		)
		return
	}

	plan.ID = state.ID
	plan.CredentialID = state.CredentialID

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *salesforceCredentialResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SalesforceCredentialResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	_, err := r.client.DeleteCredential(
		strconv.Itoa(credentialID),
		strconv.Itoa(projectID),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Salesforce credential",
			"Could not delete Salesforce credential: "+err.Error(),
		)
		return
	}
}

func (r *salesforceCredentialResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	idParts := strings.Split(req.ID, ":")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: project_id:credential_id. Got: %q", req.ID),
		)
		return
	}

	projectID, err := strconv.Atoi(idParts[0])
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Could not convert project_id to integer. Got: %q", idParts[0]),
		)
		return
	}

	credentialID, err := strconv.Atoi(idParts[1])
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Could not convert credential_id to integer. Got: %q", idParts[1]),
		)
		return
	}

	credential, err := r.client.GetSalesforceCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError("Error getting Salesforce credential", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("id"),
		fmt.Sprintf("%d:%d", projectID, credentialID),
	)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("project_id"),
		projectID,
	)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("credential_id"),
		credentialID,
	)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("username"),
		credential.UnencryptedCredentialDetails.Username,
	)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("target_name"),
		credential.UnencryptedCredentialDetails.TargetName,
	)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("num_threads"),
		credential.UnencryptedCredentialDetails.Threads,
	)...)
}
