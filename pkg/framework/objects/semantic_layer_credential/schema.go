package semantic_layer_credential

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/bigquery_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/databricks_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/redshift_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/snowflake_credential"
	config_resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var semantic_layer_config_resource_schema = config_resource_schema.Schema{
	Description: "Semantic Layer credential metadata. This contains the configuration for the semantic layer credential, but it is different than the Semantic Layer Configuration resource. It is used as part of the credential.",
	Attributes: map[string]config_resource_schema.Attribute{
		"name": config_resource_schema.StringAttribute{
			Required:    true,
			Description: "The name of the configuration",
		},
		"project_id": config_resource_schema.Int64Attribute{
			Required:    true,
			Description: "The ID of the project",
		},
		"adapter_version": config_resource_schema.StringAttribute{
			Required:    true,
			Description: "The adapter version",
		},
	},
}

var snowflake_sl_credential_resource_schema = resource_schema.Schema{
	Description: "Snowflake credential resource. This resource is composed of a Snowflake credential and a Semantic Layer configuration. It is used to create a Snowflake credential for the Semantic Layer.",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of the credential",
		},
		"configuration": resource_schema.SingleNestedAttribute{
			Required:    true,
			Description: "Semantic Layer credenttial configuration details.",
			Attributes:  semantic_layer_config_resource_schema.Attributes, // Reuse the schema
		},

		"credential": resource_schema.SingleNestedAttribute{
			Required:    true,
			Description: "Snowflake credential details, but used in the context of the Semantic Layer.",
			Attributes:  snowflake_credential.SnowflakeCredentialResourceSchema.Attributes, // Reuse the schema
		},
	},
}

var bigquery_sl_credential_resource_schema = resource_schema.Schema{
	Description: "BigQuery credential resource. This resource is composed of a BigQuery credential and a Semantic Layer configuration. It is used to create a BigQuery credential for the Semantic Layer.",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of the credential",
		},
		"configuration": resource_schema.SingleNestedAttribute{
			Required:    true,
			Description: "Semantic Layer credential configuration details.",
			Attributes:  semantic_layer_config_resource_schema.Attributes, // Reuse the schema
		},

		"credential": resource_schema.SingleNestedAttribute{
			Required:    true,
			Description: "BigQuery credential details, but used in the context of the Semantic Layer.",
			Attributes:  bigquery_credential.BigQueryResourceSchema.Attributes,
		},

		"private_key_id": resource_schema.StringAttribute{
			Required:    true,
			Description: "Private Key ID for the Service Account",
		},

		"private_key": resource_schema.StringAttribute{
			Required:    true,
			Sensitive:   true,
			Description: "Private Key for the Service Account",
		},

		"client_email": resource_schema.StringAttribute{
			Required:    true,
			Description: "Service Account email",
		},

		"client_id": resource_schema.StringAttribute{
			Required:    true,
			Description: "Client ID of the Service Account",
		},

		"auth_uri": resource_schema.StringAttribute{
			Required:    true,
			Description: "Auth URI for the Service Account",
		},

		"token_uri": resource_schema.StringAttribute{
			Required:    true,
			Description: "Token URI for the Service Account",
		},

		"auth_provider_x509_cert_url": resource_schema.StringAttribute{
			Required:    true,
			Description: "Auth Provider X509 Cert URL for the Service Account",
		},

		"client_x509_cert_url": resource_schema.StringAttribute{
			Required:    true,
			Description: "Client X509 Cert URL for the Service Account",
		},
	},
}

var redshift_sl_credential_resource_schema = resource_schema.Schema{
	Description: "Redshift credential resource. This resource is composed of a Redshift credential and a Semantic Layer configuration. It is used to create a Redshift credential for the Semantic Layer.",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of the credential",
		},
		"configuration": resource_schema.SingleNestedAttribute{
			Required:    true,
			Description: "Semantic Layer credential configuration details.",
			Attributes:  semantic_layer_config_resource_schema.Attributes, // Reuse the schema
		},
		"credential": resource_schema.SingleNestedAttribute{
			Required:    true,
			Description: "Redshift credential details, but used in the context of the Semantic Layer.",
			Attributes:  redshift_credential.RedshiftResourceSchema.Attributes, // Reuse the schema
		},
	},
}

var databricks_sl_credential_resource_schema = resource_schema.Schema{
	Description: "Databricks credential resource. This resource is composed of a Databricks credential and a Semantic Layer configuration. It is used to create a Databricks credential for the Semantic Layer.",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of the credential",
		},
		"configuration": resource_schema.SingleNestedAttribute{
			Required:    true,
			Description: "Semantic Layer credential configuration details.",
			Attributes:  semantic_layer_config_resource_schema.Attributes, // Reuse the schema
		},
		"credential": resource_schema.SingleNestedAttribute{
			Required:    true,
			Description: "Databricks credential details, but used in the context of the Semantic Layer.",
			Attributes:  databricks_credential.DatabricksResourceSchema.Attributes, // Reuse the schema
		},
	},
}
