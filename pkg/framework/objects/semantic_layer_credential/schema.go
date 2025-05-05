package semantic_layer_credential

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/snowflake_credential"
	config_resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	snowflake_sl_credential__resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

var snowflake_sl_credential_resource_schema = snowflake_sl_credential__resource_schema.Schema{
	Description: "Snowflake credential resource. This resource is composed of a Snowflake credential and a Semantic Layer configuration. It is used to create a Snowflake credential for the Semantic Layer.",
	Attributes: map[string]snowflake_sl_credential__resource_schema.Attribute{
		"id": snowflake_sl_credential__resource_schema.Int64Attribute{
			Computed:    true,
			Description: "The ID of the credential",
		},
		"configuration": snowflake_sl_credential__resource_schema.SingleNestedAttribute{
			Required:    true,
			Description: "Semantic Layer credenttial configuration details.",
			Attributes:  semantic_layer_config_resource_schema.Attributes, // Reuse the schema
		},

		"credential": snowflake_sl_credential__resource_schema.SingleNestedAttribute{
			Required:    true,
			Description: "Snowflake credential details, but used in the context of the Semantic Layer.",
			Attributes:  snowflake_credential.SnowflakeCredentialResourceSchema.Attributes, // Reuse the schema
		},
	},
}
