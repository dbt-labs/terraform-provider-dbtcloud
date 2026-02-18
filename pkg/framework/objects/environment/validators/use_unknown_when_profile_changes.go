package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UseUnknownWhenProfileChanges is a plan modifier that sets a field to unknown
// when primary_profile_id is being newly set or changed. When the profile is
// stable (config matches state), the existing state value is preserved so the
// plan can settle. This prevents both "inconsistent result after apply" errors
// and perpetual diffs for connection_id, credential_id, and extended_attributes_id,
// which the API mirrors from the profile onto the environment.
type UseUnknownWhenProfileChanges struct{}

func (m UseUnknownWhenProfileChanges) Description(_ context.Context) string {
	return "Sets the value to unknown when primary_profile_id is newly set or changed, since the API will determine the value from the profile."
}

func (m UseUnknownWhenProfileChanges) MarkdownDescription(_ context.Context) string {
	return "Sets the value to unknown when `primary_profile_id` is newly set or changed, since the API will determine the value from the profile."
}

func (m UseUnknownWhenProfileChanges) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	var configProfileID types.Int64
	diags := req.Config.GetAttribute(ctx, path.Root("primary_profile_id"), &configProfileID)
	if diags.HasError() {
		return
	}

	// If primary_profile_id is not in the config, do nothing
	if configProfileID.IsNull() {
		return
	}

	// If the config profile is known and matches state, the profile hasn't
	// changed — preserve the existing state value so the plan settles.
	var stateProfileID types.Int64
	req.State.GetAttribute(ctx, path.Root("primary_profile_id"), &stateProfileID)

	if !configProfileID.IsUnknown() && configProfileID.Equal(stateProfileID) {
		resp.PlanValue = req.StateValue
		return
	}

	// Otherwise the profile is new, changing, or unknown (referencing an
	// un-created resource) — let the API determine this field's value.
	resp.PlanValue = types.Int64Unknown()
}
