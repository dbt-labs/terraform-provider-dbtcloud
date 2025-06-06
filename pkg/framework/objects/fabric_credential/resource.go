package fabric_credential

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &fabricCredentialResource{}
	_ resource.ResourceWithConfigure   = &fabricCredentialResource{}
	_ resource.ResourceWithImportState = &fabricCredentialResource{}
)

// FabricCredentialResource is a helper function to simplify the provider implementation.
func FabricCredentialResource() resource.Resource {
	return &fabricCredentialResource{}
}

// fabricCredentialResource is the resource implementation.
type fabricCredentialResource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the resource.
func (r *fabricCredentialResource) Configure(
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
func (r *fabricCredentialResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_fabric_credential"
}

// Schema defines the schema for the resource.
func (r *fabricCredentialResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resourceSchema
}

// Create creates the resource and sets the initial Terraform state.
func (r *fabricCredentialResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Retrieve values from plan
	var plan FabricCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	user := plan.User.ValueString()
	password := plan.Password.ValueString()
	tenantId := plan.TenantId.ValueString()
	clientId := plan.ClientId.ValueString()
	clientSecret := plan.ClientSecret.ValueString()
	schema := plan.Schema.ValueString()
	schemaAuthorization := plan.SchemaAuthorization.ValueString()
	adapterVersion := plan.AdapterVersion.ValueString()

	// Create new credential
	credential, err := r.client.CreateFabricCredential(
		projectID,
		user,
		password,
		tenantId,
		clientId,
		clientSecret,
		schema,
		schemaAuthorization,
		adapterVersion,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Fabric credential",
			"Could not create Fabric credential, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate computed values
	plan.ID = types.StringValue(fmt.Sprintf("%d:%d", projectID, *credential.ID))
	plan.CredentialID = types.Int64Value(int64(*credential.ID))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *fabricCredentialResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state
	var state FabricCredentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get credential from API
	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := r.client.GetFabricCredential(projectID, credentialID)
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading Fabric credential",
			"Could not read Fabric credential ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Refresh state values
	state.Schema = types.StringValue(credential.UnencryptedCredentialDetails.Schema)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *fabricCredentialResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Retrieve values from plan
	var plan FabricCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state FabricCredentialResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())
	user := plan.User.ValueString()
	password := plan.Password.ValueString()
	tenantId := plan.TenantId.ValueString()
	clientId := plan.ClientId.ValueString()
	clientSecret := plan.ClientSecret.ValueString()
	schema := plan.Schema.ValueString()
	schemaAuthorization := plan.SchemaAuthorization.ValueString()
	adapterVersion := plan.AdapterVersion.ValueString()

	// Generate credential details
	credentialDetails, err := dbt_cloud.GenerateFabricCredentialDetails(
		user,
		password,
		tenantId,
		clientId,
		clientSecret,
		schema,
		schemaAuthorization,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Fabric credential",
			"Could not generate credential details: "+err.Error(),
		)
		return
	}

	// Create update object
	updateCredential := dbt_cloud.FabricCredential{
		ID:                &credentialID,
		Account_Id:        r.client.AccountID,
		Project_Id:        projectID,
		Type:              "adapter",
		State:             1,
		Threads:           4,
		CredentialDetails: credentialDetails,
		AdapterVersion:    adapterVersion,
	}

	// Update credential
	_, err = r.client.UpdateFabricCredential(
		projectID,
		credentialID,
		updateCredential,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Fabric credential",
			"Could not update Fabric credential, unexpected error: "+err.Error(),
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
func (r *fabricCredentialResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Retrieve values from state
	var state FabricCredentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete credential
	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	_, err := r.client.DeleteCredential(
		strconv.Itoa(credentialID),
		strconv.Itoa(projectID),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Fabric credential",
			"Could not delete Fabric credential, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *fabricCredentialResource) ImportState(
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
				"Expected import identifier with format: project_id:credential_id. Got: %q",
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

	credentialID, err := strconv.Atoi(idParts[1])
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf(
				"Could not convert credential_id to integer. Got: %q",
				idParts[1],
			),
		)
		return
	}

	// Get credential details from API
	credential, err := r.client.GetFabricCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError("Error getting fabric credential", err.Error())
		return
	}

	// Map response body to schema and populate computed values
	state := FabricCredentialResourceModel{
		ID:                  types.StringValue(fmt.Sprintf("%d:%d", projectID, credentialID)),
		ProjectID:           types.Int64Value(int64(projectID)),
		CredentialID:        types.Int64Value(int64(credentialID)),
		Schema:              types.StringValue(credential.UnencryptedCredentialDetails.Schema),
		SchemaAuthorization: types.StringValue(credential.UnencryptedCredentialDetails.SchemaAuthorization),
		User:                types.StringValue(credential.UnencryptedCredentialDetails.User),
		ClientId:            types.StringValue(credential.UnencryptedCredentialDetails.ClientId),
		TenantId:            types.StringValue(credential.UnencryptedCredentialDetails.TenantId),
		AdapterType:         types.StringValue(credential.AdapterVersion),
	}

	// Set state to fully populated data
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
