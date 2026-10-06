package webhook_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/webhook"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	testAccountID = int64(999)
	testWebhookID = "wsu_123"
)

// newWebhookTestClient builds a real *dbt_cloud.Client pointed at the given httptest
// server, skipping the credentials-validation round trip NewClient otherwise performs.
func newWebhookTestClient(t *testing.T, serverURL string) *dbt_cloud.Client {
	t.Helper()
	accountID := testAccountID
	token := "test-token"
	maxRetries := 0
	retryInterval := 0
	timeout := 5
	client, err := dbt_cloud.NewClient(
		&accountID, &token, &serverURL, &maxRetries, &retryInterval, nil, true, &timeout,
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

// reorderingWebhookAPI fakes the dbt Cloud webhook endpoints. The API does not return
// job_ids in the order they were sent, so this fake stores them reversed.
type reorderingWebhookAPI struct {
	storedJobIDs []int64
	sentJobIDs   []int64
}

func (a *reorderingWebhookAPI) serve(t *testing.T) *httptest.Server {
	t.Helper()
	createPath := fmt.Sprintf("/v3/accounts/%d/webhooks/subscriptions", testAccountID)
	webhookPath := fmt.Sprintf("/v3/accounts/%d/webhooks/subscription/%s", testAccountID, testWebhookID)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == createPath,
			r.Method == http.MethodPut && r.URL.Path == webhookPath:
			var body dbt_cloud.WebhookWrite
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			a.sentJobIDs = body.JobIds
			a.storedJobIDs = make([]int64, len(body.JobIds))
			for i, id := range body.JobIds {
				a.storedJobIDs[len(body.JobIds)-1-i] = id
			}
		case r.Method == http.MethodGet && r.URL.Path == webhookPath:
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		jobIDs := make([]string, len(a.storedJobIDs))
		for i, id := range a.storedJobIDs {
			jobIDs[i] = strconv.FormatInt(id, 10)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"id":                 testWebhookID,
				"name":               "test",
				"client_url":         "https://example.com",
				"event_types":        []string{"job.run.completed"},
				"job_ids":            jobIDs,
				"active":             true,
				"hmac_secret":        "secret",
				"account_identifier": "act_123",
			},
			"status": map[string]any{"code": 200, "is_success": true},
		}); err != nil {
			t.Fatalf("encoding response: %v", err)
		}
	}))
}

func configuredWebhookResource(t *testing.T, serverURL string) resource.Resource {
	t.Helper()
	r := webhook.WebhookResource()
	r.(resource.ResourceWithConfigure).Configure(
		context.Background(),
		resource.ConfigureRequest{ProviderData: newWebhookTestClient(t, serverURL)},
		&resource.ConfigureResponse{},
	)
	return r
}

func webhookSchema(t *testing.T) resource_schema.Schema {
	t.Helper()
	resp := &resource.SchemaResponse{}
	webhook.WebhookResource().Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema: %v", resp.Diagnostics.Errors())
	}
	return resp.Schema
}

func jobIDSet(ids ...int64) types.Set {
	values := make([]attr.Value, len(ids))
	for i, id := range ids {
		values[i] = types.Int64Value(id)
	}
	return types.SetValueMust(types.Int64Type, values)
}

func existingWebhookModel(jobIDs types.Set) webhook.WebhookResourceModel {
	return webhook.WebhookResourceModel{
		ID:                types.StringValue(testWebhookID),
		WebhookID:         types.StringValue(testWebhookID),
		Name:              types.StringValue("test"),
		Description:       types.StringValue(""),
		ClientURL:         types.StringValue("https://example.com"),
		EventTypes:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("job.run.completed")}),
		JobIDs:            jobIDs,
		Active:            types.BoolValue(true),
		HmacSecret:        types.StringValue("secret"),
		HTTPStatusCode:    types.StringNull(),
		AccountIdentifier: types.StringValue("act_123"),
	}
}

func assertSentJobIDs(t *testing.T, api *reorderingWebhookAPI, want types.Set) {
	t.Helper()
	sent := make([]attr.Value, len(api.sentJobIDs))
	for i, id := range api.sentJobIDs {
		sent[i] = types.Int64Value(id)
	}
	if got := types.SetValueMust(types.Int64Type, sent); !got.Equal(want) {
		t.Errorf("job IDs sent to the API: got %s, want %s", got, want)
	}
}

