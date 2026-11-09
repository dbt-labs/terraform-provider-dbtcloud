package job

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// The API returns the spelling it stored, which can differ from the configured
// one. State must keep the configured spelling or every plan shows a diff.
func TestRead_KeepsConfiguredDbtVersionSpelling(t *testing.T) {
	testCases := []struct {
		configured string
		apiVersion string
		expected   string
	}{
		{configured: "latest", apiVersion: "v1-latest", expected: "latest"},
		{configured: "fusion-stable", apiVersion: "v2-stable", expected: "fusion-stable"},
		{configured: "v2-stable", apiVersion: "fusion-stable", expected: "v2-stable"},
		{configured: "latest", apiVersion: "v2-stable", expected: "v2-stable"},
	}

	for _, tc := range testCases {
		t.Run(tc.configured+"/"+tc.apiVersion, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatalf("unexpected request method %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data": {"id": 42, "project_id": 1, "environment_id": 2, "name": "job", "execute_steps": ["dbt build"], "dbt_version": "` + tc.apiVersion + `", "state": 1, "settings": {"threads": 4, "target_name": "default"}, "triggers": {}, "schedule": {"cron": "0 * * * *", "date": {"type": "every_day"}, "time": {"type": "every_hour", "interval": 1}}}, "status": {"code": 200, "is_success": true}}`))
			}))
			defer server.Close()

			u, err := url.Parse(server.URL)
			if err != nil {
				t.Fatalf("failed to parse test server URL: %v", err)
			}
			j := &jobResource{client: &dbt_cloud.Client{
				HostURL:    u,
				HTTPClient: server.Client(),
				AccountID:  1,
				MaxRetries: 1,
			}}

			var schemaResp resource.SchemaResponse
			j.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("unexpected schema error: %v", schemaResp.Diagnostics)
			}

			ctx := context.Background()
			state := tfsdk.State{
				Schema: schemaResp.Schema,
				Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
			}
			state.SetAttribute(ctx, path.Root("id"), types.Int64Value(42))
			state.SetAttribute(ctx, path.Root("dbt_version"), types.StringValue(tc.configured))

			resp := &resource.ReadResponse{State: state}
			j.Read(ctx, resource.ReadRequest{State: state}, resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("Read failed: %v", resp.Diagnostics)
			}

			var got types.String
			resp.State.GetAttribute(ctx, path.Root("dbt_version"), &got)
			if got.ValueString() != tc.expected {
				t.Fatalf("dbt_version in state = %q, want %q", got.ValueString(), tc.expected)
			}
		})
	}
}
