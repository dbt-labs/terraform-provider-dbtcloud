package postgres_credential

import (
	sl_cred_validator "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var datasourceSchema = datasource_schema.Schema{
	Description: "Postgres credential data source.",
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
			Description: "Whether the Postgres credential is active",
		},
		"default_schema": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "Default schema name",
		},
		"username": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "Username for Postgres",
		},
		"num_threads": datasource_schema.Int64Attribute{
			Computed:    true,
			Description: "Number of threads to use",
		},
	},
}

var warehouseTypes = []string{"postgres", "redshift"}

var PostgresResourceSchema = resource_schema.Schema{
	Description: "Postgres credential resource.",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
		},
		"is_active": resource_schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(true),
			Description: "Whether the Postgres/Redshift/AlloyDB credential is active",
		},
		"project_id": resource_schema.Int64Attribute{
			Required:    true,
			Description: "Project ID to create the Postgres/Redshift/AlloyDB credential in.",
		},
		"credential_id": resource_schema.Int64Attribute{
			Computed:    true,
			Description: "The system Postgres/Redshift/AlloyDB credential ID.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"type": resource_schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("postgres"),
			Description: "Type of connection. One of (postgres/redshift). Use postgres for alloydb connections. Optional only when semantic_layer_credential is set to true; otherwise, this field is required.",
			Validators: []validator.String{
				stringvalidator.OneOf(
					warehouseTypes...,
				),
				sl_cred_validator.SemanticLayerCredentialValidator{FieldName: "type"},
			},
		},
		"default_schema": resource_schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("default_schema"),
			Description: "Default schema name. Optional only when semantic_layer_credential is set to true; otherwise, this field is required.",
			Validators: []validator.String{
				sl_cred_validator.SemanticLayerCredentialValidator{FieldName: "default_schema"},
			},
		},
		"target_name": resource_schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("default"),
			Description: "Default schema name",
		},
		"username": resource_schema.StringAttribute{
			Required:    true,
			Description: "Username for Postgres/Redshift/AlloyDB",
		},
		"password": resource_schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "Password for Postgres/Redshift/AlloyDB",
		},
		"num_threads": resource_schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(0),
			Description: "Number of threads to use (required for Redshift)",
		},
		"semantic_layer_credential": resource_schema.BoolAttribute{
			Optional:    true,
			Description: "This field indicates that the credential is used as part of the Semantic Layer configuration. It is used to create a Postgres credential for the Semantic Layer.",
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
	},
}