func TestWebhookResource_CreateWithReorderedJobIDs(t *testing.T) {
	ctx := context.Background()
	api := &reorderingWebhookAPI{}
	srv := api.serve(t)
	defer srv.Close()
	r := configuredWebhookResource(t, srv.URL)
	schema := webhookSchema(t)

	planned := existingWebhookModel(jobIDSet(1, 2, 3))
	planned.ID = types.StringUnknown()
	planned.WebhookID = types.StringUnknown()
	planned.HmacSecret = types.StringUnknown()
	planned.HTTPStatusCode = types.StringUnknown()
	planned.AccountIdentifier = types.StringUnknown()
	plan := tfsdk.Plan{Schema: schema}
	if diags := plan.Set(ctx, &planned); diags.HasError() {
		t.Fatalf("building plan: %v", diags.Errors())
	}

	resp := &resource.CreateResponse{State: tfsdk.State{Schema: schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Create: %v", resp.Diagnostics.Errors())
	}

	var got webhook.WebhookResourceModel
	if diags := resp.State.Get(ctx, &got); diags.HasError() {
		t.Fatalf("reading state: %v", diags.Errors())
	}
	if !got.JobIDs.Equal(planned.JobIDs) {
		t.Errorf("job_ids in state: got %s, want %s", got.JobIDs, planned.JobIDs)
	}
	assertSentJobIDs(t, api, planned.JobIDs)
}

func TestWebhookResource_UpdateWithReorderedJobIDs(t *testing.T) {
	ctx := context.Background()
	api := &reorderingWebhookAPI{storedJobIDs: []int64{2, 1}}
	srv := api.serve(t)
	defer srv.Close()
	r := configuredWebhookResource(t, srv.URL)
	schema := webhookSchema(t)

	prior := existingWebhookModel(jobIDSet(1, 2))
	planned := existingWebhookModel(jobIDSet(1, 2, 3))
	state := tfsdk.State{Schema: schema}
	if diags := state.Set(ctx, &prior); diags.HasError() {
		t.Fatalf("building state: %v", diags.Errors())
	}
	plan := tfsdk.Plan{Schema: schema}
	if diags := plan.Set(ctx, &planned); diags.HasError() {
		t.Fatalf("building plan: %v", diags.Errors())
	}

	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: plan, State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Update: %v", resp.Diagnostics.Errors())
	}

	var got webhook.WebhookResourceModel
	if diags := resp.State.Get(ctx, &got); diags.HasError() {
		t.Fatalf("reading state: %v", diags.Errors())
	}
	if !got.JobIDs.Equal(planned.JobIDs) {
		t.Errorf("job_ids in state: got %s, want %s", got.JobIDs, planned.JobIDs)
	}
	assertSentJobIDs(t, api, planned.JobIDs)
}

// State saved by provider versions that declared job_ids as a list must load into
// the set schema without a state upgrader.
func TestWebhookResource_UpgradeStateFromList(t *testing.T) {
	ctx := context.Background()
	schema := webhookSchema(t)

	server := providerserver.NewProtocol6(provider.New())()
	resp, err := server.UpgradeResourceState(ctx, &tfprotov6.UpgradeResourceStateRequest{
		TypeName: "dbtcloud_webhook",
		Version:  schema.GetVersion(),
		RawState: &tfprotov6.RawState{
			JSON: []byte(`{
				"id": "wsu_123",
				"webhook_id": "wsu_123",
				"name": "test",
				"description": "",
				"client_url": "https://example.com",
				"event_types": ["job.run.completed"],
				"job_ids": [3, 1, 2],
				"active": true,
				"hmac_secret": "secret",
				"http_status_code": null,
				"account_identifier": "act_123"
			}`),
		},
	})
	if err != nil {
		t.Fatalf("UpgradeResourceState: %v", err)
	}
	for _, d := range resp.Diagnostics {
		t.Errorf("UpgradeResourceState diagnostic: %s: %s", d.Summary, d.Detail)
	}
	if t.Failed() {
		return
	}

	upgraded, err := resp.UpgradedState.Unmarshal(schema.Type().TerraformType(ctx))
	if err != nil {
		t.Fatalf("decoding upgraded state: %v", err)
	}
	var got webhook.WebhookResourceModel
	if diags := (tfsdk.State{Schema: schema, Raw: upgraded}).Get(ctx, &got); diags.HasError() {
		t.Fatalf("reading upgraded state: %v", diags.Errors())
	}
	if want := jobIDSet(1, 2, 3); !got.JobIDs.Equal(want) {
		t.Errorf("job_ids after upgrade: got %s, want %s", got.JobIDs, want)
	}
}
