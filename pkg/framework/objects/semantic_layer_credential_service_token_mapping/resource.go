package semantic_layer_credential_service_token_mapping

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &semanticLayerCredentialServiceTokenMappingResource{}
	_ resource.ResourceWithConfigure   = &semanticLayerCredentialServiceTokenMappingResource{}
	_ resource.ResourceWithImportState = &semanticLayerCredentialServiceTokenMappingResource{}
)

func SemanticLayerCredentialServiceTokenMappingResource() resource.Resource {
	return &semanticLayerCredentialServiceTokenMappingResource{}
}

type semanticLayerCredentialServiceTokenMappingResource struct {
	client *dbt_cloud.Client
}

func (r *semanticLayerCredentialServiceTokenMappingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_semantic_layer_credential_service_token_mapping"
}

func (r *semanticLayerCredentialServiceTokenMappingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *semanticLayerCredentialServiceTokenMappingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier of the semantic layer credential service token mapping.",
			},
			"semantic_layer_credential_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the semantic layer credential to map.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"service_token_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the service token to map to the semantic layer credential.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the project to which the semantic layer credential is associated.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *semanticLayerCredentialServiceTokenMappingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read model from plan
	var plan SemanticLayerCredentialServiceTokenMapping
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project_id := plan.ProjectID.ValueInt64()
	cred_id := plan.SemanticLayerCredentialID.ValueInt64()
	token_id := plan.ServiceTokenID.ValueInt64()

	// Create the semantic layer credential service token mapping
	mapping, err := r.client.CreateSemanticLayerCredentialServiceTokenMapping(
		int(project_id),
		int(cred_id),
		int(token_id),
	)

	if err != nil {
		// Check if it's a duplicate key error - if so, try to read the existing resource
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "already exists") {
			resp.Diagnostics.AddWarning(
				"Resource already exists",
				"The Semantic Layer Credential Service Token Mapping already exists in dbt Cloud. Attempting to import the existing resource into Terraform state.",
			)

			// Try to read the existing mapping
			existingSearch := dbt_cloud.SemanticLayerCredentialServiceTokenMapping{
				SemanticLayerCredentialID: int(cred_id),
				ServiceTokenID:            int(token_id),
				ProjectID:                 int(project_id),
			}

			existingMapping, readErr := r.client.GetSemanticLayerCredentialServiceTokenMapping(existingSearch)
			if readErr != nil {
				resp.Diagnostics.AddError(
					"Error Creating Semantic Layer Credential Service Token Mapping",
					fmt.Sprintf("The resource already exists but could not be read to import into state.\n\nOriginal create error: %s\n\nRead error: %s", err.Error(), readErr.Error()),
				)
				return
			}

			// Successfully found the existing resource - use it
			mapping = existingMapping
		} else {
			// Different error - return it
			resp.Diagnostics.AddError(
				"Error Creating Semantic Layer Credential Service Token Mapping",
				err.Error(),
			)
			return
		}
	}

	// Set the ID in the state
	plan.ID = types.Int64Value(int64(*mapping.ID))
	plan.SemanticLayerCredentialID = types.Int64Value(int64(mapping.SemanticLayerCredentialID))
	plan.ServiceTokenID = types.Int64Value(int64(mapping.ServiceTokenID))
	plan.ProjectID = types.Int64Value(int64(mapping.ProjectID))

	// Set the state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *semanticLayerCredentialServiceTokenMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state SemanticLayerCredentialServiceTokenMapping
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := int(state.ID.ValueInt64())
	cred_id := int(state.SemanticLayerCredentialID.ValueInt64())
	token_id := int(state.ServiceTokenID.ValueInt64())
	project_id := int(state.ProjectID.ValueInt64())

	sm := dbt_cloud.SemanticLayerCredentialServiceTokenMapping{
		ID:                        &id,
		SemanticLayerCredentialID: cred_id,
		ServiceTokenID:            token_id,
		ProjectID:                 project_id,
	}

	// Read the semantic layer credential service token mapping
	mapping, err := r.client.GetSemanticLayerCredentialServiceTokenMapping(sm)
	if err != nil {
		// Check if the error is a "resource-not-found" error
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			// Resource no longer exists, remove it from state
			resp.Diagnostics.AddWarning(
				"Resource not found",
				"The Semantic Layer Credential Service Token Mapping was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Semantic Layer Credential Service Token Mapping",
			err.Error(),
		)
		return
	}

	// Update the state with the latest data
	state.ID = types.Int64Value(int64(*mapping.ID))
	state.SemanticLayerCredentialID = types.Int64Value(int64(mapping.SemanticLayerCredentialID))
	state.ServiceTokenID = types.Int64Value(int64(mapping.ServiceTokenID))
	state.ProjectID = types.Int64Value(int64(mapping.ProjectID))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// update not implemented, the resource is always replaced on changes
func (r *semanticLayerCredentialServiceTokenMappingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *semanticLayerCredentialServiceTokenMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state SemanticLayerCredentialServiceTokenMapping
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sm_id := int(state.ID.ValueInt64())

	// Delete the semantic layer credential service token mapping
	err := r.client.DeleteSemanticLayerCredentialServiceTokenMapping(sm_id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Semantic Layer Credential Service Token Mapping",
			err.Error(),
		)
		return
	}

	// Remove the resource from state
	resp.State.RemoveResource(ctx)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *semanticLayerCredentialServiceTokenMappingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Importing a semantic layer credential service token mapping is not supported.
	resp.Diagnostics.AddError(
		"Import Not Supported",
		"Importing a semantic layer credential service token mapping is not supported. Please create the resource using the provider.",
	)
}
