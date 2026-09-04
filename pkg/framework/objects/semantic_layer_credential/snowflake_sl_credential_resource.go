package semantic_layer_credential

import (
	"context"
	"fmt"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
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
		if helper.HandleResourceNotFound(ctx, err, &resp.Diagnostics, &resp.State, "snowflake semantic layer credential") {
			return
		}
		resp.Diagnostics.AddError(
			"Issue getting Semantic Layer credential",
			"Error: "+err.Error(),
		)
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
	var plan, config SnowflakeSLCredentialModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.Credential.ProjectID.ValueInt64()
	password := helper.ResolveWriteOnlyString(config.Credential.PasswordWo, plan.Credential.Password)
	privateKey := helper.ResolveWriteOnlyString(config.Credential.PrivateKeyWo, plan.Credential.PrivateKey)
	privateKeyPassphrase := helper.ResolveWriteOnlyString(config.Credential.PrivateKeyPassphraseWo, plan.Credential.PrivateKeyPassphrase)

	values := map[string]interface{}{
		"role":                   plan.Credential.Role.ValueString(),
		"warehouse":              plan.Credential.Warehouse.ValueString(),
		"user":                   plan.Credential.User.ValueString(),
		"password":               password,
		"private_key":            privateKey,
		"private_key_passphrase": privateKeyPassphrase,
		"auth_type":              plan.Credential.AuthType.ValueString(),
	}

	createdCredential, err := r.client.CreateSemanticLayerCredential(
		projectID,
		values,
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
	var plan, state, config SnowflakeSLCredentialModel

	// Read plan and state values into the models
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
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

	password := helper.ResolveWriteOnlyString(config.Credential.PasswordWo, plan.Credential.Password)
	privateKey := helper.ResolveWriteOnlyString(config.Credential.PrivateKeyWo, plan.Credential.PrivateKey)
	privateKeyPassphrase := helper.ResolveWriteOnlyString(config.Credential.PrivateKeyPassphraseWo, plan.Credential.PrivateKeyPassphrase)

	values := map[string]interface{}{
		"role":                   plan.Credential.Role.ValueString(),
		"warehouse":              plan.Credential.Warehouse.ValueString(),
		"user":                   plan.Credential.User.ValueString(),
		"password":               password,
		"private_key":            privateKey,
		"private_key_passphrase": privateKeyPassphrase,
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
	state.Configuration = plan.Configuration
	state.Credential = plan.Credential
	state.Credential.CredentialID = types.Int64Value(int64(*credential.ID))
	state.Credential.ID = types.StringValue(fmt.Sprintf("%d", *credential.ID))

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
