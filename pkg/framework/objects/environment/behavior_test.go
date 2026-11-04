package environment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestDelete_404IsTreatedAsSuccess is a regression test: Delete used to
// surface any API error as a hard failure, including a 404 returned when the
// environment was already gone server-side (e.g. removed via a cascading
// delete elsewhere), which failed terraform apply even though the resource
// had, in effect, been deleted.
func TestDelete_404IsTreatedAsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected request method %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status": {"code": 404, "is_success": false, "user_message": "The requested resource was not found. Please check that you have the proper permissions."}}`))
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

	r := &environmentResource{client: client}

	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema error: %v", schemaResp.Diagnostics)
	}

	model := &EnvironmentResourceModel{
		ID:                      types.StringValue("123:456"),
		EnvironmentID:           types.Int64Value(456),
		ProjectID:               types.Int64Value(123),
		CredentialID:            types.Int64Null(),
		IsActive:                types.BoolValue(true),
		Name:                    types.StringValue("test-env"),
		DbtVersion:              types.StringValue("latest"),
		Type:                    types.StringValue("deployment"),
		UseCustomBranch:         types.BoolValue(false),
		CustomBranch:            types.StringNull(),
		DeploymentType:          types.StringNull(),
		ExtendedAttributesID:    types.Int64Null(),
		ConnectionID:            types.Int64Null(),
		EnableModelQueryHistory: types.BoolValue(false),
		PrimaryProfileID:        types.Int64Null(),
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

// TestRead_404IsTreatedAsSuccess is a regression test: Read used to surface
// any API error as a hard failure, including a 404 returned when the
// environment was already gone server-side. Since terraform destroy (and
// plan) always refresh state via Read before Delete runs, this hard failure
// occurred even after Delete was made idempotent to 404s, e.g. when a prior
// CI/CD delete had actually succeeded server-side but the client timed out
// and retried.
func TestRead_404IsTreatedAsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected request method %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status": {"code": 404, "is_success": false, "user_message": "The requested resource was not found. Please check that you have the proper permissions."}}`))
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

	r := &environmentResource{client: client}

	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema error: %v", schemaResp.Diagnostics)
	}

	model := &EnvironmentResourceModel{
		ID:                      types.StringValue("123:456"),
		EnvironmentID:           types.Int64Value(456),
		ProjectID:               types.Int64Value(123),
		CredentialID:            types.Int64Null(),
		IsActive:                types.BoolValue(true),
		Name:                    types.StringValue("test-env"),
		DbtVersion:              types.StringValue("latest"),
		Type:                    types.StringValue("deployment"),
		UseCustomBranch:         types.BoolValue(false),
		CustomBranch:            types.StringNull(),
		DeploymentType:          types.StringNull(),
		ExtendedAttributesID:    types.Int64Null(),
		ConnectionID:            types.Int64Null(),
		EnableModelQueryHistory: types.BoolValue(false),
		PrimaryProfileID:        types.Int64Null(),
	}

	state := tfsdk.State{Schema: schemaResp.Schema}
	if diags := state.Set(context.Background(), model); diags.HasError() {
		t.Fatalf("failed to build state fixture: %v", diags)
	}

	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("regression: Read failed on a 404 instead of treating it as an idempotent success: %v", resp.Diagnostics)
	}
}

func newVersionTestResource(t *testing.T, handler http.HandlerFunc) (*environmentResource, resource.SchemaResponse) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse test server URL: %v", err)
	}
	r := &environmentResource{client: &dbt_cloud.Client{
		HostURL:    u,
		HTTPClient: server.Client(),
		AccountID:  1,
		MaxRetries: 1,
	}}

	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema error: %v", schemaResp.Diagnostics)
	}
	return r, schemaResp
}

func versionTestModel(dbtVersion string) *EnvironmentResourceModel {
	return &EnvironmentResourceModel{
		ID:                      types.StringValue("123:456"),
		EnvironmentID:           types.Int64Value(456),
		ProjectID:               types.Int64Value(123),
		CredentialID:            types.Int64Null(),
		IsActive:                types.BoolValue(true),
		Name:                    types.StringValue("test-env"),
		DbtVersion:              types.StringValue(dbtVersion),
		Type:                    types.StringValue("deployment"),
		UseCustomBranch:         types.BoolValue(false),
		CustomBranch:            types.StringNull(),
		DeploymentType:          types.StringNull(),
		ExtendedAttributesID:    types.Int64Null(),
		ConnectionID:            types.Int64Null(),
		EnableModelQueryHistory: types.BoolValue(false),
		PrimaryProfileID:        types.Int64Null(),
	}
}

