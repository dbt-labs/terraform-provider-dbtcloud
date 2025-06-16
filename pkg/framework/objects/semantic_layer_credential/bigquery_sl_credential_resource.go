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
	_ resource.Resource              = &bigQuerySemanticLayerCredentialResource{}
	_ resource.ResourceWithConfigure = &bigQuerySemanticLayerCredentialResource{}
)

func BigQuerySemanticLayerCredentialResource() resource.Resource {
	return &bigQuerySemanticLayerCredentialResource{}
}

type bigQuerySemanticLayerCredentialResource struct {
	client *dbt_cloud.Client
}

func (r *bigQuerySemanticLayerCredentialResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_bigquery_semantic_layer_credential"
}

func (r *bigQuerySemanticLayerCredentialResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state BigQuerySLCredentialModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	id := state.ID.ValueInt64()

	credential, err := r.client.GetSemanticLayerCredential(id)

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

		resp.Diagnostics.AddError(
			"Issue getting Semantic Layer credential",
			"Error: "+err.Error(),
		)
		return
	}

	state.ID = types.Int64Value(int64(*credential.ID))
	state.Credential.ProjectID = types.Int64Value(int64(credential.ProjectID))
	state.Credential.CredentialID = types.Int64Value(int64(*credential.ID))

	state.Configuration.ProjectID = types.Int64Value(int64(credential.ProjectID))
	state.Configuration.Name = types.StringValue(credential.Name)
	state.Configuration.AdapterVersion = types.StringValue(credential.AdapterVersion)

	state.AuthURI = getStringFromMap(credential.Values, "auth_uri")
	state.TokenURI = getStringFromMap(credential.Values, "token_uri")

	state.ClientEmail = getStringFromMap(credential.Values, "client_email")
	state.ClientID = getStringFromMap(credential.Values, "client_id")
	state.AuthProviderCertURL = getStringFromMap(credential.Values, "auth_provider_x509_cert_url")
	state.ClientCertURL = getStringFromMap(credential.Values, "client_x509_cert_url")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func getStringFromMap(m map[string]interface{}, key string) types.String {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return types.StringValue(str)
		}
	}
	return types.StringNull()
}

func (r *bigQuerySemanticLayerCredentialResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan BigQuerySLCredentialModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.Credential.ProjectID.ValueInt64()

	//add credential fields to values map
	values := map[string]interface{}{
		"private_key_id":              plan.PrivateKeyID.ValueString(),
		"private_key":                 plan.PrivateKey.ValueString(),
		"client_email":                plan.ClientEmail.ValueString(),
		"client_id":                   plan.ClientID.ValueString(),
		"auth_uri":                    plan.AuthURI.ValueString(),
		"token_uri":                   plan.TokenURI.ValueString(),
		"auth_provider_x509_cert_url": plan.AuthProviderCertURL.ValueString(),
		"client_x509_cert_url":        plan.ClientCertURL.ValueString(),
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

	plan.Credential.CredentialID = types.Int64Value(int64(*createdCredential.ID))
	plan.Credential.ID = types.StringValue(fmt.Sprintf("%d", *createdCredential.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bigQuerySemanticLayerCredentialResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state BigQuerySLCredentialModel

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

func (r *bigQuerySemanticLayerCredentialResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan, state BigQuerySLCredentialModel

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
		"private_key_id":              plan.PrivateKeyID.ValueString(),
		"private_key":                 plan.PrivateKey.ValueString(),
		"client_email":                plan.ClientEmail.ValueString(),
		"client_id":                   plan.ClientID.ValueString(),
		"auth_uri":                    plan.AuthURI.ValueString(),
		"token_uri":                   plan.TokenURI.ValueString(),
		"auth_provider_x509_cert_url": plan.AuthProviderCertURL.ValueString(),
		"client_x509_cert_url":        plan.ClientCertURL.ValueString(),
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

	updatedCredential, err := r.client.GetSemanticLayerCredential(id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Issue getting Semantic Layer configuration",
			"Error: "+err.Error(),
		)
		return
	}

	state.ID = types.Int64Value(int64(*updatedCredential.ID))
	state.Credential.ProjectID = types.Int64Value(int64(updatedCredential.ProjectID))
	state.Credential.CredentialID = types.Int64Value(int64(*updatedCredential.ID))

	state.Configuration.ProjectID = types.Int64Value(int64(updatedCredential.ProjectID))
	state.Configuration.Name = types.StringValue(updatedCredential.Name)
	state.Configuration.AdapterVersion = types.StringValue(updatedCredential.AdapterVersion)

	state.AuthURI = getStringFromMap(updatedCredential.Values, "auth_uri")
	state.TokenURI = getStringFromMap(updatedCredential.Values, "token_uri")

	state.ClientEmail = getStringFromMap(updatedCredential.Values, "client_email")
	state.ClientID = getStringFromMap(updatedCredential.Values, "client_id")
	state.AuthProviderCertURL = getStringFromMap(updatedCredential.Values, "auth_provider_x509_cert_url")
	state.ClientCertURL = getStringFromMap(updatedCredential.Values, "client_x509_cert_url")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bigQuerySemanticLayerCredentialResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*dbt_cloud.Client)
}

func (r *bigQuerySemanticLayerCredentialResource) Schema(
	_ context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = bigquery_sl_credential_resource_schema
}
