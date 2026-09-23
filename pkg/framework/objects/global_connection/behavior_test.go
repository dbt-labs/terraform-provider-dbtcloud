package global_connection

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// These are plain Go unit tests (no TF_ACC / live API needed) that exercise the
// exact code paths fixed for GitHub issue #709: importing/refreshing a
// bigquery_v1 dbtcloud_global_connection left use_latest_adapter unset in
// state, which the ModifyPlan guard then misread as a user-requested adapter
// change, and Update unconditionally sent the bigquery_v0-only
// timeout_seconds field even for bigquery_v1 connections.

func newTestClient(t *testing.T, server *httptest.Server) *dbt_cloud.Client {
	t.Helper()
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse test server URL: %v", err)
	}
	return &dbt_cloud.Client{
		HostURL:    u,
		HTTPClient: server.Client(),
		AccountID:  1,
		MaxRetries: 1,
	}
}

func testResourceSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	r := &globalConnectionResource{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema error: %v", resp.Diagnostics)
	}
	return resp
}

func bigQueryModel(useLatestAdapter bool, timeoutSeconds int64) *GlobalConnectionResourceModel {
	return &GlobalConnectionResourceModel{
		ID:                    types.Int64Value(123),
		AdapterVersion:        types.StringValue("bigquery_v1"),
		Name:                  types.StringValue("bq-conn"),
		IsSshTunnelEnabled:    types.BoolValue(false),
		PrivateLinkEndpointId: types.StringNull(),
		OauthConfigurationId:  types.Int64Null(),
		BigQueryConfig: &BigQueryConfig{
			GCPProjectID:               types.StringValue("my-project"),
			TimeoutSeconds:             types.Int64Value(timeoutSeconds),
			PrivateKeyID:               types.StringNull(),
			PrivateKey:                 types.StringNull(),
			ClientEmail:                types.StringNull(),
			ClientID:                   types.StringNull(),
			AuthURI:                    types.StringNull(),
			TokenURI:                   types.StringNull(),
			AuthProviderX509CertURL:    types.StringNull(),
			ClientX509CertURL:          types.StringNull(),
			Retries:                    types.Int64Null(),
			Scopes:                     nil,
			Priority:                   types.StringNull(),
			Location:                   types.StringNull(),
			MaximumBytesBilled:         types.Int64Null(),
			ExecutionProject:           types.StringNull(),
			ImpersonateServiceAccount:  types.StringNull(),
			JobRetryDeadlineSeconds:    types.Int64Null(),
			JobCreationTimeoutSeconds:  types.Int64Null(),
			ApplicationID:              types.StringNull(),
			ApplicationSecret:          types.StringNull(),
			GcsBucket:                  types.StringNull(),
			DataprocRegion:             types.StringNull(),
			DataprocClusterName:        types.StringNull(),
			UseLatestAdapter:           types.BoolValue(useLatestAdapter),
			JobExecutionTimeoutSeconds: types.Int64Null(),
			DeploymentEnvAuthType:      types.StringValue("service-account-json"),
		},
	}
}

// TestReadGeneric_BigQueryV1_SetsUseLatestAdapter reproduces the core of #709:
// before the fix, readGeneric (shared by Read and ImportState) never set
// BigQueryConfig.UseLatestAdapter, so it stayed null after import/refresh.
func TestReadGeneric_BigQueryV1_SetsUseLatestAdapter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": {"code": 200, "is_success": true},
			"data": {
				"id": 123,
				"name": "bq-v1-conn",
				"account_id": 1,
				"adapter_version": "bigquery_v1",
				"private_link_endpoint_id": null,
				"oauth_configuration_id": null,
				"config": {
					"gcp_project_id": "my-project",
					"priority": null,
					"location": null,
					"maximum_bytes_billed": null,
					"execution_project": null,
					"impersonate_service_account": null,
					"job_retry_deadline_seconds": null,
					"job_creation_timeout_seconds": null,
					"gcs_bucket": null,
					"dataproc_region": null,
					"dataproc_cluster_name": null
				}
			}
		}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	state := &GlobalConnectionResourceModel{
		ID:             types.Int64Value(123),
		BigQueryConfig: &BigQueryConfig{},
	}

	newState, action, err := readGeneric(client, state, "")
	if err != nil {
		t.Fatalf("readGeneric returned an unexpected error: %v", err)
	}
	if action != "" {
		t.Fatalf("expected no removal action, got %q", action)
	}
	if newState.BigQueryConfig.UseLatestAdapter.IsNull() {
		t.Fatalf("regression: use_latest_adapter is still null after reading a bigquery_v1 connection")
	}
	if !newState.BigQueryConfig.UseLatestAdapter.ValueBool() {
		t.Fatalf("expected use_latest_adapter=true for a bigquery_v1 connection, got false")
	}
}

