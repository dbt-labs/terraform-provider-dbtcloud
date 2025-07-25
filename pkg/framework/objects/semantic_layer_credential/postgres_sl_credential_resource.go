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
	_ resource.Resource              = &postgresSemanticLayerCredentialResource{}
	_ resource.ResourceWithConfigure = &postgresSemanticLayerCredentialResource{}
)

func PostgresSemanticLayerCredentialResource() resource.Resource {
	return &postgresSemanticLayerCredentialResource{}
}

// dbtCloud.Client for making API calls
type postgresSemanticLayerCredentialResource struct {
	client *dbt_cloud.Client
}

func (r *postgresSemanticLayerCredentialResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_postgres_semantic_layer_credential"
}

func (r *postgresSemanticLayerCredentialResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state PostgresSLCredentialModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	id := state.ID.ValueInt64()

	credential, err := r.client.GetSemanticLayerCredential(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Issue getting Semantic Layer credential",
			"Error: "+err.Error(),
		)

		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The Semantic Layer credential was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		return
	}

	state.ID = types.Int64Value(int64(*credential.ID))
	state.Credential.ProjectID = types.Int64Value(int64(credential.ProjectID))
	state.Credential.CredentialID = types.Int64Value(int64(*credential.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *postgresSemanticLayerCredentialResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan PostgresSLCredentialModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.Credential.ProjectID.ValueInt64()
	password := ""
	if plan.Credential.Password.ValueStringPointer() != nil {
		password = *plan.Credential.Password.ValueStringPointer()
	}

	values := map[string]interface{}{
		"username": plan.Credential.Username.ValueString(),
		"password": password,
	}

	createdCredential, err := r.client.CreateSemanticLayerCredential(
		projectID,
		values,
		plan.Configuration.Name.ValueString(),
		plan.Configuration.AdapterVersion.ValueString(),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Semantic Layer Credential",
			"Error: "+err.Error(),
		)
		return
	}
	plan.ID = types.Int64Value(int64(*createdCredential.ID))
	plan.Credential.CredentialID = types.Int64Value(int64(*createdCredential.ID))
	plan.Credential.ID = types.StringValue(fmt.Sprintf("%d", *createdCredential.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *postgresSemanticLayerCredentialResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state PostgresSLCredentialModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID
	projectID := state.Credential.ProjectID.ValueInt64()

	err := r.client.DeleteSemanticLayerCredential(projectID, id.ValueInt64())

	if err != nil {
		resp.Diagnostics.AddError(
			"Issue deleting Semantic Layer Credential",
			"Error: "+err.Error(),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *postgresSemanticLayerCredentialResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state PostgresSLCredentialModel

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
			"Issue getting Semantic Layer Credential",
			"Error: "+err.Error(),
		)
		return
	}

	//add credential fields to values map
	values := map[string]interface{}{
		"username": plan.Credential.Username.ValueString(),
		"password": plan.Credential.Password.ValueString(),
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
	state.Credential.Password = getStringFromMap(credential.Values, "password")
	state.Credential.Username = getStringFromMap(credential.Values, "username")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *postgresSemanticLayerCredentialResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}

func (r *postgresSemanticLayerCredentialResource) Schema(
	_ context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = postgres_sl_credential_resource_schema
}
