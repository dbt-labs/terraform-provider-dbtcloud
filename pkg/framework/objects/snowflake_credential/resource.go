package snowflake_credential

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
	_ resource.Resource                = &snowflakeCredentialResource{}
	_ resource.ResourceWithConfigure   = &snowflakeCredentialResource{}
	_ resource.ResourceWithImportState = &snowflakeCredentialResource{}
)

// SnowflakeCredentialResource is a helper function to simplify the provider implementation.
func SnowflakeCredentialResource() resource.Resource {
	return &snowflakeCredentialResource{}
}

// snowflakeCredentialResource is the resource implementation.
type snowflakeCredentialResource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the resource.
func (r *snowflakeCredentialResource) Configure(
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
func (r *snowflakeCredentialResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_snowflake_credential"
}

// Schema defines the schema for the resource.
func (r *snowflakeCredentialResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resourceSchema
}

// Create creates the resource and sets the initial Terraform state.
func (r *snowflakeCredentialResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Retrieve values from plan
	var plan SnowflakeCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	is_active := plan.State.ValueInt64() == 1
	project_id := int(plan.ProjectID.ValueInt64())
	auth_type := plan.AuthType.ValueString()
	database := plan.Database.ValueString()
	role := plan.Role.ValueString()
	warehouse := plan.Warehouse.ValueString()
	schema := plan.Schema.ValueString()
	user := plan.User.ValueString()
	password := plan.Password.ValueString()
	private_key := plan.PrivateKey.ValueString()
	private_key_passphrase := plan.PrivateKeyPassphrase.ValueString()
	num_threads := int(plan.Threads.ValueInt64())

	// Create new credential
	credential, err := r.client.CreateSnowflakeCredential(
		project_id,
		"snowflake",
		is_active,
		database,
		role,
		warehouse,
		schema,
		user,
		password,
		private_key,
		private_key_passphrase,
		auth_type,
		num_threads,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Snowflake credential",
			"Could not create Snowflake credential, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate computed values
	plan.ID = types.StringValue(fmt.Sprintf("%d:%d", project_id, *credential.ID))
	plan.CredentialID = types.Int64Value(int64(*credential.ID))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *snowflakeCredentialResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state
	var state SnowflakeCredentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get credential from API
	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := r.client.GetSnowflakeCredential(projectID, credentialID)
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading Snowflake credential",
			"Could not read Snowflake credential ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	if credential.Auth_Type == "password" {
		state.Password = types.StringValue(credential.Password)
	}
	if credential.Auth_Type == "keypair" {
		state.PrivateKey = types.StringValue(credential.PrivateKey)
		state.PrivateKeyPassphrase = types.StringValue(credential.PrivateKeyPassphrase)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *snowflakeCredentialResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Retrieve values from plan
	var plan SnowflakeCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state SnowflakeCredentialResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	if (state.AuthType != plan.AuthType) ||
		(state.Database != plan.Database) ||
		(state.Role != plan.Role) ||
		(state.Warehouse != plan.Warehouse) ||
		(state.Schema != plan.Schema) ||
		(state.User != plan.User) ||
		(state.Password != plan.Password) ||
		(state.PrivateKey != plan.PrivateKey) ||
		(state.PrivateKeyPassphrase != plan.PrivateKeyPassphrase) {
		credential, err := r.client.GetSnowflakeCredential(projectID, credentialID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading Snowflake credential",
				"Could not read Snowflake credential ID "+state.ID.ValueString()+": "+err.Error(),
			)
			return
		}

		credential.Auth_Type = plan.AuthType.ValueString()
		credential.Database = plan.Database.ValueString()
		credential.Role = plan.Role.ValueString()
		credential.Warehouse = plan.Warehouse.ValueString()
		credential.Schema = plan.Schema.ValueString()
		credential.User = plan.User.ValueString()
		credential.Password = plan.Password.ValueString()
		credential.PrivateKey = plan.PrivateKey.ValueString()
		credential.PrivateKeyPassphrase = plan.PrivateKeyPassphrase.ValueString()
		credential.Threads = int(plan.Threads.ValueInt64())

		_, err = r.client.UpdateSnowflakeCredential(projectID, credentialID, *credential)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating Snowflake credential",
				"Could not update Snowflake credential, unexpected error: "+err.Error(),
			)
			return
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
func (r *snowflakeCredentialResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Retrieve values from state
	var state SnowflakeCredentialResourceModel
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
			"Error deleting Snowflake credential",
			"Could not delete Snowflake credential, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *snowflakeCredentialResource) ImportState(
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
