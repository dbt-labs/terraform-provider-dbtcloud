package environment

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (r *environmentDataSource) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve data for a single environment",
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the environment",
			},
			"project_id": schema.Int64Attribute{
				Required:    true,
				Description: "The project ID to which the environment belong",
			},
			"credentials_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The project ID to which the environment belong",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the environment",
			},
			"dbt_version": schema.StringAttribute{
				Computed:    true,
				Description: "Version number of dbt to use in this environment, usually in the format 1.2.0-latest rather than core versions",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the environment",
			},
			"use_custom_branch": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to use a custom git branch in this environment",
			},
			"custom_branch": schema.StringAttribute{
				Computed:    true,
				Description: "The type of deployment environment (currently 'production', 'staging' or empty)",
			},
			"deployment_type": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the environment",
			},
			"extended_attributes_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the extended attributes applied",
			},
		},
	}
}

func (r *environmentsDataSources) Schema(
	_ context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Retrieve data for multiple environments",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.Int64Attribute{
				Optional:    true,
				Description: "The project ID to filter the environments for [Optional]",
			},
			"environments": schema.SetNestedAttribute{
				Description: "The list of environments",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"environment_id": schema.Int64Attribute{
							Computed:    true,
							Description: "The ID of the environment",
						},
						"project_id": schema.Int64Attribute{
							Computed:    true,
							Description: "The project ID to which the environment belong",
						},
						"credentials_id": schema.Int64Attribute{
							Computed:    true,
							Description: "The project ID to which the environment belong",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the environment",
						},
						"dbt_version": schema.StringAttribute{
							Computed:    true,
							Description: "Version number of dbt to use in this environment, usually in the format 1.2.0-latest rather than core versions",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the environment",
						},
						"use_custom_branch": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether to use a custom git branch in this environment",
						},
						"custom_branch": schema.StringAttribute{
							Computed:    true,
							Description: "The type of deployment environment (currently 'production', 'staging' or empty)",
						},
						"deployment_type": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the environment",
						},
						"extended_attributes_id": schema.Int64Attribute{
							Computed:    true,
							Description: "The ID of the extended attributes applied",
						},
					},
				},
			},
		},
	}
}
