package ip_restrictions_rule

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

type IPRestrictionsRuleResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Description    types.String `tfsdk:"description"`
	RuleSetEnabled types.Bool   `tfsdk:"rule_set_enabled"`
	Cidrs          []CidrModel  `tfsdk:"cidrs"`
}

type CidrModel struct {
	Cidr                types.String `tfsdk:"cidr"`
	CidrIpv6            types.String `tfsdk:"cidr_ipv6"`
	ID                  types.Int64  `tfsdk:"id"`
	IPRestrictionRuleID types.Int64  `tfsdk:"ip_restriction_rule_id"`
}

var ipRestrictionTypeNameToIDMapping = map[string]int64{
	"allow": 1,
	"deny":  2,
}

var ipRestrictionTypeIDToNameMapping = lo.Invert(ipRestrictionTypeNameToIDMapping)
