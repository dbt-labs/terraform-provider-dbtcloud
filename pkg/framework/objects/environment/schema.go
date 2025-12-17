package environment

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/environment/validators"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
				Description: "The project ID to which the environment belongs",
			},
			"credentials_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Credential ID for this environment. A credential is not required for development environments, as dbt Cloud defaults to the user's credentials, but deployment environments will have this.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the environment",
			},
			"dbt_version": schema.StringAttribute{
				Computed:    true,
				Description: "Version number of dbt to use in this environment.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of environment (must be either development or deployment)",
			},
			"use_custom_branch": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to use a custom git branch in this environment",
			},
			"custom_branch": schema.StringAttribute{
				Computed:    true,
				Description: "The custom branch name to use",
			},
			"deployment_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of deployment environment (currently 'production', 'staging' or empty)",
			},
			"extended_attributes_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the extended attributes applied",
			},
			"connection_id": schema.Int64Attribute{
				Computed: true,
				Description: "The ID of the connection to use (can be the `id` of a `dbtcloud_global_connection` or the `connection_id` of a legacy connection). " +
					"At the moment, it is optional and the environment will use the connection set in `dbtcloud_project_connection` if `connection_id` is not set in this resource. " +
					"In future versions this field will become required, so it is recommended to set it from now on. " +
					"When configuring this field, it needs to be configured for all the environments of the project. " +
					"To avoid Terraform state issues, when using this field, the `dbtcloud_project_connection` resource should be removed from the project or you need to make sure that the `connection_id` is the same in `dbtcloud_project_connection` and in the `connection_id` of the Development environment of the project",
			},
			"enable_model_query_history": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether model query history is on",
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
							Description: "Credential ID for this environment. A credential is not required for development environments, as dbt Cloud defaults to the user's credentials, but deployment environments will have this.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the environment",
						},
						"dbt_version": schema.StringAttribute{
							Computed:    true,
							Description: "Version number of dbt to use in this environment.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of environment (must be either development or deployment)",
						},
						"use_custom_branch": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether to use a custom git branch in this environment",
						},
						"custom_branch": schema.StringAttribute{
							Computed:    true,
							Description: "The custom branch name to use",
						},
						"deployment_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of deployment environment (currently 'production', 'staging' or empty)",
						},
						"extended_attributes_id": schema.Int64Attribute{
							Computed:    true,
							Description: "The ID of the extended attributes applied",
						},
						"connection_id": schema.Int64Attribute{
							Computed:    true,
							Description: "A connection ID (used with Global Connections)",
						},
						"enable_model_query_history": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether model query history is on",
						},
					},
				},
			},
		},
	}
}

func (r *environmentResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = resource_schema.Schema{
		Description: "Resource to manage dbt Cloud environments for the different dbt Cloud projects." +
			" In a given dbt Cloud project, one development environment can be defined and as many deployment environments as needed can be created." +
			" ~> In August 2024, dbt Cloud released the \"global connection\" feature, allowing connections to be defined at the account level and reused across environments and projects." +
			" This version of the provider has the connection_id as an optional field but it is recommended to start setting it up in your projects. In future versions, this field will become mandatory.",
		Attributes: map[string]resource_schema.Attribute{
			"id": resource_schema.StringAttribute{
				Computed:    true,
				Description: "The ID of environment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": resource_schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the environment. Duplicated. Here for backward compatibility.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_active": resource_schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the environment is active",
			},
			"project_id": resource_schema.Int64Attribute{
				Required:    true,
				Description: "Project ID to create the environment in",
			},
			"credential_id": resource_schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     nil,
				Description: "The Credential ID for this environment. A credential is not actionable for development environments, as users have to set their own development credentials in dbt Cloud.",
			},
			"name": resource_schema.StringAttribute{
				Required:    true,
				Description: "The name of the environment",
			},
			"dbt_version": resource_schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString("latest"),
				Description: "Version number of dbt to use in this environment. It needs to be in the format `major.minor.0-latest` (e.g. `1.5.0-latest`), `major.minor.0-pre`, `compatible`, `extended`, `versionless`, `latest` or `latest-fusion`. While `versionless` is still supported, using `latest` is recommended. Defaults to `latest` if no version is provided",
				Validators: []validator.String{
					helper.DbtVersionValidator{}, // Custom validator to check the dbt version format
				},
			},
			"type": resource_schema.StringAttribute{
				Required:    true,
				Description: "The type of environment (must be either development or deployment)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("development", "deployment"),
				},
			},
			"use_custom_branch": resource_schema.BoolAttribute{
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to use a custom git branch in this environment",
				Validators: []validator.Bool{
					validators.UseCustomBranchValidator{},
				},
			},
			"custom_branch": resource_schema.StringAttribute{
				Optional:    true,
				Description: "The custom branch name to use",
				Validators: []validator.String{
					validators.CustomBranchValidator{},
				},
			},
			"deployment_type": resource_schema.StringAttribute{
				Optional:    true,
				Description: "The type of environment. Only valid for environments of type 'deployment' and for now can only be 'production', 'staging' or left empty for generic environments",
			},
			"extended_attributes_id": resource_schema.Int64Attribute{
				Computed:    true,
				Optional:    true,
				Description: "The ID of the extended attributes applied",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"connection_id": resource_schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "A connection ID (used with Global Connections)",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"enable_model_query_history": resource_schema.BoolAttribute{
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to enable model query history in this environment. As of Oct 2024, works only for Snowflake and BigQuery.",
			},
		},
	}
}
