package postgres_credential

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
	_ resource.Resource                = &postgresCredentialResource{}
	_ resource.ResourceWithConfigure   = &postgresCredentialResource{}
	_ resource.ResourceWithImportState = &postgresCredentialResource{}
)

func PostgresCredentialResource() resource.Resource {
	return &postgresCredentialResource{}
}

type postgresCredentialResource struct {
	client *dbt_cloud.Client
}

func (p *postgresCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

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

	credential, err := p.client.GetPostgresCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Postgres credential",
			"Could not read imported Postgres credential: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("is_active"),
		credential.State == dbt_cloud.STATE_ACTIVE,
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("default_schema"),
		credential.Default_Schema,
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("username"),
		credential.Username,
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("num_threads"),
		credential.Threads,
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("type"),
		credential.Type,
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("target_name"),
		credential.Target_Name,
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("password"),
		credential.Password,
	)...)
}

func (p *postgresCredentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dbt_cloud.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *dbt_cloud.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	p.client = client
}

func (p *postgresCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PostgresCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	username := plan.Username.ValueString()
	password := plan.Password.ValueString()
	defaultSchema := plan.DefaultSchema.ValueString()
	numThreads := int(plan.NumThreads.ValueInt64())
	type_value := plan.Type.ValueString()
	targetName := plan.TargetName.ValueString()
	isActive := true // Default to active for new credentials

	if !plan.IsActive.IsNull() {
		isActive = plan.IsActive.ValueBool()
	}

	credential, err := p.client.CreatePostgresCredential(
		projectID,
		isActive,
		type_value,
		defaultSchema, 
		targetName,
		username, 
		password, 
		numThreads)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Postgres credential",
			"Could not create Postgres credential, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d:%d", credential.Project_Id, *credential.ID))
	plan.CredentialID = types.Int64Value(int64(*credential.ID))
	plan.IsActive = types.BoolValue(credential.State == dbt_cloud.STATE_ACTIVE)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *postgresCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PostgresCredentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	_, err := p.client.DeletePostgresCredential(strconv.Itoa(credentialID), strconv.Itoa(projectID))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Postgres credential",
			"Could not delete Postgres credential, unexpected error: "+err.Error(),
		)
		return
	}
}

func (p *postgresCredentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres_credential"
}

func (p *postgresCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PostgresCredentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := p.client.GetPostgresCredential(projectID, credentialID)
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading Starburst credential",
			"Could not read Starburst credential ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ProjectID = types.Int64Value(int64(credential.Project_Id))
	state.CredentialID = types.Int64Value(int64(*credential.ID))
	state.IsActive = types.BoolValue(credential.State == dbt_cloud.STATE_ACTIVE)
	state.DefaultSchema = types.StringValue(credential.Default_Schema)
	state.Username = types.StringValue(credential.Username)
	state.NumThreads = types.Int64Value(int64(credential.Threads))
	state.Type = types.StringValue(credential.Type)
	state.TargetName = types.StringValue(credential.Target_Name)
	state.Password = types.StringValue(credential.Password)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (p *postgresCredentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema
}

func (p *postgresCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PostgresCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PostgresCredentialResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	credentialID := int(plan.CredentialID.ValueInt64())

	credential, err := p.client.GetPostgresCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Postgres credential",
			"Could not get Postgres credential with ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	credential.Type = plan.Type.ValueString()
	credential.Default_Schema = plan.DefaultSchema.ValueString()
	credential.Target_Name = plan.TargetName.ValueString()
	credential.Username = plan.Username.ValueString()
	credential.Password = plan.Password.ValueString()
	credential.Threads = int(plan.NumThreads.ValueInt64())

	if plan.IsActive.ValueBool() {
		credential.State = dbt_cloud.STATE_ACTIVE
	} else {
		credential.State = dbt_cloud.STATE_DELETED
	}

	_, err = p.client.UpdatePostgresCredential(projectID, credentialID, *credential)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Postgres credential",
			"Could not update Postgres credential with ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d%s%d", credential.Project_Id, dbt_cloud.ID_DELIMITER, *credential.ID))

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
