package snowflake_credential

import (
	snowflake_credential "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/snowflake_credential/validators"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var datasourceSchema = datasource_schema.Schema{
	Description: "Snowflake credential data source",
	Attributes: map[string]datasource_schema.Attribute{
		"id": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
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
			Description: "Whether the Snowflake credential is active",
		},
		"auth_type": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "The type of Snowflake credential ('password' or 'keypair')",
		},
		"schema": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "The schema where to create models",
		},
		"user": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "Username for Snowflake",
		},
		"num_threads": datasource_schema.Int64Attribute{
			Computed:    true,
			Description: "Number of threads to use",
		},
	},
}

var SnowflakeCredentialResourceSchema = resource_schema.Schema{
	Description: "Snowflake credential resource. This resource is used both as a stand-alone credential, but also as part of the Semantic Layer credential definition for Snowflake.",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"is_active": resource_schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(true),
			Description: "Whether the Snowflake credential is active",
		},
		"project_id": resource_schema.Int64Attribute{
			Required:    true,
			Description: "Project ID to create the Snowflake credential in",
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
		"auth_type": resource_schema.StringAttribute{
			Required:    true,
			Description: "The type of Snowflake credential ('password' or 'keypair')",
			Validators: []validator.String{
				stringvalidator.OneOf(AuthTypes...),
			},
		},
		"database": resource_schema.StringAttribute{
			Optional:    true,
			Description: "The catalog to connect use",
		},
		"role": resource_schema.StringAttribute{
			Optional:    true,
			Description: "The role to assume",
		},
		"warehouse": resource_schema.StringAttribute{
			Optional:    true,
			Description: "The warehouse to use",
		},
		"schema": resource_schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("default_schema"),
			Description: "The schema where to create models. This is an optional field ONLY if the credential is used for Semantic Layer configuration, otherwise it is required.",
			Validators: []validator.String{
				helper.SchemaNameValidator(),
			},
		},
		"user": resource_schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("default_user"),
			Description: "The username for the Snowflake account. This is an optional field ONLY if the credential is used for Semantic Layer configuration, otherwise it is required. ",
		},
		"password": resource_schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "The password for the Snowflake account. Consider using `password_wo` instead, which is not stored in state.",
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Validators: []validator.String{
				snowflake_credential.ConflictValidator{ConflictingFields: []string{"private_key", "private_key_passphrase"}},
				stringvalidator.ConflictsWith(path.MatchRoot("password_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("password_wo")),
			},
		},
		"password_wo": resource_schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only alternative to `password`. The value is not stored in state. Requires `password_wo_version` to trigger updates.",
			Validators: []validator.String{
				snowflake_credential.ConflictValidator{ConflictingFields: []string{"private_key", "private_key_passphrase", "private_key_wo", "private_key_passphrase_wo"}},
			},
		},
		"password_wo_version": resource_schema.Int64Attribute{
			Optional:    true,
			Description: "Version number for `password_wo`. Increment this value to trigger an update of the password when using `password_wo`.",
		},
		"private_key": resource_schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "The private key for the Snowflake account. Consider using `private_key_wo` instead, which is not stored in state.",
			Validators: []validator.String{
				snowflake_credential.ConflictValidator{ConflictingFields: []string{"password"}},
				stringvalidator.ConflictsWith(path.MatchRoot("private_key_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("private_key_wo")),
			},
		},
		"private_key_wo": resource_schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only alternative to `private_key`. The value is not stored in state. Requires `private_key_wo_version` to trigger updates.",
			Validators: []validator.String{
				snowflake_credential.ConflictValidator{ConflictingFields: []string{"password", "password_wo"}},
			},
		},
		"private_key_wo_version": resource_schema.Int64Attribute{
			Optional:    true,
			Description: "Version number for `private_key_wo`. Increment this value to trigger an update of the private key when using `private_key_wo`.",
		},
		"private_key_passphrase": resource_schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "The passphrase for the private key. Consider using `private_key_passphrase_wo` instead, which is not stored in state.",
			Validators: []validator.String{
				snowflake_credential.ConflictValidator{ConflictingFields: []string{"password"}},
				stringvalidator.ConflictsWith(path.MatchRoot("private_key_passphrase_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("private_key_passphrase_wo")),
			},
		},
		"private_key_passphrase_wo": resource_schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only alternative to `private_key_passphrase`. The value is not stored in state. Requires `private_key_passphrase_wo_version` to trigger updates.",
			Validators: []validator.String{
				snowflake_credential.ConflictValidator{ConflictingFields: []string{"password", "password_wo"}},
			},
		},
		"private_key_passphrase_wo_version": resource_schema.Int64Attribute{
			Optional:    true,
			Description: "Version number for `private_key_passphrase_wo`. Increment this value to trigger an update of the private key passphrase when using `private_key_passphrase_wo`.",
		},
		"num_threads": resource_schema.Int64Attribute{
			Required:    true,
			Description: "Number of threads to use",
		},
		"semantic_layer_credential": resource_schema.BoolAttribute{
			Optional:    true,
			Description: "This field indicates that the credential is used as part of the Semantic Layer configuration. It is used to create a Snowflake credential for the Semantic Layer.",
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Validators: []validator.Bool{
				snowflake_credential.SemanticLayerCredentialValidator{},
			},
		},
		"resource_metadata": resource_schema.DynamicAttribute{
			Optional:    true,
			Description: "Metadata for tracking resource identity during account migrations. Stored in Terraform state only and not sent to the API.",
		},
	},
}
