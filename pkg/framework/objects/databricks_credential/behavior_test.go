package databricks_credential

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

// TestRead_404RemovesResourceFromState checks that a credential deleted
// outside of Terraform is dropped from the state on refresh, so the next plan
// recreates it instead of failing.
func TestRead_404RemovesResourceFromState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
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

	r := &databricksCredentialResource{client: client}

	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema error: %v", schemaResp.Diagnostics)
	}

	model := &DatabricksCredentialResourceModel{
		ID:                      types.StringValue("10:20"),
		CredentialID:            types.Int64Value(20),
		ProjectID:               types.Int64Value(10),
		TargetName:              types.StringValue("default"),
		Token:                   types.StringNull(),
		TokenWo:                 types.StringNull(),
		TokenWoVersion:          types.Int64Null(),
		Schema:                  types.StringValue("analytics"),
		Catalog:                 types.StringValue("main"),
		AdapterType:             types.StringValue("databricks"),
		SemanticLayerCredential: types.BoolValue(false),
	}

	state := tfsdk.State{Schema: schemaResp.Schema}
	if diags := state.Set(context.Background(), model); diags.HasError() {
		t.Fatalf("failed to build state fixture: %v", diags)
	}

	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read failed on a 404 instead of removing the resource from state: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("expected the resource to be removed from state after a 404")
	}
}
