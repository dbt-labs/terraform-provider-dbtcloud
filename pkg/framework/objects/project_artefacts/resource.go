package project_artefacts

import (
	"context"
	"strconv"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &projectArtefactsResource{}
	_ resource.ResourceWithConfigure   = &projectArtefactsResource{}
	_ resource.ResourceWithImportState = &projectArtefactsResource{}
)

type projectArtefactsResource struct {
	client *dbt_cloud.Client
}

// ImportState implements resource.ResourceWithImportState.
func (p *projectArtefactsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id_as_int, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "The ID must be an integer")
		return
	}
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
	resp.State.SetAttribute(ctx, path.Root("project_id"), id_as_int)

}

// Configure implements resource.ResourceWithConfigure.
func (p *projectArtefactsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	switch c := req.ProviderData.(type) {
	case nil: // do nothing
	case *dbt_cloud.Client:
		p.client = c
	default:
		resp.Diagnostics.AddError("Missing client", "A client is required to configure the project artefacts resource")
	}
}

// Create implements resource.Resource.
func (p *projectArtefactsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectArtefactsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectIDString := strconv.FormatInt(plan.ProjectID.ValueInt64(), 10)

	project, err := p.client.GetProject(projectIDString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get project",
			"Error: "+err.Error(),
		)

		return
	}

	if plan.DocsJobID.ValueInt64() != 0 {
		conv := int(plan.DocsJobID.ValueInt64())
		project.DocsJobId = &conv
	} else {
		project.DocsJobId = nil
	}

	if plan.FreshnessJobID.ValueInt64() != 0 {
		conv := int(plan.FreshnessJobID.ValueInt64())
		project.FreshnessJobId = &conv
	} else {
		project.FreshnessJobId = nil
	}

	if _, err := p.client.UpdateProject(projectIDString, *project); err != nil {
		resp.Diagnostics.AddError(
			"Unable to update project",
			"Error: "+err.Error(),
		)

		return
	}

	plan.ID = types.StringValue(strconv.Itoa(*project.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete implements resource.Resource.
func (p *projectArtefactsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectArtefactsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectIDString := strconv.FormatInt(state.ProjectID.ValueInt64(), 10)

	project, err := p.client.GetProject(projectIDString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get project",
			"Error: "+err.Error(),
		)

		return
	}

	project.FreshnessJobId = nil
	project.DocsJobId = nil

	_, err = p.client.UpdateProject(projectIDString, *project)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update project",
			"Error: "+err.Error(),
		)

		return
	}
}

// Metadata implements resource.Resource.
func (p *projectArtefactsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_artefacts"
}

// Read implements resource.Resource.
func (p *projectArtefactsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectArtefactsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectIDString := strconv.FormatInt(state.ProjectID.ValueInt64(), 10)

	project, err := p.client.GetProject(projectIDString)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			resp.Diagnostics.AddError(
				"Project not found",
				"The project artefacts resource was not found and has been removed from the state.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to get project",
			"Error: "+err.Error(),
		)

		return
	}

	state.ID = types.StringValue(strconv.Itoa(*project.ID))
	if project.DocsJobId != nil {
		state.DocsJobID = types.Int64PointerValue(helper.IntPointerToInt64Pointer(project.DocsJobId))
	}

	if project.FreshnessJobId != nil {
		state.FreshnessJobID = types.Int64PointerValue(helper.IntPointerToInt64Pointer(project.FreshnessJobId))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)

}

// Update implements resource.Resource.
func (p *projectArtefactsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ProjectArtefactsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectIDString := strconv.FormatInt(plan.ProjectID.ValueInt64(), 10)

	project, err := p.client.GetProject(projectIDString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get project",
			"Error: "+err.Error(),
		)

		return
	}

	if !state.DocsJobID.Equal(plan.DocsJobID) {
		if plan.DocsJobID.ValueInt64() != 0 {
			conv := int(plan.DocsJobID.ValueInt64())
			project.DocsJobId = &conv
		} else {
			project.DocsJobId = nil
		}
	}

	if !state.FreshnessJobID.Equal(plan.FreshnessJobID) {
		if plan.FreshnessJobID.ValueInt64() != 0 {
			conv := int(plan.FreshnessJobID.ValueInt64())
			project.FreshnessJobId = &conv
		} else {
			project.FreshnessJobId = nil
		}
	}

	project, err = p.client.UpdateProject(projectIDString, *project)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update project",
			"Error: "+err.Error(),
		)

		return
	}

	plan.DocsJobID = types.Int64PointerValue(helper.IntPointerToInt64Pointer(project.DocsJobId))
	plan.FreshnessJobID = types.Int64PointerValue(helper.IntPointerToInt64Pointer(project.FreshnessJobId))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func ProjectArtefactsResource() resource.Resource {
	return &projectArtefactsResource{}
}
