package notification_setting

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestDelete_404IsTreatedAsSuccess is a regression test for #718: Delete used
// to surface any API error as a hard failure, including a 404 returned when
// the notification setting was already gone server-side, which failed
// terraform apply even though the resource had, in effect, been deleted.
func TestDelete_404IsTreatedAsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected request method %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status": {"code": 404, "is_success": false, "user_message": "not found"}}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse test server URL: %v", err)
	}
	client := &dbt_cloud.Client{
		HostURL:    u,
		HTTPClient: server.Client(),
		AccountID:  1,
		MaxRetries: 1,
	}

	r := &notificationSettingResource{client: client}

	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema error: %v", schemaResp.Diagnostics)
	}

	model := &NotificationSettingResourceModel{
		ID:          types.Int64Value(42),
		Name:        types.StringValue("test-setting"),
		Description: types.StringNull(),
		Channels:    []NotificationSettingChannelModel{},
		Rules:       []NotificationSettingRuleModel{},
	}

	state := tfsdk.State{Schema: schemaResp.Schema}
	if diags := state.Set(context.Background(), model); diags.HasError() {
		t.Fatalf("failed to build state fixture: %v", diags)
	}

	resp := &resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("regression: Delete failed on a 404 instead of treating it as an idempotent success: %v", resp.Diagnostics)
	}
}
