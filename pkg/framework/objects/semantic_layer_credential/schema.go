package semantic_layer_credential

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/bigquery_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/databricks_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/postgres_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/redshift_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/snowflake_credential"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	config_resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
			Attributes:  bigquery_credential.SemanticLayerAttributes,
		},

		"private_key_id": resource_schema.StringAttribute{
			Required:    true,
			Description: "Private Key ID for the Service Account",
		},

		"private_key": resource_schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "Private Key for the Service Account. Consider using `private_key_wo` instead, which is not stored in state.",
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRoot("private_key_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("private_key_wo")),
			},
		},

		"private_key_wo": resource_schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only alternative to `private_key`. The value is not stored in state. Requires `private_key_wo_version` to trigger updates.",
		},

		"private_key_wo_version": resource_schema.Int64Attribute{
			Optional:    true,
			Description: "Version number for `private_key_wo`. Increment this value to trigger an update of the private key when using `private_key_wo`.",
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

		"execution_project": resource_schema.StringAttribute{
			Optional:    true,
			Description: "The GCP project that should execute BigQuery jobs for the semantic layer. When not set, jobs will execute in the project associated with the service account.",
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

var postgres_sl_credential_resource_schema = resource_schema.Schema{
	Description: "Postgres credential resource. This resource is composed of a Postgres credential and a Semantic Layer configuration. It is used to create a Postgres credential for the Semantic Layer.",
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
			Description: "Postgres credential details, but used in the context of the Semantic Layer.",
			Attributes:  postgres_credential.PostgresResourceSchema.Attributes, // Reuse the schema
		},
	},
}
