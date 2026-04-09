package teradata_credential

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var resourceSchema = resource_schema.Schema{
	Description: "Teradata credential resource",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_id": resource_schema.Int64Attribute{
			Required:    true,
			Description: "Project ID to create the Teradata/Trino credential in",
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
		"user": resource_schema.StringAttribute{
			Required:    true,
			Description: "The username for the Teradata account ",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"password": resource_schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "The password for the Teradata account. Consider using `password_wo` instead, which is not stored in state.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ConflictsWith(path.MatchRoot("password_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("password_wo")),
			},
		},
		"password_wo": resource_schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only alternative to `password`. The value is not stored in state. Requires `password_wo_version` to trigger updates.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ConflictsWith(path.MatchRoot("password")),
			},
		},
		"password_wo_version": resource_schema.Int64Attribute{
			Optional:    true,
			Description: "Version number for `password_wo`. Increment this value to trigger an update of the password when using `password_wo`.",
		},
		"schema": resource_schema.StringAttribute{
			Required:    true,
			Description: "The schema where to create models",
			Validators: []validator.String{
				helper.SchemaNameValidator(),
			},
		},
		"threads": resource_schema.Int64Attribute{
			Optional:    true,
			Description: "The number of threads to use. Default is 1",
			Computed:    true,
			Default:     int64default.StaticInt64(1),
		},
		"resource_metadata": resource_schema.DynamicAttribute{
			Optional:    true,
			Description: "Metadata for tracking resource identity during account migrations. Stored in Terraform state only and not sent to the API.",
		},
	},
}

var datasourceSchema = datasource_schema.Schema{
	Description: "Teradata credential data source",
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
		"user": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "The username for the Teradata account ",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"password": datasource_schema.StringAttribute{
			Computed:    true,
			Sensitive:   true,
			Description: "The password for the Teradata account",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"threads": datasource_schema.Int64Attribute{
			Description: "The number of threads to use. Default is 1",
			Computed:    true,
		},
	},
}
