package project

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

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &projectResource{}
var _ resource.ResourceWithImportState = &projectResource{}

// Resource defines the resource implementation.
type projectResource struct {
	client *dbt_cloud.Client
}

// ProjectResource creates a new resource
func ProjectResource() resource.Resource {
	return &projectResource{}
}

// Metadata returns the resource type name.
func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the schema for the resource.
func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema
}

// Configure adds the provider configured client to the resource.
func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read model from plan
	var plan ProjectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get values from plan
	name := plan.Name.ValueString()
	description := ""
	if !plan.Description.IsNull() {
		description = plan.Description.ValueString()
	}
	dbtProjectSubdirectory := ""

	if !plan.DbtProjectSubdirectory.IsNull() {
		dbtProjectSubdirectory = plan.DbtProjectSubdirectory.ValueString()
	}

	// Call CreateProject with the correct signature (string, string, string)
	project, err := r.client.CreateProject(name, description, dbtProjectSubdirectory)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project",
			"Could not create project: "+err.Error(),
		)
		return
	}

	// Map response body to model
	plan.ID = types.Int64Value(int64(*project.ID))
	plan.Description = types.StringValue(project.Description)

	if project.DbtProjectSubdirectory != nil {
		plan.DbtProjectSubdirectory = types.StringValue(*project.DbtProjectSubdirectory)
	} else {
		plan.DbtProjectSubdirectory = types.StringNull()
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ProjectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed project value from dbt Cloud
	projectID := int(state.ID.ValueInt64())
	projectIDString := strconv.Itoa(projectID)

	project, err := r.client.GetProject(projectIDString)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading dbt Cloud Project",
			"Could not read project ID "+state.ID.String()+": "+err.Error(),
		)
		return
	}

	// Update state with refreshed values
	state.Name = types.StringValue(project.Name)
	state.Description = types.StringValue(project.Description)

	if project.DbtProjectSubdirectory != nil {
		state.DbtProjectSubdirectory = types.StringValue(*project.DbtProjectSubdirectory)
	} else {
		state.DbtProjectSubdirectory = types.StringNull()
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read model from plan
	var plan ProjectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state ProjectResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare project ID
	projectID := int(state.ID.ValueInt64())
	projectIDString := strconv.Itoa(projectID)

	updateProject, err := r.client.GetProject(projectIDString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving dbt Cloud Project",
			"Could not read project "+state.ID.String()+": "+err.Error(),
		)
		return
	}

	updateProject.Name = plan.Name.ValueString()
	// // Create a minimal update object with just the required fields
	// updateProject := dbt_cloud.Project{
	// 	ID:   &projectID,
	// 	Name: plan.Name.ValueString(),
	// }

	// Set optional fields if they're in the plan
	if !plan.Description.IsNull() {
		description := plan.Description.ValueString()
		updateProject.Description = description
	}

	if !plan.DbtProjectSubdirectory.IsNull() {
		dbtProjectSubdir := plan.DbtProjectSubdirectory.ValueString()
		updateProject.DbtProjectSubdirectory = &dbtProjectSubdir
	}

	// Perform the update
	project, err := r.client.UpdateProject(projectIDString, *updateProject)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating project",
			"Could not update project "+state.ID.String()+": "+err.Error(),
		)
		return
	}

	// Update state with updated values
	plan.ID = types.Int64Value(int64(*project.ID))
	plan.Description = types.StringValue(project.Description)

	if project.DbtProjectSubdirectory != nil {
		plan.DbtProjectSubdirectory = types.StringValue(*project.DbtProjectSubdirectory)
	} else {
		plan.DbtProjectSubdirectory = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state ProjectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete project
	projectID := int(state.ID.ValueInt64())
	projectIDString := strconv.Itoa(projectID)

	// Get the project first
	project, err := r.client.GetProject(projectIDString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving dbt Cloud Project for deletion",
			"Could not read project "+state.ID.String()+": "+err.Error(),
		)
		return
	}

	// Set the state to deleted (soft delete)
	project.State = dbt_cloud.STATE_DELETED

	// Update the project with the deleted state
	_, err = r.client.UpdateProject(projectIDString, *project)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting dbt Cloud Project",
			"Could not delete project "+state.ID.String()+": "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Convert the import ID (project ID) to an int64
	projectID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing project",
			"Could not convert project ID to integer: "+err.Error(),
		)
		return
	}

	// Set the ID attribute
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), projectID)...)
}
