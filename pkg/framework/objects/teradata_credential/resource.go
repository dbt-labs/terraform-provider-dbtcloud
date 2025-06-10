package teradata_credential

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

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &teradataCredentialResource{}
	_ resource.ResourceWithConfigure   = &teradataCredentialResource{}
	_ resource.ResourceWithImportState = &teradataCredentialResource{}
)

// TeradataCredentialResource is a helper function to simplify the provider implementation.
func TeradataCredentialResource() resource.Resource {
	return &teradataCredentialResource{}
}

// teradataCredentialResource is the resource implementation.
type teradataCredentialResource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the resource.
func (r *teradataCredentialResource) Configure(
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
func (r *teradataCredentialResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_teradata_credential"
}

// Schema defines the schema for the resource.
func (r *teradataCredentialResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resourceSchema
}

// Create creates the resource and sets the initial Terraform state.
func (r *teradataCredentialResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Retrieve values from plan
	var plan TeradataCredentialModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := plan.User.ValueString()
	password := plan.Password.ValueString()
	schema := plan.Schema.ValueString()
	projectID := int(plan.ProjectID.ValueInt64())
	threads := plan.Threads.ValueInt64()
	adapterVersion := plan.AdapterVersion.ValueString()

	// Create new credential
	credential, err := r.client.CreateTeradataCredential(
		ctx,
		projectID,
		user,
		password,
		schema,
		int(threads),
		adapterVersion,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Teradata credential",
			"Could not create Teradata credential, unexpected error: "+err.Error(),
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
func (r *teradataCredentialResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state
	var state TeradataCredentialModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get credential from API
	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := r.client.GetTeradataCredential(projectID, credentialID)
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading Teradata credential",
			"Could not read Teradata credential ID "+state.ID.ValueString()+": "+err.Error(),
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
func (r *teradataCredentialResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Retrieve values from plan
	var plan TeradataCredentialModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state TeradataCredentialModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())
	user := plan.User.ValueString()
	password := plan.Password.ValueString()
	schema := plan.Schema.ValueString()
	adapterVersion := plan.AdapterVersion.ValueString()

	// Generate credential details
	credentialDetails, err := dbt_cloud.GenerateTeradataCredentialDetails(
		user,
		password,
		schema,
		int(plan.Threads.ValueInt64()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Teradata credential",
			"Could not generate credential details: "+err.Error(),
		)
		return
	}

	// Create update object
	updateCredential := dbt_cloud.TeradataCredentialRequest{
		ID:                &credentialID,
		AccountID:         r.client.AccountID,
		ProjectID:         projectID,
		Type:              "adapter",
		State:             1,
		Threads:           4,
		CredentialDetails: credentialDetails,
		AdapterVersion:    adapterVersion,
	}

	// Update credential
	_, err = r.client.UpdateTeradataCredential(
		projectID,
		credentialID,
		updateCredential,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Teradata credential",
			"Could not update Teradata credential, unexpected error: "+err.Error(),
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
func (r *teradataCredentialResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Retrieve values from state
	var state TeradataCredentialModel
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
			"Error deleting Teradata credential",
			"Could not delete Teradata credential, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *teradataCredentialResource) ImportState(
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

	credentialResponse, err := r.client.GetTeradataCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError("Error getting teradata credential", err.Error())
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
		path.Root("schema"),
		credentialResponse.UnencryptedCredentialDetails.Schema,
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("threads"),
		credentialResponse.Threads,
	)...)
}
