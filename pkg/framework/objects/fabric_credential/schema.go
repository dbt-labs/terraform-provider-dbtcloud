package fabric_credential

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var resourceSchema = resource_schema.Schema{
	Description: "Fabric credential resource",
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
			Description: "Project ID to create the Fabric credential in",
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
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "The username of the Fabric account to connect to. Only used when connection with AD user/pass",
			Validators: []validator.String{
				conflictingFieldsValidator{
					conflictingFields: []string{"tenant_id", "client_id", "client_secret"},
				},
			},
		},
		"password": resource_schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "The password for the account to connect to. Only used when connection with AD user/pass",
			Validators: []validator.String{
				conflictingFieldsValidator{
					conflictingFields: []string{"tenant_id", "client_id", "client_secret"},
				},
			},
		},
		"tenant_id": resource_schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "The tenant ID of the Azure Active Directory instance. This is only used when connecting to Azure SQL with a service principal.",
			Validators: []validator.String{
				conflictingFieldsValidator{
					conflictingFields: []string{"user", "password"},
				},
			},
		},
		"client_id": resource_schema.StringAttribute{
			Computed:    true,
			Optional:    true,
			Default:     stringdefault.StaticString(""),
			Description: "The client ID of the Azure Active Directory service principal. This is only used when connecting to Azure SQL with an AAD service principal.",
			Validators: []validator.String{
				conflictingFieldsValidator{
					conflictingFields: []string{"user", "password"},
				},
			},
		},
		"client_secret": resource_schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Sensitive:   true,
			Default:     stringdefault.StaticString(""),
			Description: "The client secret of the Azure Active Directory service principal. This is only used when connecting to Azure SQL with an AAD service principal.",
			Validators: []validator.String{
				conflictingFieldsValidator{
					conflictingFields: []string{"user", "password"},
				},
			},
		},
		"schema": resource_schema.StringAttribute{
			Required:    true,
			Description: "The schema where to create the dbt models",
		},
		"schema_authorization": resource_schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			Description: "Optionally set this to the principal who should own the schemas created by dbt",
		},
		"adapter_type": resource_schema.StringAttribute{
			Description: "The type of the adapter (fabric)",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf("fabric"),
			},
		},
	},
}

type conflictingFieldsValidator struct {
	conflictingFields []string
}

func (v conflictingFieldsValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("Field conflicts with: %v", v.conflictingFields)
}

func (v conflictingFieldsValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Field conflicts with: %v", v.conflictingFields)
}

func (v conflictingFieldsValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() {
		return
	}

	for _, field := range v.conflictingFields {
		var conflictingValue types.String
		req.Config.GetAttribute(ctx, path.Root(field), &conflictingValue)
		if !conflictingValue.IsNull() {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Conflicting Fields",
				fmt.Sprintf("Cannot set both %s and %s", req.Path, field),
			)
		}
	}
}
