package clickhouse_credential

import (
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
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
	Description: "ClickHouse credential data source",
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
		"user": datasource_schema.StringAttribute{
			Description: "The ClickHouse username",
			Computed:    true,
		},
		"schema": datasource_schema.StringAttribute{
			Description: "The schema where to create the dbt models",
			Computed:    true,
		},
		"target_name": datasource_schema.StringAttribute{
			Description: "Target name",
			Computed:    true,
		},
		"threads": datasource_schema.Int64Attribute{
			Description: "The number of threads to use for dbt operations",
			Computed:    true,
		},
	},
}

var ClickhouseResourceSchema = resource_schema.Schema{
	Description: "ClickHouse credential resource",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_id": resource_schema.Int64Attribute{
			Description: "Project ID to create the ClickHouse credential in",
			Required:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"credential_id": resource_schema.Int64Attribute{
			Description: "The system ClickHouse credential ID",
			Computed:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"user": resource_schema.StringAttribute{
			Description: "The ClickHouse username",
			Required:    true,
		},
		"password": resource_schema.StringAttribute{
			Description: "The ClickHouse password. Consider using `password_wo` instead, which is not stored in state.",
			Optional:    true,
			Sensitive:   true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRoot("password_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("password_wo")),
			},
		},
		"password_wo": resource_schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only alternative to `password`. The value is not stored in state. Requires `password_wo_version` to trigger updates.",
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRoot("password")),
			},
		},
		"password_wo_version": resource_schema.Int64Attribute{
			Optional:    true,
			Description: "Version number for `password_wo`. Increment this value to trigger an update of the password when using `password_wo`.",
		},
		"schema": resource_schema.StringAttribute{
			Description: "The schema where to create the dbt models",
			Required:    true,
			Validators: []validator.String{
				helper.SchemaNameValidator(),
			},
		},
		"target_name": resource_schema.StringAttribute{
			Description: "Target name",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("default"),
		},
		"threads": resource_schema.Int64Attribute{
			Description: "The number of threads to use for dbt operations.",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
	},
}
