package ip_restrictions_rule

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

var (
	_ resource.Resource                = &ipRestrictionsRuleResource{}
	_ resource.ResourceWithConfigure   = &ipRestrictionsRuleResource{}
	_ resource.ResourceWithImportState = &ipRestrictionsRuleResource{}
)

func IPRestrictionsRuleResource() resource.Resource {
	return &ipRestrictionsRuleResource{}
}

type ipRestrictionsRuleResource struct {
	client *dbt_cloud.Client
}

func (r *ipRestrictionsRuleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_ip_restrictions_rule"
}

func (r *ipRestrictionsRuleResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*dbt_cloud.Client)
}

func (r *ipRestrictionsRuleResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state IPRestrictionsRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetIPRestrictionsRule(state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading IP Restrictions Rule",
			err.Error(),
		)
		return
	}

	if rule == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(rule.Name)
	state.Type = types.StringValue(ipRestrictionTypeIDToNameMapping[rule.Type])
	state.Description = types.StringValue(rule.Description)
	state.RuleSetEnabled = types.BoolValue(rule.RuleSetEnabled)

	state.Cidrs = make([]CidrModel, 0, len(rule.Cidrs))
	for _, cidr := range rule.Cidrs {
		state.Cidrs = append(state.Cidrs, CidrModel{
			Cidr:                types.StringValue(cidr.Cidr),
			CidrIpv6:            types.StringValue(cidr.CidrIpv6),
			ID:                  types.Int64Value(cidr.ID),
			IPRestrictionRuleID: types.Int64Value(cidr.IPRestrictionRuleID),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
func (r *ipRestrictionsRuleResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan IPRestrictionsRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipRestriction := dbt_cloud.IPRestrictionsRule{
		Name:           plan.Name.ValueString(),
		Type:           ipRestrictionTypeNameToIDMapping[plan.Type.ValueString()],
		Description:    plan.Description.ValueString(),
		RuleSetEnabled: plan.RuleSetEnabled.ValueBool(),
		Cidrs:          make([]dbt_cloud.Cidrs, 0, len(plan.Cidrs)),
	}

	for _, cidr := range plan.Cidrs {
		ipRestriction.Cidrs = append(ipRestriction.Cidrs, dbt_cloud.Cidrs{
			Cidr:                cidr.Cidr.ValueString(),
			CidrIpv6:            cidr.CidrIpv6.ValueString(),
			ID:                  cidr.ID.ValueInt64(),
			IPRestrictionRuleID: cidr.IPRestrictionRuleID.ValueInt64(),
		})
	}

	created, err := r.client.CreateIPRestrictionsRule(ipRestriction)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating IP Restrictions Rule",
			err.Error(),
		)
		return
	}

	plan.ID = types.Int64Value(created.ID)
	plan.Cidrs = make([]CidrModel, 0, len(created.Cidrs))

	for _, cidr := range created.Cidrs {
		plan.Cidrs = append(plan.Cidrs, CidrModel{
			Cidr:                types.StringValue(cidr.Cidr),
			CidrIpv6:            types.StringValue(cidr.CidrIpv6),
			ID:                  types.Int64Value(cidr.ID),
			IPRestrictionRuleID: types.Int64Value(cidr.IPRestrictionRuleID),
		})
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ipRestrictionsRuleResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan IPRestrictionsRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state IPRestrictionsRuleResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipRestrictionsRule := dbt_cloud.IPRestrictionsRule{
		ID:             plan.ID.ValueInt64(),
		Name:           plan.Name.ValueString(),
		Type:           ipRestrictionTypeNameToIDMapping[plan.Type.ValueString()],
		Description:    plan.Description.ValueString(),
		RuleSetEnabled: plan.RuleSetEnabled.ValueBool(),
		Cidrs:          []dbt_cloud.Cidrs{},
	}

	for _, cidr := range plan.Cidrs {
		ipRestrictionsRule.Cidrs = append(ipRestrictionsRule.Cidrs, dbt_cloud.Cidrs{
			Cidr:                cidr.Cidr.ValueString(),
			CidrIpv6:            cidr.CidrIpv6.ValueString(),
			ID:                  cidr.ID.ValueInt64(),
			IPRestrictionRuleID: cidr.IPRestrictionRuleID.ValueInt64(),
		})
	}

	for _, cidr := range state.Cidrs {
		foundInPlan := lo.Filter(plan.Cidrs, func(c CidrModel, _ int) bool {
			return c.ID.ValueInt64() == cidr.ID.ValueInt64()
		})
		if len(foundInPlan) == 0 {
			ipRestrictionsRule.Cidrs = append(ipRestrictionsRule.Cidrs, dbt_cloud.Cidrs{
				Cidr:                cidr.Cidr.ValueString(),
				CidrIpv6:            cidr.CidrIpv6.ValueString(),
				ID:                  cidr.ID.ValueInt64(),
				IPRestrictionRuleID: cidr.IPRestrictionRuleID.ValueInt64(),
				State:               dbt_cloud.STATE_DELETED,
			})

		}

	}

	created, err := r.client.UpdateIPRestrictionsRule(
		strconv.FormatInt(plan.ID.ValueInt64(), 10),
		ipRestrictionsRule,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating IP Restrictions Rule",
			err.Error(),
		)
		return
	}
	plan.Cidrs = make([]CidrModel, 0, len(created.Cidrs))

	for _, cidr := range created.Cidrs {
		plan.Cidrs = append(plan.Cidrs, CidrModel{
			Cidr:                types.StringValue(cidr.Cidr),
			CidrIpv6:            types.StringValue(cidr.CidrIpv6),
			ID:                  types.Int64Value(cidr.ID),
			IPRestrictionRuleID: types.Int64Value(cidr.IPRestrictionRuleID),
		})
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ipRestrictionsRuleResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state IPRestrictionsRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteIPRestrictionsRule(state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting IP Restrictions Rule",
			err.Error(),
		)
		return
	}
}

func (r *ipRestrictionsRuleResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing IP Restrictions Rule",
			fmt.Sprintf("Invalid ID format: %s. The ID should be an integer", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