// TestReadGeneric_BigQueryV0_SetsUseLatestAdapterFalse confirms the derivation
// discriminates correctly rather than always reporting true.
func TestReadGeneric_BigQueryV0_SetsUseLatestAdapterFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": {"code": 200, "is_success": true},
			"data": {
				"id": 456,
				"name": "bq-v0-conn",
				"account_id": 1,
				"adapter_version": "bigquery_v0",
				"private_link_endpoint_id": null,
				"oauth_configuration_id": null,
				"config": {
					"gcp_project_id": "my-project",
					"timeout_seconds": 300,
					"priority": null,
					"location": null,
					"maximum_bytes_billed": null,
					"execution_project": null,
					"impersonate_service_account": null,
					"job_retry_deadline_seconds": null,
					"job_creation_timeout_seconds": null,
					"gcs_bucket": null,
					"dataproc_region": null,
					"dataproc_cluster_name": null
				}
			}
		}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	state := &GlobalConnectionResourceModel{
		ID:             types.Int64Value(456),
		BigQueryConfig: &BigQueryConfig{},
	}

	newState, _, err := readGeneric(client, state, "")
	if err != nil {
		t.Fatalf("readGeneric returned an unexpected error: %v", err)
	}
	if newState.BigQueryConfig.UseLatestAdapter.IsNull() || newState.BigQueryConfig.UseLatestAdapter.ValueBool() {
		t.Fatalf("expected use_latest_adapter=false for a bigquery_v0 connection")
	}
}

// TestModifyPlan_BigQueryV1_NoFalsePositiveWhenAdapterUnchanged reproduces the
// user-facing symptom of #709: a plain no-op plan on an imported bigquery_v1
// connection used to fail with "Adapter version cannot be changed" because
// state.BigQueryConfig.UseLatestAdapter was null. With the read-path fix,
// state now correctly says true, so the guard must not fire.
func TestModifyPlan_BigQueryV1_NoFalsePositiveWhenAdapterUnchanged(t *testing.T) {
	schemaResp := testResourceSchema(t)

	plan := bigQueryModel(true, 300)
	state := bigQueryModel(true, 300) // what readGeneric now produces after the fix

	tfPlan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := tfPlan.Set(context.Background(), plan); diags.HasError() {
		t.Fatalf("failed to build plan fixture: %v", diags)
	}
	tfState := tfsdk.State{Schema: schemaResp.Schema}
	if diags := tfState.Set(context.Background(), state); diags.HasError() {
		t.Fatalf("failed to build state fixture: %v", diags)
	}

	r := globalConnectionResource{}
	var resp resource.ModifyPlanResponse
	r.ModifyPlan(
		context.Background(),
		resource.ModifyPlanRequest{Plan: tfPlan, State: tfState},
		&resp,
	)

	if resp.Diagnostics.HasError() {
		t.Fatalf("ModifyPlan unexpectedly errored on a no-op plan (this is the #709 regression): %v", resp.Diagnostics)
	}
}

// TestModifyPlan_BigQueryV1_StillBlocksGenuineAdapterChange makes sure the fix
// didn't neuter the guard: a real v0 -> v1 change must still be rejected.
func TestModifyPlan_BigQueryV1_StillBlocksGenuineAdapterChange(t *testing.T) {
	schemaResp := testResourceSchema(t)

	plan := bigQueryModel(true, 300)
	state := bigQueryModel(false, 300)

	tfPlan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := tfPlan.Set(context.Background(), plan); diags.HasError() {
		t.Fatalf("failed to build plan fixture: %v", diags)
	}
	tfState := tfsdk.State{Schema: schemaResp.Schema}
	if diags := tfState.Set(context.Background(), state); diags.HasError() {
		t.Fatalf("failed to build state fixture: %v", diags)
	}

	r := globalConnectionResource{}
	var resp resource.ModifyPlanResponse
	r.ModifyPlan(
		context.Background(),
		resource.ModifyPlanRequest{Plan: tfPlan, State: tfState},
		&resp,
	)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected ModifyPlan to still block a genuine adapter version change")
	}
}

