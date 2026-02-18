package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// UseStateWhenConfigSet is a plan modifier for Optional+Computed attributes
// that preserves the prior state value only when the attribute is present in
// config. When the user removes the attribute from config (null), the planned
// value is set to null instead of carrying forward the prior state.
//
// This differs from the built-in UseStateForUnknown which always copies state
// into an unknown plan value, even when the config no longer sets the field.
type UseStateWhenConfigSet struct{}

func (m UseStateWhenConfigSet) Description(_ context.Context) string {
	return "Uses the prior state value when the attribute is present in config; sets null when removed from config."
}

func (m UseStateWhenConfigSet) MarkdownDescription(_ context.Context) string {
	return "Uses the prior state value when the attribute is present in config; sets null when removed from config."
}

func (m UseStateWhenConfigSet) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// If config explicitly sets the value (non-null), use state for unknown
	// plan values like UseStateForUnknown would.
	if !req.ConfigValue.IsNull() {
		if req.PlanValue.IsUnknown() && !req.StateValue.IsNull() {
			resp.PlanValue = req.StateValue
		}
		return
	}

	// Config is null — the user removed this attribute. Force null in the
	// plan so the provider's Update sees the removal.
	resp.PlanValue = req.ConfigValue
}
