package postgres_credential

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var datasourceSchema = datasource_schema.Schema{
	Description: "Postgres credential data source.",
	Attributes: map[string]datasource_schema.Attribute{
		"id" : datasource_schema.StringAttribute{
			Computed: true,
			Description: "The ID of this data source. Contains the project ID and the credential ID.",
		},
		"project_id": datasource_schema.Int64Attribute{
			Required: true,
			Description: "Project ID",
		},
		"credential_id": datasource_schema.Int64Attribute{
			Required: true,
			Description: "Credential ID",
		},
		"is_active": datasource_schema.BoolAttribute{
			Computed: true,
			Description: "Whether the Postgres credential is active",
		},
		"default_schema": datasource_schema.StringAttribute{
			Computed: true,
			Description: "Default schema name",
		},
		"username": datasource_schema.StringAttribute{
			Computed: true,
			Description: "Username for Postgres",
		},
		"num_threads": datasource_schema.Int64Attribute{
			Computed: true,
			Description: "Number of threads to use",
		},
	},
}

var warehouseTypes = []string{"postgres", "redshift"}

var resourceSchema = resource_schema.Schema{
	Description: "Postgres credential resource.",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Computed: true,
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
		},
		"is_active": resource_schema.BoolAttribute{
			Optional: true,
			Default:  booldefault.StaticBool(true),
			Description: "Whether the Postgres/Redshift/AlloyDB credential is active",
		},
		"project_id": resource_schema.Int64Attribute{
			Required: true,
			Description: "Project ID to create the Postgres/Redshift/AlloyDB credential in.",
		},
		"credential_id": resource_schema.Int64Attribute{
			Computed: true,
			Description: "The system Postgres/Redshift/AlloyDB credential ID.",
		},
		"type": resource_schema.StringAttribute{
			Required: true,
			Description: "Type of connection. One of (postgres/redshift). Use postgres for alloydb connections",
			Validators: []validator.String{
				stringvalidator.OneOf(
					warehouseTypes...,
				),
			},
		},
		"default_schema": resource_schema.StringAttribute{
			Required: true,
			Description: "Default schema name",
		},
		"target_name": resource_schema.StringAttribute{
			Optional: true,
			Default: stringdefault.StaticString("default"),
			Description: "Default schema name",
		},
		"username": resource_schema.StringAttribute{
			Required: true,
			Description: "Username for Postgres/Redshift/AlloyDB",
		},
		"password": resource_schema.StringAttribute{
			Optional: true,
			Sensitive: true,
			Description: "Password for Postgres/Redshift/AlloyDB",
		},
		"num_threads": resource_schema.Int64Attribute{
			Optional: true,
			Description: "Number of threads to use",
		},
	},
}