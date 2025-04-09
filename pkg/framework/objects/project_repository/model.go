package project_repository

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model represents the resource model for a project repository
type Model struct {
	ID           types.String `tfsdk:"id"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	RepositoryID types.Int64  `tfsdk:"repository_id"`
}
