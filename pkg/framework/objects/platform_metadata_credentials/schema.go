package platform_metadata_credentials

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Common attributes shared between Snowflake and Databricks resources
func commonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The unique identifier for this resource (account_id:credential_id).",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"credential_id": schema.Int64Attribute{
			Description: "The ID of the platform metadata credential.",
			Computed:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"connection_id": schema.Int64Attribute{
			Description: "The ID of the global connection this credential is associated with. Cannot be changed after creation.",
			Required:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"catalog_ingestion_enabled": schema.BoolAttribute{
			Description: "Whether catalog ingestion is enabled for this credential. When enabled, dbt Cloud will ingest metadata about tables, views, and other objects from your data warehouse.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"cost_optimization_enabled": schema.BoolAttribute{
			Description: "Whether cost optimization data collection is enabled for this credential.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"cost_insights_enabled": schema.BoolAttribute{
			Description: "Whether cost insights is enabled for this credential.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"adapter_version": schema.StringAttribute{
			Description: "The adapter version derived from the connection (e.g., 'snowflake_v0', 'databricks_v0'). This is read-only and determined by the connection.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

// SnowflakePlatformMetadataCredentialSchema returns the schema for Snowflake platform metadata credentials
var SnowflakePlatformMetadataCredentialSchema = schema.Schema{
	Description: helper.DocString(
		`Manages Snowflake platform metadata credentials for external metadata ingestion in dbt Cloud.
		
This resource configures credentials that allow dbt Cloud to connect directly to your Snowflake warehouse 
to ingest metadata outside of normal dbt project runs. This enables features like:

- **Catalog Ingestion**: Ingest metadata about tables/views not defined in dbt
- **Cost Optimization**: Query warehouse cost and performance data
- **Cost Insights**: Enhanced cost visibility and analysis

~> **Note:** At least one of ~~~catalog_ingestion_enabled~~~, ~~~cost_optimization_enabled~~~, or 
~~~cost_insights_enabled~~~ must be enabled for the credential to be usable.

~> **Note:** The ~~~connection_id~~~ cannot be changed after creation. To use a different connection, 
you must destroy and recreate the resource.`,
	),
	Attributes: mergeAttributes(commonAttributes(), map[string]schema.Attribute{
		"auth_type": schema.StringAttribute{
			Description: "The authentication type. Must be 'password' or 'keypair'.",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.OneOf("password", "keypair"),
			},
		},
		"user": schema.StringAttribute{
			Description: "The Snowflake user name.",
			Required:    true,
		},
		"password": schema.StringAttribute{
			Description: "The password for password authentication. Required when auth_type is 'password'.",
			Optional:    true,
			Sensitive:   true,
		},
		"private_key": schema.StringAttribute{
			Description: "The private key for keypair authentication. Required when auth_type is 'keypair'.",
			Optional:    true,
			Sensitive:   true,
		},
		"private_key_passphrase": schema.StringAttribute{
			Description: "The passphrase for the private key, if encrypted. Optional when auth_type is 'keypair'.",
			Optional:    true,
			Sensitive:   true,
		},
		"role": schema.StringAttribute{
			Description: "The Snowflake role to use.",
			Required:    true,
		},
		"warehouse": schema.StringAttribute{
			Description: "The Snowflake warehouse to use.",
			Required:    true,
		},
	}),
}

// DatabricksPlatformMetadataCredentialSchema returns the schema for Databricks platform metadata credentials
var DatabricksPlatformMetadataCredentialSchema = schema.Schema{
	Description: helper.DocString(
		`Manages Databricks platform metadata credentials for external metadata ingestion in dbt Cloud.
		
This resource configures credentials that allow dbt Cloud to connect directly to your Databricks workspace 
to ingest metadata outside of normal dbt project runs. This enables features like:

- **Catalog Ingestion**: Ingest metadata about tables/views not defined in dbt
- **Cost Optimization**: Query warehouse cost and performance data
- **Cost Insights**: Enhanced cost visibility and analysis

~> **Note:** At least one of ~~~catalog_ingestion_enabled~~~, ~~~cost_optimization_enabled~~~, or 
~~~cost_insights_enabled~~~ must be enabled for the credential to be usable.

~> **Note:** The ~~~connection_id~~~ cannot be changed after creation. To use a different connection, 
you must destroy and recreate the resource.`,
	),
	Attributes: mergeAttributes(commonAttributes(), map[string]schema.Attribute{
		"token": schema.StringAttribute{
			Description: "The Databricks personal access token.",
			Required:    true,
			Sensitive:   true,
		},
		"catalog": schema.StringAttribute{
			Description: "The Unity Catalog name to use.",
			Required:    true,
		},
	}),
}

// mergeAttributes combines two attribute maps
func mergeAttributes(base, additional map[string]schema.Attribute) map[string]schema.Attribute {
	result := make(map[string]schema.Attribute)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range additional {
		result[k] = v
	}
	return result
}
