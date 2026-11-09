package helper

import (
	"context"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DbtVersionValidator struct{}

func (v DbtVersionValidator) Description(ctx context.Context) string {
	return "Validates that the dbt_version is in the format " + validDbtVersionsDescription + "."
}

func (v DbtVersionValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that the `dbt_version` is in the format " + validDbtVersionsDescription + "."
}

func (v DbtVersionValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip validation if the value is unknown or null
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	// Get the value of dbt_version
	dbtVersion := req.ConfigValue.ValueString()

	if !validDbtVersionPattern.MatchString(dbtVersion) {
		resp.Diagnostics.AddError(
			"Invalid dbt_version Format",
			fmt.Sprintf("The `dbt_version` must be in the format "+validDbtVersionsDescription+". Got: %s", dbtVersion),
		)
	}
}

// IsFusionVersion reports whether the supplied dbt_version string is one of
// the Fusion release tracks.
func IsFusionVersion(dbtVersion string) bool {
	switch CanonicalDbtVersion(dbtVersion) {
	case "latest-fusion", "fusion-stable", "fusion-extended", "fusion-nightly", "fusion-fallback":
		return true
	}
	return false
}

// The release tracks the API accepts under their own name. Aliases of them
// live in dbtVersionAliases.
var dbtReleaseTracks = []string{
	"compatible",
	"extended",
	"latest",
	"fallback",
	"latest-fusion",
	"fusion-stable",
	"fusion-extended",
	"fusion-nightly",
	"fusion-fallback",
}

// Mirrors dbt-cloud's RELEASE_TRACK_ALIASES, plus versionless, which the API
// saves as latest.
var dbtVersionAliases = map[string]string{
	"versionless":   "latest",
	"v1-latest":     "latest",
	"v1-compatible": "compatible",
	"v1-extended":   "extended",
	"v2-stable":     "fusion-stable",
	"v2-extended":   "fusion-extended",
	"v2-nightly":    "fusion-nightly",
	"v2-fallback":   "fusion-fallback",
}

var validDbtVersionPattern, validDbtVersionsDescription = dbtVersionRules()

func dbtVersionRules() (*regexp.Regexp, string) {
	aliases := slices.Sorted(maps.Keys(dbtVersionAliases))
	names := append(slices.Clone(dbtReleaseTracks), aliases...)
	quoted := make([]string, len(names))
	for i, name := range names {
		quoted[i] = regexp.QuoteMeta(name)
	}
	pattern := regexp.MustCompile(`^(` + strings.Join(quoted, "|") + `|[0-9]+\.[0-9]+\.0-(latest|pre))$`)

	description := "`major.minor.0-latest`, `major.minor.0-pre`, `" +
		strings.Join(dbtReleaseTracks, "`, `") + "`, or the equivalent `" +
		strings.Join(aliases, "`, `") + "`"
	return pattern, description
}

func CanonicalDbtVersion(dbtVersion string) string {
	if canonical, ok := dbtVersionAliases[dbtVersion]; ok {
		return canonical
	}
	return dbtVersion
}

// DbtVersionFromAPI keeps the configured spelling when it names the same version
// as the API's, so a config of latest shows no diff when the API returns v1-latest.
func DbtVersionFromAPI(configured types.String, apiVersion string) types.String {
	if !configured.IsNull() && !configured.IsUnknown() &&
		CanonicalDbtVersion(configured.ValueString()) == CanonicalDbtVersion(apiVersion) {
		return configured
	}
	return types.StringValue(apiVersion)
}

// KeepEquivalentDbtVersion plans the prior state's spelling when the config
// names the same version, which the protocol allows. After an import, state
// holds the API's spelling. Not used on dbtcloud_job: a plan that differs from
// state marks job_type and cost_optimization_features unknown, and this would
// keep that diff on every plan.
type KeepEquivalentDbtVersion struct{}

func (m KeepEquivalentDbtVersion) Description(ctx context.Context) string {
	return "Keeps the stored dbt_version when the configured one names the same version."
}

func (m KeepEquivalentDbtVersion) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m KeepEquivalentDbtVersion) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	if CanonicalDbtVersion(req.PlanValue.ValueString()) == CanonicalDbtVersion(req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}
