package redshift_credential

import (
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var RedshiftResourceSchema = resource_schema.Schema{
	Description: "Redshift credential resource",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"is_active": resource_schema.BoolAttribute{
			Computed:    true,
			Optional:    true,
			Default:     booldefault.StaticBool(true),
			Description: "Whether the Redshift credential is active",
		},
		"project_id": resource_schema.Int64Attribute{
			Required:    true,
			Description: "Project ID to create the Redshift credential in",
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

		"username": resource_schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("default_user"),
			Description: "The username for the Redshift account. ",
		},
		"password": resource_schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "The password for the Redshift account",
			Computed:    true,
			Default:     stringdefault.StaticString(""),
		},
		"default_schema": resource_schema.StringAttribute{
			Required:    true,
			Description: "Default schema name",
		},
		"num_threads": resource_schema.Int64Attribute{
			Required:    true,
			Description: "Number of threads to use",
		},
	},
}

var RedshiftDatasourceSchema = datasource_schema.Schema{
	Description: "Redshift credential data source",
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
			Description: "Whether the Redshift credential is active",
		},
		"num_threads": datasource_schema.Int64Attribute{
			Computed:    true,
			Description: "Number of threads to use",
		},
		"default_schema": datasource_schema.StringAttribute{
			Required:    true,
			Description: "Default schema name",
		},
	},
}
