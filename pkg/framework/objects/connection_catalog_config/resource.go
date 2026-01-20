package connection_catalog_config

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
	_ resource.Resource                = &connectionCatalogConfigResource{}
	_ resource.ResourceWithConfigure   = &connectionCatalogConfigResource{}
	_ resource.ResourceWithImportState = &connectionCatalogConfigResource{}
)

// ConnectionCatalogConfigResource returns a new resource instance
func ConnectionCatalogConfigResource() resource.Resource {
	return &connectionCatalogConfigResource{}
}

type connectionCatalogConfigResource struct {
	client *dbt_cloud.Client
}

func (r *connectionCatalogConfigResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_connection_catalog_config"
}

func (r *connectionCatalogConfigResource) Configure(
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

func (r *connectionCatalogConfigResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ConnectionCatalogConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionID := plan.ConnectionID.ValueInt64()

	// Build the API request
	config := r.buildConfigFromModel(ctx, plan)

	// Create uses PATCH since there's no POST endpoint
	_, err := r.client.UpdateConnectionCatalogConfig(connectionID, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating connection catalog config",
			"Could not create connection catalog config: "+err.Error(),
		)
		return
	}

	// Set the ID
	plan.ID = types.StringValue(strconv.FormatInt(connectionID, 10))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *connectionCatalogConfigResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state ConnectionCatalogConfigResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionID := state.ConnectionID.ValueInt64()

	config, err := r.client.GetConnectionCatalogConfig(connectionID)
	if err != nil {
		if strings.Contains(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading connection catalog config",
			"Could not read connection catalog config: "+err.Error(),
		)
		return
	}

	// Update state with values from API
	state.DatabaseAllow = r.stringSliceToList(ctx, config.DatabaseAllow)
	state.DatabaseDeny = r.stringSliceToList(ctx, config.DatabaseDeny)
	state.SchemaAllow = r.stringSliceToList(ctx, config.SchemaAllow)
	state.SchemaDeny = r.stringSliceToList(ctx, config.SchemaDeny)
	state.TableAllow = r.stringSliceToList(ctx, config.TableAllow)
	state.TableDeny = r.stringSliceToList(ctx, config.TableDeny)
	state.ViewAllow = r.stringSliceToList(ctx, config.ViewAllow)
	state.ViewDeny = r.stringSliceToList(ctx, config.ViewDeny)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *connectionCatalogConfigResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ConnectionCatalogConfigResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionID := plan.ConnectionID.ValueInt64()

	// Build the API request
	config := r.buildConfigFromModel(ctx, plan)

	// Update the config
	_, err := r.client.UpdateConnectionCatalogConfig(connectionID, config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating connection catalog config",
			"Could not update connection catalog config: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *connectionCatalogConfigResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state ConnectionCatalogConfigResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectionID := state.ConnectionID.ValueInt64()

	// Delete by setting all fields to null
	err := r.client.DeleteConnectionCatalogConfig(connectionID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting connection catalog config",
			"Could not delete connection catalog config: "+err.Error(),
		)
		return
	}
}

func (r *connectionCatalogConfigResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	// Support both "connection_id" and "account_id:connection_id" formats
	idParts := strings.Split(req.ID, ":")

	var connectionID int64
	var err error

	if len(idParts) == 1 {
		// Just connection_id
		connectionID, err = strconv.ParseInt(idParts[0], 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				fmt.Sprintf("Could not parse connection_id as integer: %s", idParts[0]),
			)
			return
		}
	} else if len(idParts) == 2 {
		// account_id:connection_id format
		connectionID, err = strconv.ParseInt(idParts[1], 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Import ID",
				fmt.Sprintf("Could not parse connection_id as integer: %s", idParts[1]),
			)
			return
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: connection_id or account_id:connection_id. Got: %s", req.ID),
		)
		return
	}

	// Set the ID and connection_id for the Read operation
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("id"),
		strconv.FormatInt(connectionID, 10),
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx,
		path.Root("connection_id"),
		connectionID,
	)...)
}

// buildConfigFromModel converts the Terraform model to the API config struct
func (r *connectionCatalogConfigResource) buildConfigFromModel(
	ctx context.Context,
	model ConnectionCatalogConfigResourceModel,
) dbt_cloud.ConnectionCatalogConfig {
	config := dbt_cloud.ConnectionCatalogConfig{}

	config.DatabaseAllow = r.listToStringSlice(ctx, model.DatabaseAllow)
	config.DatabaseDeny = r.listToStringSlice(ctx, model.DatabaseDeny)
	config.SchemaAllow = r.listToStringSlice(ctx, model.SchemaAllow)
	config.SchemaDeny = r.listToStringSlice(ctx, model.SchemaDeny)
	config.TableAllow = r.listToStringSlice(ctx, model.TableAllow)
	config.TableDeny = r.listToStringSlice(ctx, model.TableDeny)
	config.ViewAllow = r.listToStringSlice(ctx, model.ViewAllow)
	config.ViewDeny = r.listToStringSlice(ctx, model.ViewDeny)

	return config
}

// listToStringSlice converts a types.List to a []string
func (r *connectionCatalogConfigResource) listToStringSlice(ctx context.Context, list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	elements := make([]types.String, 0, len(list.Elements()))
	list.ElementsAs(ctx, &elements, false)

	result := make([]string, len(elements))
	for i, elem := range elements {
		result[i] = elem.ValueString()
	}
	return result
}

// stringSliceToList converts a []string to a types.List
func (r *connectionCatalogConfigResource) stringSliceToList(ctx context.Context, slice []string) types.List {
	if slice == nil {
		return types.ListNull(types.StringType)
	}

	elements := make([]types.String, len(slice))
	for i, s := range slice {
		elements[i] = types.StringValue(s)
	}

	list, _ := types.ListValueFrom(ctx, types.StringType, elements)
	return list
}
