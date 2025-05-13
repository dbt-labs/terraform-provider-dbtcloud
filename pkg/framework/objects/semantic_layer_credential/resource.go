package semantic_layer_credential

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &snowflakeSemanticLayerCredentialResource{}
	_ resource.ResourceWithConfigure = &snowflakeSemanticLayerCredentialResource{}
)

func SnowflakeSemanticLayerCredentialResource() resource.Resource {
	return &snowflakeSemanticLayerCredentialResource{}
}

// dbtCloud.Client for making API calls
type snowflakeSemanticLayerCredentialResource struct {
	client *dbt_cloud.Client
}

func (r *snowflakeSemanticLayerCredentialResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_snowflake_semantic_layer_credential"
}

func (r *snowflakeSemanticLayerCredentialResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state SnowflakeSLCredentialModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	id := state.ID.ValueInt64()

	credential, err := r.client.GetSemanticLayerCredential(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Issue getting Semantic Layer credential",
			"Error: "+err.Error(),
		)
		return
	}

	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The Semantic Layer credential was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error getting the Semantic Layer configuration", err.Error())
		return
	}

	state.ID = types.Int64Value(int64(*credential.ID))
	state.Credential.ProjectID = types.Int64Value(int64(credential.ProjectID))
	state.Credential.CredentialID = types.Int64Value(int64(*credential.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *snowflakeSemanticLayerCredentialResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan SnowflakeSLCredentialModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.Credential.ProjectID.ValueInt64()

	createdCredential, err := r.client.CreateSemanticLayerCredential(
		projectID,
		plan.Credential.IsActive.ValueBool(),
		plan.Credential.Database.ValueString(),
		plan.Credential.Role.ValueString(),
		plan.Credential.Warehouse.ValueString(),
		plan.Credential.Schema.ValueString(),
		plan.Credential.User.ValueString(),
		plan.Credential.Password.ValueString(),
		plan.Credential.PrivateKey.ValueString(),
		plan.Credential.PrivateKeyPassphrase.ValueString(),
		plan.Credential.AuthType.ValueString(),
		int(plan.Credential.NumThreads.ValueInt64()),
		plan.Configuration.Name.ValueString(),
		plan.Configuration.AdapterVersion.ValueString(),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Semantic Layer configuration",
			"Error: "+err.Error(),
		)
		return
	}

	plan.ID = types.Int64Value(int64(*createdCredential.ID))

	//snowflake credential ids, not used in this case
	plan.Credential.CredentialID = types.Int64Value(int64(*createdCredential.ID))
	plan.Credential.ID = types.StringValue(fmt.Sprintf("%d", *createdCredential.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *snowflakeSemanticLayerCredentialResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state SnowflakeSLCredentialModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID
	projectID := state.Credential.ProjectID.ValueInt64()

	err := r.client.DeleteSemanticLayerCredential(projectID, id.ValueInt64())

	if err != nil {
		resp.Diagnostics.AddError(
			"Issue deleting Semantic Layer Configuration",
			"Error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *snowflakeSemanticLayerCredentialResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state SnowflakeSLCredentialModel

	// Read plan and state values into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueInt64()

	credential, err := r.client.GetSemanticLayerCredential(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Issue getting Semantic Layer configuration",
			"Error: "+err.Error(),
		)
		return
	}

	values := map[string]interface{}{
		"role":                   plan.Credential.Role.ValueString(),
		"warehouse":              plan.Credential.Warehouse.ValueString(),
		"user":                   plan.Credential.User.ValueString(),
		"password":               plan.Credential.Password.ValueString(),
		"private_key":            plan.Credential.PrivateKey.ValueString(),
		"private_key_passphrase": plan.Credential.PrivateKeyPassphrase.ValueString(),
		"auth_type":              plan.Credential.AuthType.ValueString(),
	}

	credential.Name = plan.Configuration.Name.ValueString()
	credential.Values = values

	_, err = r.client.UpdateSemanticLayerCredential(
		id,
		*credential,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update Semantic Layer credential",
			"Error: "+err.Error(),
		)
		return
	}

	state.ID = types.Int64Value(int64(*credential.ID))
	state.Credential.CredentialID = types.Int64Value(int64(*credential.ID))

	//update config fields
	state.Configuration.Name = types.StringValue(credential.Name)

	//update credential fields
	state.Credential.AuthType = types.StringValue(credential.Values["auth_type"].(string))
	state.Credential.Role = types.StringValue(credential.Values["role"].(string))
	state.Credential.Warehouse = types.StringValue(credential.Values["warehouse"].(string))
	state.Credential.Password = types.StringValue(credential.Values["password"].(string))
	state.Credential.User = types.StringValue(credential.Values["user"].(string))
	state.Credential.PrivateKey = types.StringValue(credential.Values["private_key"].(string))
	state.Credential.PrivateKeyPassphrase = types.StringValue(credential.Values["private_key_passphrase"].(string))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *snowflakeSemanticLayerCredentialResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}

func (r *snowflakeSemanticLayerCredentialResource) Schema(
	_ context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = snowflake_sl_credential_resource_schema
}
