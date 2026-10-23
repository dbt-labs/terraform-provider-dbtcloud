package global_connection

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newConnectionTestClient(t *testing.T, serverURL string) *dbt_cloud.Client {
	t.Helper()
	accountID, token := int64(123), "test-token"
	maxRetries, retryInterval, timeout := 1, 0, 5
	client, err := dbt_cloud.NewClient(
		&accountID, &token, &serverURL, &maxRetries, &retryInterval, nil, true, &timeout,
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

// Rule: a connection shows the analytics credential attached to it.
// given: the API returns a Snowflake connection with, without, or with a null analytics credential
// when: the connection is read
// then: the state carries that credential id, or null.
func TestReadGeneric_AnalyticsCredentialsID(t *testing.T) {
	cases := map[string]struct {
		field string
		want  types.Int64
	}{
		"attached": {`,"analytics_credentials_id":42`, types.Int64Value(42)},
		"null":     {`,"analytics_credentials_id":null`, types.Int64Null()},
		"absent":   {``, types.Int64Null()},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"data":{"id":9,"name":"conn","private_link_endpoint_id":null,"oauth_configuration_id":null,"config":{"role":null}` + tc.field + `}}`))
			}))
			defer srv.Close()

			state := &GlobalConnectionResourceModel{ID: types.Int64Value(9)}
			got, _, err := readGeneric(newConnectionTestClient(t, srv.URL), state, "snowflake_v0")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.AnalyticsCredentialsId.Equal(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got.AnalyticsCredentialsId)
			}
		})
	}
}
