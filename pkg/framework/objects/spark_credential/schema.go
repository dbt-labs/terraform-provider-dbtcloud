package spark_credential

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var dataSourceSchema = datasource_schema.Schema{
	Description: "Apache Spark credential data source",
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
		"schema": datasource_schema.StringAttribute{
			Description: "The schema where to create models",
			Computed:    true,
		},
	},
}

var SparkResourceSchema = resource_schema.Schema{
	Description: "Apache Spark credential resource",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_id": resource_schema.Int64Attribute{
			Description: "Project ID to create the Apache Spark credential in",
			Required:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"credential_id": resource_schema.Int64Attribute{
			Description: "The system Apache Spark credential ID",
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
			Description: "Token for Apache Spark user. Consider using `token_wo` instead, which is not stored in state.",
			Optional:    true,
			Sensitive:   true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRoot("token_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("token_wo")),
			},
		},
		"token_wo": resource_schema.StringAttribute{
			Description: "Write-only alternative to `token`. The value is not stored in state. Requires `token_wo_version` to trigger updates.",
			Optional:    true,
			WriteOnly:   true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRoot("token")),
			},
		},
		"token_wo_version": resource_schema.Int64Attribute{
			Description: "Version number for `token_wo`. Increment this value to trigger an update of the token when using `token_wo`.",
			Optional:    true,
		},
		"schema": resource_schema.StringAttribute{
			Description: "The schema where to create models",
			Required:    true,
		},
		"resource_metadata": resource_schema.DynamicAttribute{
			Optional:    true,
			Description: "Metadata for tracking resource identity during account migrations. Stored in Terraform state only and not sent to the API.",
		},
	},
}
