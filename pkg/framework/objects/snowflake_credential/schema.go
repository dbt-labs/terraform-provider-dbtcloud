package snowflake_credential

import (
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
