package webhook

import (
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var dataSourceSchema = datasource_schema.Schema{
	Description: "Retrieve webhook details",
	Attributes: map[string]datasource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Computed:    true,
			Description: "Webhook's ID",
		},
		"webhook_id": datasource_schema.StringAttribute{
			Required:           true,
			Description:        "Webhook's ID",
			DeprecationMessage: "Use `id` instead",
		},
		"name": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "Webhooks Name",
		},
		"description": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "Webhooks Description",
		},
		"client_url": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "Webhooks Client URL",
		},
		"event_types": datasource_schema.ListAttribute{
			Computed:    true,
			Description: "Webhooks Event Types",
			ElementType: types.StringType,
		},
		"job_ids": datasource_schema.ListAttribute{
			Computed:    true,
			Description: "List of job IDs to trigger the webhook",
			ElementType: types.Int64Type,
		},
		"active": datasource_schema.BoolAttribute{
			Computed:    true,
			Description: "Webhooks active flag",
		},
		"http_status_code": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "Webhooks HTTP Status Code",
		},
		"account_identifier": datasource_schema.StringAttribute{
			Computed:    true,
			Description: "Webhooks Account Identifier",
		},
	},
}
var resourceSchema = resource_schema.Schema{
	Description: "Webhook details",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Computed:    true,
			Description: "Webhook's ID",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"webhook_id": resource_schema.StringAttribute{
			Computed:           true,
			Description:        "Webhook's ID",
			DeprecationMessage: "Use `id` instead",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": resource_schema.StringAttribute{
			Description: "Webhooks Name",
			Required:    true,
		},
		"description": resource_schema.StringAttribute{
			Description: "Webhooks Description",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
		},
		"client_url": resource_schema.StringAttribute{
			Description: "Webhooks Client URL",
			Required:    true,
		},
		"event_types": resource_schema.ListAttribute{
			Description: "Webhooks Event Types",
			ElementType: types.StringType,
			Required:    true,
		},
		"job_ids": resource_schema.ListAttribute{
			Description: "List of job IDs to trigger the webhook. When null or empty, the webhook will trigger on all jobs",
			ElementType: types.Int64Type,
			Optional:    true,
		},
		"active": resource_schema.BoolAttribute{
			Description: "Webhooks active flag",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(true),
		},
		"http_status_code": resource_schema.StringAttribute{
			Description: "Latest HTTP status of the webhook",
			Computed:    true,
		},
		"hmac_secret": resource_schema.StringAttribute{
			Computed:    true,
			Sensitive:   true,
			Description: "Secret key for the webhook. Can be used to validate the authenticity of the webhook.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"account_identifier": resource_schema.StringAttribute{
			Description: "Webhooks Account Identifier",
			Computed:    true,
		},
	},
}
