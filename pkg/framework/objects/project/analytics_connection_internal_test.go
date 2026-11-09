package project

import (
	"encoding/json"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Rule: the analytics connection comes from the API response, and a response without the field reports nothing.
// given: a project response with the id, with null, and without the field
// when: the analytics connection is mapped with a planned value of 42
// then: the id and null are taken from the response, and the planned value is kept when the field is absent.
func TestAnalyticsConnectionFromProject(t *testing.T) {
	cases := map[string]struct {
		body string
		want types.Int64
	}{
		"reported":    {`{"analytics_connection_id":7}`, types.Int64Value(7)},
		"null":        {`{"analytics_connection_id":null}`, types.Int64Null()},
		"not in body": {`{}`, types.Int64Value(42)},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var project dbt_cloud.Project
			if err := json.Unmarshal([]byte(tc.body), &project); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			got := analyticsConnectionFromProject(&project, types.Int64Value(42))
			if !got.Equal(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
