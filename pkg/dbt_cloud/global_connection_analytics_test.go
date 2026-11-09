package dbt_cloud

import (
	"encoding/json"
	"testing"
)

// Rule: the analytics credential of a connection is never sent back to the API.
// given: a connection payload built for a create or update, without the read-only field
// when: it is serialized
// then: the body has no analytics_credentials_id key.
func TestGlobalConnectionCommon_NeverSendsAnalyticsCredentialsID(t *testing.T) {
	name := "conn"
	body, err := json.Marshal(GlobalConnectionCommon{Name: &name})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	parsed := map[string]any{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := parsed["analytics_credentials_id"]; ok {
		t.Fatalf("analytics_credentials_id must not be sent, got body %s", body)
	}
}