func environmentJSON(dbtVersion string) []byte {
	return []byte(`{"data": {"id": 456, "account_id": 1, "project_id": 123, "name": "test-env", "dbt_version": "` + dbtVersion + `", "type": "deployment", "use_custom_branch": false, "state": 1}, "status": {"code": 200, "is_success": true}}`)
}

// The API returns the spelling it stored, which can differ from the configured
// one (the UI saved v1-latest, config says latest). State must keep the
// configured spelling or every plan shows a diff.
func TestRead_KeepsConfiguredDbtVersionSpelling(t *testing.T) {
	testCases := []struct {
		configured string
		apiVersion string
		expected   string
	}{
		{configured: "latest", apiVersion: "v1-latest", expected: "latest"},
		{configured: "versionless", apiVersion: "latest", expected: "versionless"},
		{configured: "fusion-stable", apiVersion: "v2-stable", expected: "fusion-stable"},
		{configured: "v2-stable", apiVersion: "fusion-stable", expected: "v2-stable"},
		{configured: "latest", apiVersion: "v2-stable", expected: "v2-stable"},
	}

	for _, tc := range testCases {
		t.Run(tc.configured+"/"+tc.apiVersion, func(t *testing.T) {
			r, schemaResp := newVersionTestResource(t, func(w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodGet {
					t.Fatalf("unexpected request method %s", req.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(environmentJSON(tc.apiVersion))
			})

			state := tfsdk.State{Schema: schemaResp.Schema}
			if diags := state.Set(context.Background(), versionTestModel(tc.configured)); diags.HasError() {
				t.Fatalf("failed to build state fixture: %v", diags)
			}

			resp := &resource.ReadResponse{State: state}
			r.Read(context.Background(), resource.ReadRequest{State: state}, resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("Read failed: %v", resp.Diagnostics)
			}

			var got types.String
			resp.State.GetAttribute(context.Background(), path.Root("dbt_version"), &got)
			if got.ValueString() != tc.expected {
				t.Fatalf("dbt_version in state = %q, want %q", got.ValueString(), tc.expected)
			}
		})
	}
}

// Update sends dbt_version only when the config changed. An unchanged config
// must not overwrite the spelling the API stored, and a changed one must reach
// the API so a spelling the account can't save fails at apply.
func TestUpdate_SendsDbtVersionOnlyWhenConfigChanges(t *testing.T) {
	testCases := []struct {
		name       string
		state      string
		plan       string
		apiVersion string
		wantSent   string
	}{
		{name: "unchanged config keeps the stored alias", state: "latest", plan: "latest", apiVersion: "v1-latest", wantSent: "v1-latest"},
		{name: "unchanged versionless keeps latest", state: "versionless", plan: "versionless", apiVersion: "latest", wantSent: "latest"},
		{name: "new version is sent", state: "latest", plan: "v2-stable", apiVersion: "latest", wantSent: "v2-stable"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var sent map[string]any
			r, schemaResp := newVersionTestResource(t, func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch req.Method {
				case http.MethodGet:
					_, _ = w.Write(environmentJSON(tc.apiVersion))
				case http.MethodPost:
					if err := json.NewDecoder(req.Body).Decode(&sent); err != nil {
						t.Fatalf("failed to decode update body: %v", err)
					}
					_, _ = w.Write(environmentJSON(sent["dbt_version"].(string)))
				default:
					t.Fatalf("unexpected request method %s", req.Method)
				}
			})

			state := tfsdk.State{Schema: schemaResp.Schema}
			if diags := state.Set(context.Background(), versionTestModel(tc.state)); diags.HasError() {
				t.Fatalf("failed to build state fixture: %v", diags)
			}
			plan := tfsdk.Plan{Schema: schemaResp.Schema}
			if diags := plan.Set(context.Background(), versionTestModel(tc.plan)); diags.HasError() {
				t.Fatalf("failed to build plan fixture: %v", diags)
			}

			resp := &resource.UpdateResponse{State: state}
			r.Update(context.Background(), resource.UpdateRequest{State: state, Plan: plan}, resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("Update failed: %v", resp.Diagnostics)
			}

			if sent["dbt_version"] != tc.wantSent {
				t.Fatalf("dbt_version sent = %v, want %q", sent["dbt_version"], tc.wantSent)
			}
			var got types.String
			resp.State.GetAttribute(context.Background(), path.Root("dbt_version"), &got)
			if got.ValueString() != tc.plan {
				t.Fatalf("dbt_version in state = %q, want %q", got.ValueString(), tc.plan)
			}
		})
	}
}
