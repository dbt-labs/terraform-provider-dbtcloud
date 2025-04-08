package project_repository

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
var _ resource.Resource = &projectRepositoryResource{}
var _ resource.ResourceWithImportState = &projectRepositoryResource{}

// Resource defines the resource implementation.
type projectRepositoryResource struct {
	client *dbt_cloud.Client
}

// NewResource creates a new resource
func ProjectRepositoryResource() resource.Resource {
	return &projectRepositoryResource{}
}

// Metadata returns the resource type name.
func (r *projectRepositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_repository"
}

// Schema defines the schema for the resource.
func (r *projectRepositoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = Schema()
}

// Configure adds the provider configured client to the resource.
func (r *projectRepositoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *projectRepositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read model from plan
	var plan Model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(plan.ProjectID.ValueInt64())
	repositoryID := int(plan.RepositoryID.ValueInt64())
	projectIDString := strconv.Itoa(projectID)

	project, err := r.client.GetProject(projectIDString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving project",
			"Could not get project: "+err.Error(),
		)
		return
	}

	// Issue #362
	// we don't want to update the connection ID when we set a project otherwise it will update all envs
	project.ConnectionID = nil

	project.RepositoryID = &repositoryID

	_, err = r.client.UpdateProject(projectIDString, *project)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project repository",
			"Could not create project repository: "+err.Error(),
		)
		return
	}

	// Set resource ID
	plan.ID = plan.ProjectID

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *projectRepositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	projectIDString := strconv.Itoa(projectID)

	project, err := r.client.GetProject(projectIDString)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading DBT Cloud Project",
			"Could not read project: "+err.Error(),
		)
		return
	}

	// Update state with refreshed values
	state.ProjectID = state.ID
	if project.RepositoryID != nil {
		state.RepositoryID = types.Int64Value(int64(*project.RepositoryID))
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *projectRepositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Project repository is ForceNew for all attributes, so this should never be called
	resp.Diagnostics.AddError(
		"Error updating project",
		"Project repository doesn't support updates, all changes require replacement",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *projectRepositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state Model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := int(state.ProjectID.ValueInt64())
	projectIDString := strconv.Itoa(projectID)

	project, err := r.client.GetProject(projectIDString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving project",
			"Could not get project: "+err.Error(),
		)
		return
	}

	// Issue #362
	// we don't want to update the connection ID when we set a project otherwise it will update all envs
	project.ConnectionID = nil

	project.RepositoryID = nil

	_, err = r.client.UpdateProject(projectIDString, *project)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting project repository",
			"Could not delete project repository: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *projectRepositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Split the ID to get the project_id and repository_id
	idParts := strings.Split(req.ID, dbt_cloud.ID_DELIMITER)

	if len(idParts) != 2 {
		resp.Diagnostics.AddError(
			"Error importing project repository",
			fmt.Sprintf("Expected import identifier with format: project_id%srepository_id", dbt_cloud.ID_DELIMITER),
		)
		return
	}

	projectID, err := strconv.Atoi(idParts[0])
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing project repository",
			"Could not convert project ID to integer: "+err.Error(),
		)
		return
	}

	repositoryID, err := strconv.Atoi(idParts[1])
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing project repository",
			"Could not convert repository ID to integer: "+err.Error(),
		)
		return
	}

	// Set the ID and required fields
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), int64(projectID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), int64(projectID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("repository_id"), int64(repositoryID))...)
}