func captureUpdatePatchBody(t *testing.T, capture *map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read request body: %v", err)
			}
			if err := json.Unmarshal(body, capture); err != nil {
				t.Fatalf("failed to unmarshal request body: %v", err)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": {"code": 200, "is_success": true},
			"data": {"id": 123, "account_id": 1, "config": {"gcp_project_id": "my-project"}}
		}`))
	}))
}

// TestUpdate_BigQueryV1_DoesNotSendTimeoutSecondsToAPI reproduces the second
// half of #709: Update used to send timeout_seconds unconditionally whenever
// it differed from state, even though the field is bigquery_v0-only and the
// API rejects it for bigquery_v1.
func TestUpdate_BigQueryV1_DoesNotSendTimeoutSecondsToAPI(t *testing.T) {
	var captured map[string]interface{}
	server := captureUpdatePatchBody(t, &captured)
	defer server.Close()

	schemaResp := testResourceSchema(t)
	client := newTestClient(t, server)

	// state carries a stale timeout_seconds value (e.g. left over before the
	// connection was switched to the v1 adapter), plan has a different value -
	// this is exactly the condition that used to trigger the send.
	state := bigQueryModel(true, 600)
	plan := bigQueryModel(true, 300)

	tfPlan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := tfPlan.Set(context.Background(), plan); diags.HasError() {
		t.Fatalf("failed to build plan fixture: %v", diags)
	}
	tfState := tfsdk.State{Schema: schemaResp.Schema}
	if diags := tfState.Set(context.Background(), state); diags.HasError() {
		t.Fatalf("failed to build state fixture: %v", diags)
	}

	r := &globalConnectionResource{client: client}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Update(
		context.Background(),
		resource.UpdateRequest{Plan: tfPlan, State: tfState},
		resp,
	)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Update returned an unexpected error: %v", resp.Diagnostics)
	}
	if captured == nil {
		t.Fatalf("expected a PATCH request to have been sent")
	}
	config, _ := captured["config"].(map[string]interface{})
	if _, present := config["timeout_seconds"]; present {
		t.Fatalf("regression: timeout_seconds was sent to the API for a bigquery_v1 connection: %v", config)
	}
}

// TestUpdate_BigQueryV0_StillSendsTimeoutSecondsToAPI confirms the new guard
// only excludes timeout_seconds for v1, and doesn't break v0 behavior.
func TestUpdate_BigQueryV0_StillSendsTimeoutSecondsToAPI(t *testing.T) {
	var captured map[string]interface{}
	server := captureUpdatePatchBody(t, &captured)
	defer server.Close()

	schemaResp := testResourceSchema(t)
	client := newTestClient(t, server)

	state := bigQueryModel(false, 600)
	plan := bigQueryModel(false, 300)

	tfPlan := tfsdk.Plan{Schema: schemaResp.Schema}
	if diags := tfPlan.Set(context.Background(), plan); diags.HasError() {
		t.Fatalf("failed to build plan fixture: %v", diags)
	}
	tfState := tfsdk.State{Schema: schemaResp.Schema}
	if diags := tfState.Set(context.Background(), state); diags.HasError() {
		t.Fatalf("failed to build state fixture: %v", diags)
	}

	r := &globalConnectionResource{client: client}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	r.Update(
		context.Background(),
		resource.UpdateRequest{Plan: tfPlan, State: tfState},
		resp,
	)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Update returned an unexpected error: %v", resp.Diagnostics)
	}
	if captured == nil {
		t.Fatalf("expected a PATCH request to have been sent")
	}
	config, _ := captured["config"].(map[string]interface{})
	got, present := config["timeout_seconds"]
	if !present {
		t.Fatalf("expected timeout_seconds to still be sent for a bigquery_v0 connection")
	}
	if got != float64(300) {
		t.Fatalf("expected timeout_seconds=300, got %v", got)
	}
}
