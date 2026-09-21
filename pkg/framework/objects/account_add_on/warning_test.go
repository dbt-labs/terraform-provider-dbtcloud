package account_add_on

import (
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
)

// TestNotLiveWarning checks which states warn. A trial ends on its own and a
// product can be cancelled elsewhere, and neither changes the configuration, so a
// read that stays quiet would leave the customer with a product they no longer
// have and an empty plan. The live states must stay quiet, or the warning becomes
// noise on every plan.
func TestNotLiveWarning(t *testing.T) {
	for _, testCase := range []struct {
		state string
		warns bool
	}{
		{dbt_cloud.AddOnStateActive, false},
		{dbt_cloud.AddOnStateTrial, false},
		{dbt_cloud.AddOnStateExpired, true},
		{dbt_cloud.AddOnStateCancelled, true},
		{"", false},
	} {
		warning := notLiveWarning("wizard", testCase.state)
		if got := warning != ""; got != testCase.warns {
			t.Errorf("notLiveWarning for %q warned = %v, want %v", testCase.state, got, testCase.warns)
		}
		if testCase.warns && !strings.Contains(warning, "wizard") {
			t.Errorf("the warning for %q does not name the product: %s", testCase.state, warning)
		}
	}
}
