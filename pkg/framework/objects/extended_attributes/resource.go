package extended_attributes

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &extendedAttributesResource{}
	_ resource.ResourceWithConfigure   = &extendedAttributesResource{}
	_ resource.ResourceWithImportState = &extendedAttributesResource{}
)

// ExtendedAttributesResource is a helper function to simplify the provider implementation.
func ExtendedAttributesResource() resource.Resource {
	return &extendedAttributesResource{}
}

// extendedAttributesResource is the resource implementation.
type extendedAttributesResource struct {
	client *dbt_cloud.Client
}

// Configure adds the provider configured client to the resource.
func (r *extendedAttributesResource) Configure(
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
func (r *extendedAttributesResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_extended_attributes"
}

// Schema defines the schema for the resource.
func (r *extendedAttributesResource) Schema(
	ctx context.Context,
	req resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resourceSchema
}

func (r *extendedAttributesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The ID format is "project_id:extended_attributes_id"
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format 'project_id:extended_attributes_id'",
		)
		return
	}

	projectID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid project ID",
			"Project ID must be a valid integer",
		)
		return
	}

	extendedAttributesID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid extended attributes ID",
			"Extended attributes ID must be a valid integer",
		)
		return
	}

	// Set the state values
	resp.State.SetAttribute(ctx, path.Root("id"), fmt.Sprintf("%d%s%d", projectID, dbt_cloud.ID_DELIMITER, extendedAttributesID))
	resp.State.SetAttribute(ctx, path.Root("project_id"), projectID)
	resp.State.SetAttribute(ctx, path.Root("extended_attributes_id"), extendedAttributesID)
}

func (r *extendedAttributesResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	// Retrieve values from plan
	var plan ExtendedAttributesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	state := int(plan.State.ValueInt64())
	extendedAttributesRaw := json.RawMessage([]byte(plan.ExtendedAttributes.ValueString()))

	// Create new extended attributes
	extendedAttributes, err := r.client.CreateExtendedAttributes(state, projectID, extendedAttributesRaw)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating extended attributes",
			"Could not create extended attributes, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate computed values
	plan.ID = types.StringValue(fmt.Sprintf(
		"%d%s%d",
		extendedAttributes.ProjectID,
		dbt_cloud.ID_DELIMITER,
		*extendedAttributes.ID,
	))

	plan.ExtendedAttributesID = types.Int64Value(int64(*extendedAttributes.ID))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *extendedAttributesResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state
	var state ExtendedAttributesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get extended attributes from API
	projectID, extendedAttributesID, err := helper.SplitIDToInts(
		state.ID.ValueString(),
		"dbtcloud_extended_attributes",
	)
	if err != nil {
		return
	}

	extendedAttributes, err := r.client.GetExtendedAttributes(projectID, extendedAttributesID)
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading extended attributes",
			"Could not read extended attributes ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Refresh state values
	state.ExtendedAttributes = types.StringValue(string(extendedAttributes.ExtendedAttributes))
	state.State = types.Int64Value(int64(extendedAttributes.State))
	state.ProjectID = types.Int64Value(int64(extendedAttributes.ProjectID))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *extendedAttributesResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Retrieve values from plan
	var plan ExtendedAttributesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state ExtendedAttributesResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, extendedAttributesID, err := helper.SplitIDToInts(
		state.ID.ValueString(),
		"dbtcloud_extended_attributes",
	)
	if err != nil {
		return
	}

	if (plan.State != state.State) ||
		(plan.ProjectID != state.ProjectID) ||
		(plan.ExtendedAttributes != state.ExtendedAttributes) {

		extendedAttributes, err := r.client.GetExtendedAttributes(
			projectID,
			extendedAttributesID,
		)

		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating extended attributes",
				"Could not update extended attributes ID "+state.ID.ValueString()+": "+err.Error(),
			)
			return
		}

		extendedAttributes.State = int(plan.State.ValueInt64())
		extendedAttributes.ProjectID = int(plan.ProjectID.ValueInt64())
		attributes := helper.NormalizeJSONString(plan.ExtendedAttributes.ValueString())
		extendedAttributes.ExtendedAttributes = json.RawMessage([]byte(attributes))

		_, err = r.client.UpdateExtendedAttributes(
			projectID,
			extendedAttributesID,
			*extendedAttributes,
		)

		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating extended attributes",
				"Could not update extended attributes ID "+state.ID.ValueString()+": "+err.Error(),
			)
			return
		}

		plan.ExtendedAttributes = types.StringValue(helper.NormalizeJSONString(plan.ExtendedAttributes.ValueString()))
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *extendedAttributesResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Retrieve values from state
	var state ExtendedAttributesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, extendedAttributesID, err := helper.SplitIDToInts(
		state.ID.ValueString(),
		"dbtcloud_extended_attributes",
	)
	if err != nil {
		return
	}

	_, err = r.client.DeleteExtendedAttributes(
		projectID,
		extendedAttributesID,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting extended attributes",
			"Could not delete extended attributes, unexpected error: "+err.Error(),
		)
		return
	}
}
