package account_add_on

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	activationPaid  = "paid"
	activationTrial = "trial"
)

type AccountAddOnResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Product               types.String `tfsdk:"product"`
	Activation            types.String `tfsdk:"activation"`
	SpendLimitNanodollars types.Int64  `tfsdk:"spend_limit_nanodollars"`
	State                 types.String `tfsdk:"state"`
	TrialStartedAt        types.String `tfsdk:"trial_started_at"`
	TrialEndsAt           types.String `tfsdk:"trial_ends_at"`
	TrialConsumed         types.Bool   `tfsdk:"trial_consumed"`
	CanActivate           types.Bool   `tfsdk:"can_activate"`
	CanTrial              types.Bool   `tfsdk:"can_trial"`
}

type AccountAddOnDataSourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Product               types.String `tfsdk:"product"`
	SpendLimitNanodollars types.Int64  `tfsdk:"spend_limit_nanodollars"`
	State                 types.String `tfsdk:"state"`
	TrialStartedAt        types.String `tfsdk:"trial_started_at"`
	TrialEndsAt           types.String `tfsdk:"trial_ends_at"`
	TrialConsumed         types.Bool   `tfsdk:"trial_consumed"`
	CanActivate           types.Bool   `tfsdk:"can_activate"`
	CanTrial              types.Bool   `tfsdk:"can_trial"`
}

type AccountAddOnsDataSourceModel struct {
	AddOns []AccountAddOnDataSourceModel `tfsdk:"add_ons"`
}
