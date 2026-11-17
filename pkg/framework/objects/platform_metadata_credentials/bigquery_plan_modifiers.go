package platform_metadata_credentials

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// impersonationURLPlanModifier plans service_account_impersonation_url when it is not set in the
// configuration.
//
// dbt Cloud fills the URL from the connection when a WIF credential is created or switched to WIF,
// and only clears it when the credential switches to a service account key. Apply can only fill a
// value the plan left unknown, so the decision has to be made here:
//   - auth_type unknown: unknown;
//   - create, or auth_type changes: null for a service account key, unknown for WIF;
//   - auth_type unchanged: the value in state.
//
// UseStateForUnknown can't be used: it copies the state value even when auth_type changes.
type impersonationURLPlanModifier struct{}

func (m impersonationURLPlanModifier) Description(_ context.Context) string {
	return "Keeps the stored impersonation URL while auth_type is unchanged, and lets dbt Cloud fill it when switching to WIF."
}

func (m impersonationURLPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m impersonationURLPlanModifier) PlanModifyString(
	ctx context.Context,
	req planmodifier.StringRequest,
	resp *planmodifier.StringResponse,
) {
	// set in the configuration (or unknown there): keep the planned value
	if !req.ConfigValue.IsNull() {
		return
	}

	// nothing to plan on destroy
	if req.Plan.Raw.IsNull() {
		return
	}

	authTypePath := req.Path.ParentPath().AtName("auth_type")

	var plannedAuthType types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, authTypePath, &plannedAuthType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plannedAuthType.IsUnknown() {
		resp.PlanValue = types.StringUnknown()
		return
	}

	valueForNewAuthType := func() types.String {
		if plannedAuthType.ValueString() == bigQueryAuthTypeWIF {
			return types.StringUnknown()
		}
		return types.StringNull()
	}

	// create: test the whole state, not the URL's state value, which is also null for an
	// existing WIF credential without a URL
	if req.State.Raw.IsNull() {
		resp.PlanValue = valueForNewAuthType()
		return
	}

	var stateAuthType types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, authTypePath, &stateAuthType)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plannedAuthType.Equal(stateAuthType) {
		resp.PlanValue = valueForNewAuthType()
		return
	}

	resp.PlanValue = req.StateValue
}
