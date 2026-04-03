package salesforce_credential

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var dataSourceSchema = datasource_schema.Schema{
	Description: "Salesforce credential data source",
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
		"username": datasource_schema.StringAttribute{
			Description: "The Salesforce username for OAuth JWT bearer flow authentication",
			Computed:    true,
		},
		"target_name": datasource_schema.StringAttribute{
			Description: "Target name",
			Computed:    true,
		},
		"num_threads": datasource_schema.Int64Attribute{
			Description: "The number of threads to use for dbt operations",
			Computed:    true,
		},
	},
}

var SalesforceResourceSchema = resource_schema.Schema{
	Description: "Salesforce credential resource",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"project_id": resource_schema.Int64Attribute{
			Description: "Project ID to create the Salesforce credential in",
			Required:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"credential_id": resource_schema.Int64Attribute{
			Description: "The system Salesforce credential ID",
			Computed:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"username": resource_schema.StringAttribute{
			Description: "The Salesforce username for OAuth JWT bearer flow authentication",
			Required:    true,
		},
		"client_id": resource_schema.StringAttribute{
			Description: "The OAuth connected app client/consumer ID. Consider using `client_id_wo` instead, which is not stored in state.",
			Optional:    true,
			Sensitive:   true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRoot("client_id_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("client_id_wo")),
			},
		},
		"client_id_wo": resource_schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only alternative to `client_id`. The value is not stored in state. Requires `client_id_wo_version` to trigger updates.",
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRoot("client_id")),
			},
		},
		"client_id_wo_version": resource_schema.Int64Attribute{
			Optional:    true,
			Description: "Version number for `client_id_wo`. Increment this value to trigger an update of the client ID when using `client_id_wo`.",
		},
		"private_key": resource_schema.StringAttribute{
			Description: "The private key for JWT bearer flow authentication. Consider using `private_key_wo` instead, which is not stored in state.",
			Optional:    true,
			Sensitive:   true,
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRoot("private_key_wo")),
				stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("private_key_wo")),
			},
		},
		"private_key_wo": resource_schema.StringAttribute{
			Optional:    true,
			WriteOnly:   true,
			Description: "Write-only alternative to `private_key`. The value is not stored in state. Requires `private_key_wo_version` to trigger updates.",
			Validators: []validator.String{
				stringvalidator.ConflictsWith(path.MatchRoot("private_key")),
			},
		},
		"private_key_wo_version": resource_schema.Int64Attribute{
			Optional:    true,
			Description: "Version number for `private_key_wo`. Increment this value to trigger an update of the private key when using `private_key_wo`.",
		},
		"target_name": resource_schema.StringAttribute{
			Description: "Target name",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("default"),
		},
		"num_threads": resource_schema.Int64Attribute{
			Description: "The number of threads to use for dbt operations",
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(6),
		},
	},
}
