package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UseUnknownWhenProfileSet is a plan modifier that sets a field to unknown
// when primary_profile_id is configured. This prevents "inconsistent result
// after apply" errors because the API mirrors the profile's values onto the
// environment for connection_id, credential_id, and extended_attributes_id.
type UseUnknownWhenProfileSet struct{}

func (m UseUnknownWhenProfileSet) Description(_ context.Context) string {
	return "Sets the value to unknown when primary_profile_id is configured, since the API will determine the value from the profile."
}

func (m UseUnknownWhenProfileSet) MarkdownDescription(_ context.Context) string {
	return "Sets the value to unknown when `primary_profile_id` is configured, since the API will determine the value from the profile."
}

func (m UseUnknownWhenProfileSet) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	var profileID types.Int64
	diags := req.Config.GetAttribute(ctx, path.Root("primary_profile_id"), &profileID)
	if diags.HasError() {
		return
	}

	// If primary_profile_id is not in the config, do nothing
	if profileID.IsNull() {
		return
	}

	// If primary_profile_id is set (known value) or unknown (e.g. referencing
	// another resource), the API will determine this field's value from the
	// profile. Mark as unknown so Terraform shows "(known after apply)".
	resp.PlanValue = types.Int64Unknown()
}
