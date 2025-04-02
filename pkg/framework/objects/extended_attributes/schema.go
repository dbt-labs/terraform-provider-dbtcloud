package extended_attributes

import (
	"context"
	"encoding/json"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var dataSourceSchema = datasource_schema.Schema{
	Description: "Extended attributes data source",
	Attributes: map[string]datasource_schema.Attribute{
		"id": datasource_schema.StringAttribute{
			Description: "The ID of this resource. Contains the project ID and the credential ID.",
			Computed:    true,
		},
		"project_id": datasource_schema.Int64Attribute{
			Description: "Project ID",
			Required:    true,
		},
		"state": datasource_schema.Int64Attribute{
			Description: "The state of the extended attributes (1 = active, 2 = inactive)",
			Computed:    true,
		},
		"extended_attributes": datasource_schema.StringAttribute{
			Description: "Extended attributes",
			Computed:    true,
		},
		"extended_attributes_id": resource_schema.Int64Attribute{
			Description: "Extended attributes ID",
			Required:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
	},
}

var resourceSchema = resource_schema.Schema{
	Description: "Extended attributes resource",
	Attributes: map[string]resource_schema.Attribute{
		"id": resource_schema.StringAttribute{
			Description: "The ID of this resource. Contains the project ID and the extended attributes ID.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"extended_attributes_id": resource_schema.Int64Attribute{
			Description: "Extended attributes ID",
			Computed:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"project_id": resource_schema.Int64Attribute{
			Description: "Project ID to create the extended attributes in",
			Required:    true,
		},
		"state": resource_schema.Int64Attribute{
			Description: "The state of the extended attributes (1 = active, 2 = inactive)",
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(1),
		},
		"extended_attributes": resource_schema.StringAttribute{
			Description: "A JSON string listing the extended attributes mapping. The keys are the connections attributes available in the `profiles.yml` for a given adapter. Any fields entered will override connection details or credentials set on the environment or project. To avoid incorrect Terraform diffs, it is recommended to create this string using `jsonencode` in your Terraform code. (see example)",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				extendedAttributesPlanModifier{},
			},
		},
	},
}

type extendedAttributesPlanModifier struct{}

func (m extendedAttributesPlanModifier) Description(ctx context.Context) string {
	return "Suppresses diff when JSON content is semantically equal"
}

func (m extendedAttributesPlanModifier) MarkdownDescription(ctx context.Context) string {
	return "Suppresses diff when JSON content is semantically equal"
}

func (m extendedAttributesPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() {
		return
	}

	var oldJSON, newJSON map[string]interface{}
	if err := json.Unmarshal([]byte(req.StateValue.ValueString()), &oldJSON); err != nil {
		return
	}
	if err := json.Unmarshal([]byte(req.PlanValue.ValueString()), &newJSON); err != nil {
		return
	}

	// Compare the raw JSON strings after removing whitespace
	oldStr := helper.NormalizeJSONString(req.StateValue.ValueString())
	newStr := helper.NormalizeJSONString(req.PlanValue.ValueString())

	if oldStr == newStr {
		// Use the plan value to preserve the original formatting
		resp.PlanValue = req.PlanValue
	}
}
