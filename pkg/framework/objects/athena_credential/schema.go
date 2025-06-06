package athena_credential

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

var resourceSchema = resource_schema.Schema{
	Description: "Athena credential resource",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"credential_id": resource_schema.Int64Attribute{
			Computed:    true,
			Description: "The internal credential ID",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"project_id": resource_schema.Int64Attribute{
			Required:    true,
			Description: "Project ID to create the Athena credential in",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"aws_access_key_id": resource_schema.StringAttribute{
			Required:    true,
			Sensitive:   true,
			Description: "AWS access key ID for Athena user",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"aws_secret_access_key": resource_schema.StringAttribute{
			Required:    true,
			Sensitive:   true,
			Description: "AWS secret access key for Athena user",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"schema": resource_schema.StringAttribute{
			Required:    true,
			Description: "The schema where to create models",
		},
		"adapter_version": resource_schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("athena_v0"),
			Description: "Adapter version (athena_v0, etc.)",
		},
	},
}

var datasourceSchema = datasource_schema.Schema{
	Description: "Athena credential data source",
	Attributes: map[string]datasource_schema.Attribute{
		"id": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
		},
		"credential_id": datasource_schema.Int64Attribute{
			Required:    true,
			Description: "Credential ID",
		},
		"project_id": datasource_schema.Int64Attribute{
			Required:    true,
			Description: "Project ID",
		},
		"schema": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "The schema where to create models",
		},
	},
}
