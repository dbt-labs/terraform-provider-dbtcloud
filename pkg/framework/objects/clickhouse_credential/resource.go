package clickhouse_credential

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &clickhouseCredentialResource{}
	_ resource.ResourceWithConfigure   = &clickhouseCredentialResource{}
	_ resource.ResourceWithImportState = &clickhouseCredentialResource{}
)

func ClickhouseCredentialResource() resource.Resource {
	return &clickhouseCredentialResource{}
}

type clickhouseCredentialResource struct {
	client *dbt_cloud.Client
}

func (r *clickhouseCredentialResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_clickhouse_credential"
}

func (r *clickhouseCredentialResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = ClickhouseResourceSchema
}

func (r *clickhouseCredentialResource) Configure(
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

func (r *clickhouseCredentialResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ClickhouseCredentialResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve config to access write-only attributes
	var config ClickhouseCredentialResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	user := plan.User.ValueString()
	password := helper.ResolveWriteOnlyString(config.PasswordWo, plan.Password)
	schema := plan.Schema.ValueString()
	targetName := plan.TargetName.ValueString()
	threads := threadsPointer(plan.Threads)

	credential, err := r.client.CreateClickhouseCredential(
		projectID,
		user,
		password,
		schema,
		targetName,
		threads,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ClickHouse credential",
			"Could not create ClickHouse credential: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d:%d", projectID, *credential.ID))
	plan.CredentialID = types.Int64Value(int64(*credential.ID))
	plan.Threads = threadsInt64Value(threads)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clickhouseCredentialResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state ClickhouseCredentialResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := r.client.GetClickhouseCredential(projectID, credentialID)
	if err != nil {
		if helper.HandleResourceNotFound(ctx, err, &resp.Diagnostics, &resp.State, "clickhouse credential") {
			return
		}
		resp.Diagnostics.AddError("Error getting ClickHouse credential", err.Error())
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%d:%d", projectID, *credential.ID))
	state.CredentialID = types.Int64Value(int64(*credential.ID))
	state.User = types.StringValue(credential.UnencryptedCredentialDetails.User)
	state.Schema = types.StringValue(credential.UnencryptedCredentialDetails.Schema)
	state.TargetName = types.StringValue(credential.UnencryptedCredentialDetails.TargetName)
	state.Threads = threadsInt64Value(credential.UnencryptedCredentialDetails.Threads)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clickhouseCredentialResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state ClickhouseCredentialResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve config to access write-only attributes
	var config ClickhouseCredentialResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	password := helper.ResolveWriteOnlyString(config.PasswordWo, plan.Password)

	credentialDetails, err := dbt_cloud.GenerateClickhouseCredentialDetails(
		plan.User.ValueString(),
		password,
		plan.Schema.ValueString(),
		plan.TargetName.ValueString(),
		threadsPointer(plan.Threads),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error generating ClickHouse credential details",
			err.Error(),
		)
		return
	}

	clickhouseCredential := dbt_cloud.ClickhouseCredentialGlobConnPatch{
		ID:                credentialID,
		CredentialDetails: credentialDetails,
	}

	_, err = r.client.UpdateClickhouseCredential(projectID, credentialID, clickhouseCredential)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ClickHouse credential",
			err.Error(),
		)
		return
	}

	credential, err := r.client.GetClickhouseCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading back ClickHouse credential after update",
			err.Error(),
		)
		return
	}

	plan.ID = state.ID
	plan.CredentialID = state.CredentialID
	plan.User = types.StringValue(credential.UnencryptedCredentialDetails.User)
	plan.Schema = types.StringValue(credential.UnencryptedCredentialDetails.Schema)
	plan.TargetName = types.StringValue(credential.UnencryptedCredentialDetails.TargetName)
	plan.Threads = threadsInt64Value(credential.UnencryptedCredentialDetails.Threads)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clickhouseCredentialResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state ClickhouseCredentialResourceModel

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
			"Error deleting ClickHouse credential",
			"Could not delete ClickHouse credential: "+err.Error(),
		)
		return
	}
}

func (r *clickhouseCredentialResource) ImportState(
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

	credential, err := r.client.GetClickhouseCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError("Error getting ClickHouse credential", err.Error())
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
		path.Root("user"),
		credential.UnencryptedCredentialDetails.User,
	)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("schema"),
		credential.UnencryptedCredentialDetails.Schema,
	)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("target_name"),
		credential.UnencryptedCredentialDetails.TargetName,
	)...)

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("threads"),
		threadsInt64Value(credential.UnencryptedCredentialDetails.Threads),
	)...)
}

// threadsPointer converts a possibly-null/unknown Terraform Int64 into a *int,
// so an unset `threads` is sent to the API as null instead of defaulting to 0.
func threadsPointer(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	t := int(v.ValueInt64())
	return &t
}

// threadsInt64Value converts the API's nullable threads back into a Terraform
// Int64, preserving null when the API has no value set.
func threadsInt64Value(threads *int) types.Int64 {
	if threads == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*threads))
}
