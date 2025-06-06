package synapse_credential

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
	_ resource.Resource                = &synapseCredentialResource{}
	_ resource.ResourceWithConfigure   = &synapseCredentialResource{}
	_ resource.ResourceWithImportState = &synapseCredentialResource{}
)

// SynapseCredentialResource is a helper function to simplify the provider implementation.
func SynapseCredentialResource() resource.Resource {
	return &synapseCredentialResource{}
}

// synapseCredentialResource is the resource implementation.
type synapseCredentialResource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the resource.
func (r *synapseCredentialResource) Configure(
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
func (r *synapseCredentialResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_synapse_credential"
}

// Schema defines the schema for the resource.
func (r *synapseCredentialResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resourceSchema
}

// Create creates the resource and sets the initial Terraform state.
func (r *synapseCredentialResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Retrieve values from plan
	var plan SynapseCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	authentication := plan.Authentication.ValueString()
	user := plan.User.ValueString()
	password := plan.Password.ValueString()
	tenantId := plan.TenantId.ValueString()
	clientId := plan.ClientId.ValueString()
	clientSecret := plan.ClientSecret.ValueString()
	schema := plan.Schema.ValueString()
	schemaAuthorization := plan.SchemaAuthorization.ValueString()
	adapterVersion := plan.AdapterVersion.ValueString()

	if (authentication == "ServicePrincipal" && user != "") || (authentication != "ServicePrincipal" && user == "") {
		resp.Diagnostics.AddError(
			"Error creating Synapse credential",
			"Could not validate authentication method, please check the configuration for "+authentication+". ",
		)
		return
	}

	// Create new credential
	credential, err := r.client.CreateSynapseCredential(
		projectID,
		authentication,
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
			"Error creating Synapse credential",
			"Could not create Synapse credential, unexpected error: "+err.Error(),
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
func (r *synapseCredentialResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state
	var state SynapseCredentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get credential from API
	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := r.client.GetSynapseCredential(projectID, credentialID)
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading Synapse credential",
			"Could not read Synapse credential ID "+state.ID.ValueString()+": "+err.Error(),
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
func (r *synapseCredentialResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Retrieve values from plan
	var plan SynapseCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state SynapseCredentialResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())
	authentication := plan.Authentication.ValueString()
	user := plan.User.ValueString()
	password := plan.Password.ValueString()
	tenantId := plan.TenantId.ValueString()
	clientId := plan.ClientId.ValueString()
	clientSecret := plan.ClientSecret.ValueString()
	schema := plan.Schema.ValueString()
	schemaAuthorization := plan.SchemaAuthorization.ValueString()

	if (authentication == "ServicePrincipal" && user != "") || (authentication != "ServicePrincipal" && user == "") {
		resp.Diagnostics.AddError(
			"Error creating Synapse credential",
			"Could not validate authentication method, please check the configuration for "+authentication+". ",
		)
		return
	}

	// Generate credential details
	credentialDetails, err := dbt_cloud.GenerateSynapseCredentialDetails(
		authentication,
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
			"Error updating Synapse credential",
			"Could not generate credential details: "+err.Error(),
		)
		return
	}

	// Create update object
	updateCredential := dbt_cloud.SynapseCredential{
		ID:                &credentialID,
		Account_Id:        r.client.AccountID,
		Project_Id:        projectID,
		Type:              "adapter",
		State:             1,
		Threads:           4,
		CredentialDetails: credentialDetails,
		AdapterVersion:    "synapse_v0",
	}

	// Update credential
	_, err = r.client.UpdateSynapseCredential(
		projectID,
		credentialID,
		updateCredential,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Synapse credential",
			"Could not update Synapse credential, unexpected error: "+err.Error(),
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
func (r *synapseCredentialResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Retrieve values from state
	var state SynapseCredentialResourceModel
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
			"Error deleting Synapse credential",
			"Could not delete Synapse credential, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *synapseCredentialResource) ImportState(
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
	credential, err := r.client.GetSynapseCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError("Error getting synapse credential", err.Error())
		return
	}

	// Map response body to schema and populate computed values
	state := SynapseCredentialResourceModel{
		ID:                  types.StringValue(fmt.Sprintf("%d:%d", projectID, credentialID)),
		ProjectID:           types.Int64Value(int64(projectID)),
		CredentialID:        types.Int64Value(int64(credentialID)),
		Authentication:      types.StringValue(credential.UnencryptedCredentialDetails.Authentication),
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
