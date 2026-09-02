package bigquery_credential

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var BigQueryResourceSchema = resource_schema.Schema{
	Description: "Bigquery credential resource",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"is_active": resource_schema.BoolAttribute{
			Computed:    true,
			Optional:    true,
			Default:     booldefault.StaticBool(true),
			Description: "Whether the BigQuery credential is active",
		},
		"project_id": resource_schema.Int64Attribute{
			Required:    true,
			Description: "Project ID to create the BigQuery credential in",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"credential_id": resource_schema.Int64Attribute{
			Computed:    true,
			Description: "The internal credential ID",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"dataset": resource_schema.StringAttribute{
			Required:    true,
			Description: "Default dataset name",
			Validators: []validator.String{
				helper.SchemaNameValidator(),
			},
		},
		"num_threads": resource_schema.Int64Attribute{
			Required:    true,
			Description: "Number of threads to use",
		},
		"connection_id": resource_schema.Int64Attribute{
			Optional:    true,
			Description: "The ID of the global connection to use for this credential. When provided, the credential will automatically use the correct adapter version based on the connection's configuration (e.g., bigquery_v1 for connections with use_latest_adapter=true).",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"auth_type": resource_schema.StringAttribute{
			Optional:    true,
			Description: "The authentication method for the BigQuery credential. Supported values: `service-account-json`, `oauth-secrets`, `external-oauth-wif`. Only applicable for v1 credentials (when `connection_id` is set).",
			Validators: []validator.String{
				stringvalidator.OneOf(AuthTypes...),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"workload_pool_provider_path": resource_schema.StringAttribute{
			Optional:    true,
			Description: "The fully specified resource name of the workload pool provider, required when `auth_type` is `external-oauth-wif`. Only applicable for v1 credentials (when `connection_id` is set).",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"service_account_impersonation_url": resource_schema.StringAttribute{
			Optional:    true,
			Description: "The URL for the service account impersonation request, used when `auth_type` is `external-oauth-wif`. Only applicable for v1 credentials (when `connection_id` is set).",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	},
}

var datasourceSchema = datasource_schema.Schema{
	Description: "Bigquery credential data source",
	Attributes: map[string]datasource_schema.Attribute{
		"id": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this data source. Contains the project ID and the credential ID.",
		},
		"project_id": datasource_schema.Int64Attribute{
			Required:    true,
			Description: "Project ID",
		},
		"credential_id": datasource_schema.Int64Attribute{
			Required:    true,
			Description: "Credential ID",
		},
		"is_active": datasource_schema.BoolAttribute{
			Computed:    true,
			Description: "Whether the BigQuery credential is active",
		},
		"dataset": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "Default dataset name",
		},
		"num_threads": datasource_schema.Int64Attribute{
			Computed:    true,
			Description: "Number of threads to use",
		},
		"auth_type": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "The authentication method for the BigQuery credential",
		},
		"workload_pool_provider_path": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "The fully specified resource name of the workload pool provider",
		},
		"service_account_impersonation_url": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "The URL for the service account impersonation request",
		},
	},
}
