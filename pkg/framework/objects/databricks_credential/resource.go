package databricks_credential

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
	_ resource.Resource                = &databricksCredentialResource{}
	_ resource.ResourceWithConfigure   = &databricksCredentialResource{}
	_ resource.ResourceWithImportState = &databricksCredentialResource{}
)

func DatabricksCredentialResource() resource.Resource {
	return &databricksCredentialResource{}
}

type databricksCredentialResource struct {
	client *dbt_cloud.Client
}

func (d *databricksCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

	credentialResponse, err := d.client.GetDatabricksCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError("Error getting databricks credential", err.Error())
		return
	}

	// Set ID
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("id"),
		fmt.Sprintf("%d:%d", projectID, credentialID),
	)...)

	// Set project_id
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("project_id"),
		projectID,
	)...)

	// Set credential_id
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("credential_id"),
		credentialID,
	)...)

	// Set target_name
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("target_name"),
		credentialResponse.Target_Name,
	)...)

	// Set schema from API response
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("schema"),
		credentialResponse.UnencryptedCredentialDetails.Schema,
	)...)

	// Set catalog
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("catalog"),
		credentialResponse.UnencryptedCredentialDetails.Catalog,
	)...)

	// Set adapter_type - this is required but not returned from the API
	// Since it's in the ImportStateVerifyIgnore list, we can skip it

	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *databricksCredentialResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	d.client = client
}

func (d *databricksCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DatabricksCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	d.createGlobal(ctx, &plan, resp)
}

func (d *databricksCredentialResource) createGlobal(ctx context.Context, plan *DatabricksCredentialResourceModel, resp *resource.CreateResponse) {
	projectID := int(plan.ProjectID.ValueInt64())
	token := plan.Token.ValueString()
	schema := plan.Schema.ValueString()
	targetName := plan.TargetName.ValueString()
	catalog := plan.Catalog.ValueString()
	adapterType := plan.AdapterType.ValueString()

	// For now, just supporting databricks
	if adapterType == "spark" {
		resp.Diagnostics.AddError(
			"Spark adapter not supported",
			"Spark adapter is not supported currently for global connections credentials. Please raise a GitHub issue if you need it",
		)
		return
	}

	databricksCredential, err := d.client.CreateDatabricksCredential(
		projectID,
		token,
		schema,
		targetName,
		catalog,
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Databricks credential", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d%s%d", databricksCredential.Project_Id, dbt_cloud.ID_DELIMITER, *databricksCredential.ID))
	plan.CredentialID = types.Int64Value(int64(*databricksCredential.ID))

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *databricksCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DatabricksCredentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	d.deleteGlobal(ctx, &state, resp)
}

func (d *databricksCredentialResource) deleteGlobal(_ context.Context, state *DatabricksCredentialResourceModel, resp *resource.DeleteResponse) {
	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	_, err := d.client.DeleteCredential(
		strconv.Itoa(credentialID),
		strconv.Itoa(projectID),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting Databricks credential", err.Error())
		return
	}
}

func (d *databricksCredentialResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databricks_credential"
}

func (d *databricksCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DatabricksCredentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	credentialID := int(state.CredentialID.ValueInt64())

	credential, err := d.client.GetDatabricksCredential(projectID, credentialID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Databricks credential", "Could not read Databricks credential ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.Catalog = types.StringValue(credential.UnencryptedCredentialDetails.Catalog)
	state.Schema = types.StringValue(credential.UnencryptedCredentialDetails.Schema)
	state.TargetName = types.StringValue(credential.UnencryptedCredentialDetails.TargetName)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *databricksCredentialResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DatabricksResourceSchema
}

func (d *databricksCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state DatabricksCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	d.updateGlobal(ctx, &plan, &state, resp)
}

func (d *databricksCredentialResource) updateGlobal(ctx context.Context, plan, state *DatabricksCredentialResourceModel, resp *resource.UpdateResponse) {
	projectID, credentialID, err := helper.SplitIDToInts(
		state.ID.ValueString(),
		"databricks_credential",
	)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID format", err.Error())
		return
	}

	// Check if any relevant fields have changed
	if !plan.Token.Equal(state.Token) ||
		!plan.TargetName.Equal(state.TargetName) ||
		!plan.Catalog.Equal(state.Catalog) ||
		!plan.Schema.Equal(state.Schema) {

		patchCredentialsDetails, err := dbt_cloud.GenerateDatabricksCredentialDetails(
			plan.Token.ValueString(),
			plan.Schema.ValueString(),
			plan.TargetName.ValueString(),
			plan.Catalog.ValueString(),
		)
		if err != nil {
			resp.Diagnostics.AddError("Error generating credential details", err.Error())
			return
		}

		// Remove unchanged fields
		for key := range patchCredentialsDetails.Fields {
			switch key {
			case "token":
				if plan.Token.Equal(state.Token) {
					delete(patchCredentialsDetails.Fields, key)
				}
			case "schema":
				if plan.Schema.Equal(state.Schema) {
					delete(patchCredentialsDetails.Fields, key)
				}
			case "target_name":
				if plan.TargetName.Equal(state.TargetName) {
					delete(patchCredentialsDetails.Fields, key)
				}
			case "catalog":
				if plan.Catalog.Equal(state.Catalog) {
					delete(patchCredentialsDetails.Fields, key)
				}
			}
		}

		databricksPatch := dbt_cloud.DatabricksCredentialGLobConnPatch{
			ID:                credentialID,
			CredentialDetails: patchCredentialsDetails,
		}

		_, err = d.client.UpdateDatabricksCredentialGlobConn(projectID, credentialID, databricksPatch)
		if err != nil {
			resp.Diagnostics.AddError("Error updating Databricks credential", err.Error())
			return
		}
	}

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
