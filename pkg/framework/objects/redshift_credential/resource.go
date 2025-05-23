package redshift_credential

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
	_ resource.Resource                = &redshiftCredentialResource{}
	_ resource.ResourceWithConfigure   = &redshiftCredentialResource{}
	_ resource.ResourceWithImportState = &redshiftCredentialResource{}
)

// RedshiftCredentialResourceModel is a helper function to simplify the provider implementation.
func RedshiftCredentialResource() resource.Resource {
	return &redshiftCredentialResource{}
}

// RedshiftCredentialResourceModel is the resource implementation.S
type redshiftCredentialResource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the resource.
func (r *redshiftCredentialResource) Configure(
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
func (r *redshiftCredentialResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_redshift_credential"
}

// Schema defines the schema for the resource.
func (r *redshiftCredentialResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = RedshiftResourceSchema
}

// Create creates the resource and sets the initial Terraform state.
func (r *redshiftCredentialResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Retrieve values from plan
	var plan RedshiftCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	isActive := plan.IsActive.ValueBool()
	projectID := int(plan.ProjectID.ValueInt64())
	dataset := plan.Schema.ValueString()
	numThreads := int(plan.NumThreads.ValueInt64())

	// Create new credential
	credential, err := r.client.CreateRedshiftCredential(
		projectID,
		"redshift",
		isActive,
		dataset,
		numThreads,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Redshift credential",
			"Could not create Redshift credential, unexpected error: "+err.Error(),
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
func (r *redshiftCredentialResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state
	var state RedshiftCredentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get credential from API
	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := r.client.GetRedshiftCredential(projectID, credentialID)
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading Redshift credential",
			"Could not read Redshift credential ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Refresh state values
	state.CredentialID = types.Int64Value(int64(*credential.ID))
	state.IsActive = types.BoolValue(credential.State == dbt_cloud.STATE_ACTIVE)
	state.ProjectID = types.Int64Value(int64(credential.Project_Id))
	state.Schema = types.StringValue(credential.Schema)
	state.NumThreads = types.Int64Value(int64(credential.Threads))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *redshiftCredentialResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Retrieve values from plan
	var plan RedshiftCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state RedshiftCredentialResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())
	dataset := plan.Schema.ValueString()
	numThreads := int(plan.NumThreads.ValueInt64())

	if (state.Schema.ValueString() != dataset) || (state.NumThreads.ValueInt64() != int64(numThreads)) {
		credential, err := r.client.GetRedshiftCredential(projectID, credentialID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading Redshift credential",
				"Could not read Redshift credential ID "+state.ID.ValueString()+": "+err.Error(),
			)
			return
		}

		if state.Schema.ValueString() != dataset {
			credential.Schema = dataset
		}
		if state.NumThreads.ValueInt64() != int64(numThreads) {
			credential.Threads = numThreads
		}

		_, err = r.client.UpdateRedshiftCredential(
			projectID,
			credentialID,
			*credential,
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating Redshift credential",
				"Could not update Redshift credential, unexpected error: "+err.Error(),
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
func (r *redshiftCredentialResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Retrieve values from state
	var state RedshiftCredentialResourceModel
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
			"Error deleting Redshift credential",
			"Could not delete Redshift credential, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *redshiftCredentialResource) ImportState(
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
