package athena_credential

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
			Optional:    true,
			Sensitive:   true,
			Description: "AWS access key ID for Athena user. Consider using `aws_access_key_id_wo` instead, which is not stored in state.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ConflictsWith(path.MatchRoot("aws_access_key_id_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("aws_access_key_id_wo")),
			},
		},
		"aws_access_key_id_wo": resource_schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only alternative to `aws_access_key_id`. The value is not stored in state. Requires `aws_access_key_id_wo_version` to trigger updates.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"aws_access_key_id_wo_version": resource_schema.Int64Attribute{
			Optional:    true,
			Description: "Version number for `aws_access_key_id_wo`. Increment this value to trigger an update of the AWS access key ID when using `aws_access_key_id_wo`.",
		},
		"aws_secret_access_key": resource_schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "AWS secret access key for Athena user. Consider using `aws_secret_access_key_wo` instead, which is not stored in state.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ConflictsWith(path.MatchRoot("aws_secret_access_key_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("aws_secret_access_key_wo")),
			},
		},
		"aws_secret_access_key_wo": resource_schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only alternative to `aws_secret_access_key`. The value is not stored in state. Requires `aws_secret_access_key_wo_version` to trigger updates.",
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		"aws_secret_access_key_wo_version": resource_schema.Int64Attribute{
			Optional:    true,
			Description: "Version number for `aws_secret_access_key_wo`. Increment this value to trigger an update of the AWS secret access key when using `aws_secret_access_key_wo`.",
		},
		"schema": resource_schema.StringAttribute{
			Required:    true,
			Description: "The schema where to create models",
			Validators: []validator.String{
				helper.SchemaNameValidator(),
			},
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
