package ip_restrictions_rule

import (
	"context"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/helper"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *ipRestrictionsRuleResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: helper.DocString(`
            Manages IP restriction rules in dbt Cloud. IP restriction rules allow you to control access to your dbt Cloud instance based on IP address ranges.
        `),
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the IP restriction rule",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the IP restriction rule",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the IP restriction rule (allow or deny)",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"allow",
						"deny",
					),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the IP restriction rule",
			},
			"rule_set_enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the IP restriction rule set is enabled or not. Important!: This value needs to be the same for all rules if multiple rules are defined. All rules must be active or inactive at the same time.",
			},
			"cidrs": schema.SetNestedAttribute{
				Required:    true,
				Description: "Set of CIDR ranges for this rule",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cidr": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "IP CIDR range (can be IPv4 or IPv6)",
						},
						"cidr_ipv6": schema.StringAttribute{
							Computed:    true,
							Description: "IPv6 CIDR range (read-only)",
						},
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "ID of the CIDR range",
						},
						"ip_restriction_rule_id": schema.Int64Attribute{
							Computed:    true,
							Description: "ID of the IP restriction rule",
						},
					},
				},
			},
		},
	}
}
