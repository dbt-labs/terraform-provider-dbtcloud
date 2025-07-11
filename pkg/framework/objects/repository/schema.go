package repository

import (
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func ResourceSchema() resource_schema.Schema {
	return resource_schema.Schema{
		Description: "Manages a dbt Cloud repository.",
		Attributes: map[string]resource_schema.Attribute{
			"id": resource_schema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"repository_id": resource_schema.Int64Attribute{
				Computed:    true,
				Description: "Repository Identifier",
			},
			"is_active": resource_schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the repository is active",
			},
			"project_id": resource_schema.Int64Attribute{
				Required:    true,
				Description: "Project ID to create the repository in",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"remote_url": resource_schema.StringAttribute{
				Required:    true,
				Description: "Git URL for the repository or \\<Group>/\\<Project> for Gitlab",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"git_clone_strategy": resource_schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("deploy_key"),
				Description: "Git clone strategy for the repository. Can be `deploy_key` (default) for cloning via SSH Deploy Key, `github_app` for GitHub native integration, `deploy_token` for the GitLab native integration and `azure_active_directory_app` for ADO native integration",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repository_credentials_id": resource_schema.Int64Attribute{
				Computed:    true,
				Description: "Credentials ID for the repository (From the repository side not the dbt Cloud ID)",
			},
			"gitlab_project_id": resource_schema.Int64Attribute{
				Optional:    true,
				Description: "Identifier for the Gitlab project -  (for GitLab native integration only)",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"github_installation_id": resource_schema.Int64Attribute{
				Optional:    true,
				Description: "Identifier for the GitHub App - (for GitHub native integration only)",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"private_link_endpoint_id": resource_schema.StringAttribute{
				Optional:    true,
				Description: "Identifier for the PrivateLink endpoint.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"azure_active_directory_project_id": resource_schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "The Azure Dev Ops project ID. It can be retrieved via the Azure API or using the data source `dbtcloud_azure_dev_ops_project` and the project name - (for ADO native integration only)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"azure_active_directory_repository_id": resource_schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "The Azure Dev Ops repository ID. It can be retrieved via the Azure API or using the data source `dbtcloud_azure_dev_ops_repository` along with the ADO Project ID and the repository name - (for ADO native integration only)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"azure_bypass_webhook_registration_failure": resource_schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If set to False (the default), the connection will fail if the service user doesn't have access to set webhooks (required for auto-triggering CI jobs). If set to True, the connection will be successful but no automated CI job will be triggered - (for ADO native integration only)",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"fetch_deploy_key": resource_schema.BoolAttribute{
				Optional:           true,
				Computed:           true,
				Default:            booldefault.StaticBool(false),
				Description:        "Whether we should return the public deploy key - (for the `deploy_key` strategy)",
				DeprecationMessage: "This field is deprecated and will be removed in a future version of the provider, please remove it from your configuration. The key is always fetched when the clone strategy is `deploy_key`",
			},
			"deploy_key": resource_schema.StringAttribute{
				Computed:    true,
				Description: "Public key generated by dbt when using `deploy_key` clone strategy",
			},
			"pull_request_url_template": resource_schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "URL template for creating a pull request. If it is not set, the default template will create a PR from the current branch to the branch configured in the Development environment.",
			},
		},
	}
}

func DataSourceSchema() datasource_schema.Schema {
	return datasource_schema.Schema{
		Description: "Retrieve data for a single repository",
		Attributes: map[string]datasource_schema.Attribute{
			"id": datasource_schema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource",
			},
			"repository_id": datasource_schema.Int64Attribute{
				Required:    true,
				Description: "ID for the repository",
			},
			"project_id": datasource_schema.Int64Attribute{
				Required:    true,
				Description: "Project ID to create the repository in",
			},
			"fetch_deploy_key": datasource_schema.BoolAttribute{
				Optional:           true,
				Computed:           true,
				Description:        "Whether we should return the public deploy key",
				DeprecationMessage: "This field is deprecated and will be removed in a future version of the provider. The key is always fetched when the clone strategy is `deploy_key`",
			},
			"is_active": datasource_schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the repository is active",
			},
			"remote_url": datasource_schema.StringAttribute{
				Computed:    true,
				Description: "Git URL for the repository or \\<Group>/\\<Project> for Gitlab",
			},
			"git_clone_strategy": datasource_schema.StringAttribute{
				Computed:    true,
				Description: "Git clone strategy for the repository",
			},
			"repository_credentials_id": datasource_schema.Int64Attribute{
				Computed:    true,
				Description: "Credentials ID for the repository (From the repository side not the dbt Cloud ID)",
			},
			"gitlab_project_id": datasource_schema.Int64Attribute{
				Computed:    true,
				Description: "Identifier for the Gitlab project",
			},
			"github_installation_id": datasource_schema.Int64Attribute{
				Computed:    true,
				Description: "Identifier for the GitHub installation",
			},
			"private_link_endpoint_id": datasource_schema.StringAttribute{
				Computed:    true,
				Description: "Identifier for the PrivateLink endpoint.",
			},
			"deploy_key": datasource_schema.StringAttribute{
				Computed:    true,
				Description: "Public key generated by dbt when using `deploy_key` clone strategy",
			},
			"pull_request_url_template": datasource_schema.StringAttribute{
				Computed:    true,
				Description: "The pull request URL template to be used when opening a pull request from within dbt Cloud's IDE",
			},
			"azure_active_directory_project_id": datasource_schema.StringAttribute{
				Computed:    true,
				Description: "The Azure Dev Ops project ID",
			},
			"azure_active_directory_repository_id": datasource_schema.StringAttribute{
				Computed:    true,
				Description: "The Azure Dev Ops repository ID",
			},
			"azure_bypass_webhook_registration_failure": datasource_schema.BoolAttribute{
				Computed:    true,
				Description: "If set to False (the default), the connection will fail if the service user doesn't have access to set webhooks",
			},
		},
	}
}
