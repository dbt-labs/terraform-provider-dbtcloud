package snowflake_credential

import (
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

var (
	authTypes = []string{
		"password",
		"keypair",
	}
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

var resourceSchema = resource_schema.Schema{
	Description: "Snowflake credential resource",
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
				stringvalidator.OneOf(authTypes...),
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
			Required:    true,
			Description: "The schema where to create models",
		},
		"user": resource_schema.StringAttribute{
			Required:    true,
			Description: "The username for the Snowflake account ",
		},
		"password": resource_schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "The password for the Snowflake account",
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Validators: []validator.String{
				stringvalidator.ConflictsWith(
					path.MatchRoot("private_key"),
					path.MatchRoot("private_key_passphrase"),
				),
			},
		},
		"private_key": resource_schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "The private key for the Snowflake account",
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRoot("password")),
			},
		},
		"private_key_passphrase": resource_schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "The passphrase for the private key",
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRoot("password")),
			},
		},
		"num_threads": resource_schema.Int64Attribute{
			Required:    true,
			Description: "Number of threads to use",
		},
	},
}
