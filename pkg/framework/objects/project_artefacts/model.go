package project_artefacts

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProjectArtefactsResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProjectID      types.Int64  `tfsdk:"project_id"`
	DocsJobID      types.Int64  `tfsdk:"docs_job_id"`
	FreshnessJobID types.Int64  `tfsdk:"freshness_job_id"`
}
