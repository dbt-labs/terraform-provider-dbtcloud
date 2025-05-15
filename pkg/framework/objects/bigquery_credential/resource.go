package bigquery_credential

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
	_ resource.Resource                = &bigqueryCredentialResource{}
	_ resource.ResourceWithConfigure   = &bigqueryCredentialResource{}
	_ resource.ResourceWithImportState = &bigqueryCredentialResource{}
)

// BigqueryCredentialResource is a helper function to simplify the provider implementation.
func BigqueryCredentialResource() resource.Resource {
	return &bigqueryCredentialResource{}
}

// bigqueryCredentialResource is the resource implementation.S
type bigqueryCredentialResource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the resource.
func (r *bigqueryCredentialResource) Configure(
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
func (r *bigqueryCredentialResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_bigquery_credential"
}

// Schema defines the schema for the resource.
func (r *bigqueryCredentialResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = BigQueryResourceSchema
}

// Create creates the resource and sets the initial Terraform state.
func (r *bigqueryCredentialResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Retrieve values from plan
	var plan BigqueryCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	isActive := plan.IsActive.ValueBool()
	projectID := int(plan.ProjectID.ValueInt64())
	dataset := plan.Dataset.ValueString()
	numThreads := int(plan.NumThreads.ValueInt64())

	// Create new credential
	credential, err := r.client.CreateBigQueryCredential(
		projectID,
		"bigquery",
		isActive,
		dataset,
		numThreads,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Bigquery credential",
			"Could not create Bigquery credential, unexpected error: "+err.Error(),
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
func (r *bigqueryCredentialResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state
	var state BigqueryCredentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get credential from API
	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := r.client.GetBigQueryCredential(projectID, credentialID)
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading Bigquery credential",
			"Could not read Bigquery credential ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Refresh state values
	state.CredentialID = types.Int64Value(int64(*credential.ID))
	state.IsActive = types.BoolValue(credential.State == dbt_cloud.STATE_ACTIVE)
	state.ProjectID = types.Int64Value(int64(credential.Project_Id))
	state.Dataset = types.StringValue(credential.Dataset)
	state.NumThreads = types.Int64Value(int64(credential.Threads))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *bigqueryCredentialResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Retrieve values from plan
	var plan BigqueryCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state BigqueryCredentialResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())
	dataset := plan.Dataset.ValueString()
	numThreads := int(plan.NumThreads.ValueInt64())

	if (state.Dataset.ValueString() != dataset) || (state.NumThreads.ValueInt64() != int64(numThreads)) {
		credential, err := r.client.GetBigQueryCredential(projectID, credentialID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading Bigquery credential",
				"Could not read Bigquery credential ID "+state.ID.ValueString()+": "+err.Error(),
			)
			return
		}

		if state.Dataset.ValueString() != dataset {
			credential.Dataset = dataset
		}
		if state.NumThreads.ValueInt64() != int64(numThreads) {
			credential.Threads = numThreads
		}

		_, err = r.client.UpdateBigQueryCredential(
			projectID,
			credentialID,
			*credential,
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating Bigquery credential",
				"Could not update Bigquery credential, unexpected error: "+err.Error(),
			)
		}
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *bigqueryCredentialResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Retrieve values from state
	var state BigqueryCredentialResourceModel
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
			"Error deleting Bigquery credential",
			"Could not delete Bigquery credential, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *bigqueryCredentialResource) ImportState(
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
}
