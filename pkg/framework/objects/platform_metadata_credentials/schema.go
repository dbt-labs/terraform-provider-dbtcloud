package platform_metadata_credentials

import (
	snowflake_credential_validators "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/snowflake_credential/validators"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
		"resource_metadata": schema.DynamicAttribute{
			Optional:    true,
			Description: "Metadata for tracking resource identity during account migrations. Stored in Terraform state only and not sent to the API.",
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
			Description: "The password for password authentication. Required when auth_type is 'password'. Cannot be used with private_key or private_key_passphrase. Consider using `password_wo` instead, which is not stored in state.",
			Optional:    true,
			Sensitive:   true,
			Validators: []validator.String{
				snowflake_credential_validators.ConflictValidator{
					ConflictingFields: []string{"private_key", "private_key_passphrase"},
				},
				stringvalidator.ConflictsWith(path.MatchRoot("password_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("password_wo")),
			},
		},
		"password_wo": schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only alternative to `password`. The value is not stored in state. Requires `password_wo_version` to trigger updates.",
			Validators: []validator.String{
				snowflake_credential_validators.ConflictValidator{
					ConflictingFields: []string{"private_key", "private_key_passphrase", "private_key_wo", "private_key_passphrase_wo"},
				},
			},
		},
		"password_wo_version": schema.Int64Attribute{
			Optional:    true,
			Description: "Version number for `password_wo`. Increment this value to trigger an update of the password when using `password_wo`.",
		},
		"private_key": schema.StringAttribute{
			Description: "The private key for keypair authentication. Required when auth_type is 'keypair'. Cannot be used with password. Consider using `private_key_wo` instead, which is not stored in state.",
			Optional:    true,
			Sensitive:   true,
			Validators: []validator.String{
				snowflake_credential_validators.ConflictValidator{
					ConflictingFields: []string{"password"},
				},
				stringvalidator.ConflictsWith(path.MatchRoot("private_key_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("private_key_wo")),
			},
		},
		"private_key_wo": schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only alternative to `private_key`. The value is not stored in state. Requires `private_key_wo_version` to trigger updates.",
			Validators: []validator.String{
				snowflake_credential_validators.ConflictValidator{
					ConflictingFields: []string{"password", "password_wo"},
				},
			},
		},
		"private_key_wo_version": schema.Int64Attribute{
			Optional:    true,
			Description: "Version number for `private_key_wo`. Increment this value to trigger an update of the private key when using `private_key_wo`.",
		},
		"private_key_passphrase": schema.StringAttribute{
			Description: "The passphrase for the private key, if encrypted. Optional when auth_type is 'keypair'. Cannot be used with password. Consider using `private_key_passphrase_wo` instead, which is not stored in state.",
			Optional:    true,
			Sensitive:   true,
			Validators: []validator.String{
				snowflake_credential_validators.ConflictValidator{
					ConflictingFields: []string{"password"},
				},
				stringvalidator.ConflictsWith(path.MatchRoot("private_key_passphrase_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("private_key_passphrase_wo")),
			},
		},
		"private_key_passphrase_wo": schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only alternative to `private_key_passphrase`. The value is not stored in state. Requires `private_key_passphrase_wo_version` to trigger updates.",
			Validators: []validator.String{
				snowflake_credential_validators.ConflictValidator{
					ConflictingFields: []string{"password", "password_wo"},
				},
			},
		},
		"private_key_passphrase_wo_version": schema.Int64Attribute{
			Optional:    true,
			Description: "Version number for `private_key_passphrase_wo`. Increment this value to trigger an update of the private key passphrase when using `private_key_passphrase_wo`.",
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
			Description: "The Databricks personal access token. Consider using `token_wo` instead, which is not stored in state.",
			Optional:    true,
			Sensitive:   true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRoot("token_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("token_wo")),
			},
		},
		"token_wo": schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only alternative to `token`. The value is not stored in state. Requires `token_wo_version` to trigger updates.",
		},
		"token_wo_version": schema.Int64Attribute{
			Optional:    true,
			Description: "Version number for `token_wo`. Increment this value to trigger an update of the token when using `token_wo`.",
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
