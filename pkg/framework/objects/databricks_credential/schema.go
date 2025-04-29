package databricks_credential

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var dataSourceSchema = datasource_schema.Schema{
	Description: "Databricks credential data source",
	Attributes: map[string]datasource_schema.Attribute{
		"id": datasource_schema.StringAttribute{
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
			Computed:    true,
		},
		"project_id": datasource_schema.Int64Attribute{
			Description: "Project ID",
			Required:    true,
		},
		"credential_id": datasource_schema.Int64Attribute{
			Description: "Credential ID",
			Required:    true,
		},
		"target_name": datasource_schema.StringAttribute{
			Description: "Target name",
			Computed:    true,
		},
		"num_threads": datasource_schema.Int64Attribute{
			Description: "The number of threads to use",
			Computed:    true,
		},
		"catalog": datasource_schema.StringAttribute{
			Description: "The catalog where to create models",
			Computed:    true,
		},
		"schema": datasource_schema.StringAttribute{
			Description: "The schema where to create models",
			Computed:    true,
		},
		"adapter_type": datasource_schema.StringAttribute{
			Description: "The type of the adapter (databricks or spark)",
			Computed:    true,
		},
	},
}

var resourceSchema = resource_schema.Schema{
	Description: "Databricks credential resource",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_id": resource_schema.Int64Attribute{
			Description: "Project ID to create the Databricks credential in",
			Required:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"credential_id": resource_schema.Int64Attribute{
			Description: "The system Databricks credential ID",
			Computed:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"target_name": resource_schema.StringAttribute{
			Description:        "Target name",
			Optional:           true,
			Computed:           true,
			Default:            stringdefault.StaticString("default"),
			DeprecationMessage: "This field is deprecated at the environment level (it was never possible to set it in the UI) and will be removed in a future release. Please remove it and set the target name at the job level or leverage environment variables.",
		},
		"token": resource_schema.StringAttribute{
			Description: "Token for Databricks user",
			Required:    true,
			Sensitive:   true,
		},
		"catalog": resource_schema.StringAttribute{
			Description: "The catalog where to create models (only for the databricks adapter)",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
		},
		"schema": resource_schema.StringAttribute{
			Description: "The schema where to create models",
			Required:    true,
		},
		"adapter_type": resource_schema.StringAttribute{
			Description: "The type of the adapter (databricks or spark)",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf("databricks", "spark"),
			},
		},
	},
}
