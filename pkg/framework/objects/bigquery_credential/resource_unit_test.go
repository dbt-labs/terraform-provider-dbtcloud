package bigquery_credential_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	bqcredential "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/bigquery_credential"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// newBigQueryTestClient builds a real *dbt_cloud.Client pointed at the given httptest
// server, skipping the credentials-validation round trip NewClient otherwise performs.
func newBigQueryTestClient(t *testing.T, serverURL string, accountID int64) *dbt_cloud.Client {
	t.Helper()
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

func credentialAttributeTypes() map[string]tftypes.Type {
	return map[string]tftypes.Type{
		"id":                                tftypes.String,
		"is_active":                         tftypes.Bool,
		"project_id":                        tftypes.Number,
		"credential_id":                     tftypes.Number,
		"dataset":                           tftypes.String,
		"num_threads":                       tftypes.Number,
		"connection_id":                     tftypes.Number,
		"auth_type":                         tftypes.String,
		"workload_pool_provider_path":       tftypes.String,
		"service_account_impersonation_url": tftypes.String,
	}
}

func credentialConfig(connectionID, authType, workloadPoolProviderPath tftypes.Value) tfsdk.Config {
	raw := tftypes.NewValue(
		tftypes.Object{AttributeTypes: credentialAttributeTypes()},
		map[string]tftypes.Value{
			"id":                                tftypes.NewValue(tftypes.String, nil),
			"is_active":                         tftypes.NewValue(tftypes.Bool, true),
			"project_id":                        tftypes.NewValue(tftypes.Number, 123),
			"credential_id":                     tftypes.NewValue(tftypes.Number, nil),
			"dataset":                           tftypes.NewValue(tftypes.String, "my_dataset"),
			"num_threads":                       tftypes.NewValue(tftypes.Number, 4),
			"connection_id":                     connectionID,
			"auth_type":                         authType,
			"workload_pool_provider_path":       workloadPoolProviderPath,
			"service_account_impersonation_url": tftypes.NewValue(tftypes.String, nil),
		},
	)
	return tfsdk.Config{Schema: bqcredential.BigQueryResourceSchema, Raw: raw}
}

func runCredentialValidateConfig(t *testing.T, cfg tfsdk.Config) *resource.ValidateConfigResponse {
	t.Helper()
	r, ok := bqcredential.BigqueryCredentialResource().(resource.ResourceWithValidateConfig)
	if !ok {
		t.Fatal("bigquery credential resource does not implement ResourceWithValidateConfig")
	}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: cfg}, resp)
	return resp
}

func runCredentialAttributeValidators(
	t *testing.T,
	cfg tfsdk.Config,
	name string,
	value string,
) *validator.StringResponse {
	t.Helper()
	attribute, ok := bqcredential.BigQueryResourceSchema.Attributes[name].(resource_schema.StringAttribute)
	if !ok {
		t.Fatalf("%s is not a string attribute", name)
	}

	req := validator.StringRequest{
		Path:           path.Root(name),
		PathExpression: path.MatchRoot(name),
		Config:         cfg,
		ConfigValue:    types.StringValue(value),
	}
	resp := &validator.StringResponse{}
	for _, v := range attribute.Validators {
		v.ValidateString(context.Background(), req, resp)
	}
	return resp
}

func TestBigQueryCredentialValidateConfig_WIFWithoutWorkloadPoolProviderPath(t *testing.T) {
	t.Parallel()
	resp := runCredentialValidateConfig(t, credentialConfig(
		tftypes.NewValue(tftypes.Number, 456),
		tftypes.NewValue(tftypes.String, "external-oauth-wif"),
		tftypes.NewValue(tftypes.String, nil),
	))
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for missing workload_pool_provider_path with external-oauth-wif")
	}
}

func TestBigQueryCredentialValidateConfig_WIFWithWorkloadPoolProviderPath(t *testing.T) {
	t.Parallel()
	resp := runCredentialValidateConfig(t, credentialConfig(
		tftypes.NewValue(tftypes.Number, 456),
		tftypes.NewValue(tftypes.String, "external-oauth-wif"),
		tftypes.NewValue(tftypes.String, "projects/1/locations/global/workloadIdentityPools/p/providers/pr"),
	))
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no errors, got: %s", resp.Diagnostics.Errors())
	}
}

func TestBigQueryCredentialValidateConfig_UnknownWorkloadPoolProviderPath(t *testing.T) {
	t.Parallel()
	resp := runCredentialValidateConfig(t, credentialConfig(
		tftypes.NewValue(tftypes.Number, 456),
		tftypes.NewValue(tftypes.String, "external-oauth-wif"),
		tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	))
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected unknown values to be deferred to apply, got: %s", resp.Diagnostics.Errors())
	}
}

func TestBigQueryCredentialValidateConfig_ServiceAccountJSONWithoutWIFFields(t *testing.T) {
	t.Parallel()
	resp := runCredentialValidateConfig(t, credentialConfig(
		tftypes.NewValue(tftypes.Number, 456),
		tftypes.NewValue(tftypes.String, "service-account-json"),
		tftypes.NewValue(tftypes.String, nil),
	))
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no errors, got: %s", resp.Diagnostics.Errors())
	}
}

func TestBigQueryCredentialWIFAttributes_RequireConnectionID(t *testing.T) {
	t.Parallel()

	withoutConnection := credentialConfig(
		tftypes.NewValue(tftypes.Number, nil),
		tftypes.NewValue(tftypes.String, "external-oauth-wif"),
		tftypes.NewValue(tftypes.String, "projects/1/locations/global/workloadIdentityPools/p/providers/pr"),
	)
	withConnection := credentialConfig(
		tftypes.NewValue(tftypes.Number, 456),
		tftypes.NewValue(tftypes.String, "external-oauth-wif"),
		tftypes.NewValue(tftypes.String, "projects/1/locations/global/workloadIdentityPools/p/providers/pr"),
	)

	for name, value := range map[string]string{
		"auth_type":                         "external-oauth-wif",
		"workload_pool_provider_path":       "projects/1/locations/global/workloadIdentityPools/p/providers/pr",
		"service_account_impersonation_url": "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/svc:generateAccessToken",
	} {
		if resp := runCredentialAttributeValidators(t, withoutConnection, name, value); !resp.Diagnostics.HasError() {
			t.Errorf("expected error for %s without connection_id", name)
		}
		if resp := runCredentialAttributeValidators(t, withConnection, name, value); resp.Diagnostics.HasError() {
			t.Errorf("expected no errors for %s with connection_id, got: %s", name, resp.Diagnostics.Errors())
		}
	}
}

// The Semantic Layer credential resource reuses these attributes and its API does not
// accept the v1 WIF fields, so they must not leak into its schema.
func TestBigQueryCredentialSemanticLayerAttributes_ExcludeWIFFields(t *testing.T) {
	t.Parallel()
	for _, name := range []string{
		"auth_type",
		"workload_pool_provider_path",
		"service_account_impersonation_url",
	} {
		if _, ok := bqcredential.SemanticLayerAttributes[name]; ok {
			t.Errorf("%s should not be part of the Semantic Layer credential attributes", name)
		}
		if _, ok := bqcredential.BigQueryResourceSchema.Attributes[name]; !ok {
			t.Errorf("%s should be part of the BigQuery credential resource schema", name)
		}
	}
}

// TestBigQueryCredentialUpdate_SendsWIFFieldChanges is a regression test for Update() only
// ever pushing dataset/num_threads to the API: changing auth_type,
// workload_pool_provider_path, or service_account_impersonation_url must actually reach
// UpdateBigQueryCredential, not just get written into Terraform state.
func TestBigQueryCredentialUpdate_SendsWIFFieldChanges(t *testing.T) {
	t.Parallel()

	const (
		accountID    = int64(999)
		projectID    = 1
		credentialID = 111
	)

	wantPath := fmt.Sprintf("/v3/accounts/%d/projects/%d/credentials/%d/", accountID, projectID, credentialID)
	var updateCalls int
	var lastUpdateBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("unexpected request path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			fmt.Fprintf(w, `{
				"data": {
					"id": %d, "account_id": %d, "project_id": %d,
					"type": "adapter", "state": 1, "threads": 4, "schema": "",
					"adapter_version": "bigquery_v1"
				},
				"status": {"code": 200, "is_success": true}
			}`, credentialID, accountID, projectID)
		case http.MethodPost:
			updateCalls++
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("reading update request body: %v", err)
			}
			if err := json.Unmarshal(body, &lastUpdateBody); err != nil {
				t.Fatalf("unmarshalling update request body: %v. body: %s", err, body)
			}
			fmt.Fprintf(w, `{
				"data": {
					"id": %d, "account_id": %d, "project_id": %d,
					"type": "adapter", "state": 1, "threads": 4,
					"adapter_version": "bigquery_v1"
				},
				"status": {"code": 200, "is_success": true}
			}`, credentialID, accountID, projectID)
		default:
			t.Errorf("unexpected method: %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer srv.Close()

	client := newBigQueryTestClient(t, srv.URL, accountID)

	credResource := bqcredential.BigqueryCredentialResource()
	configurable, ok := credResource.(resource.ResourceWithConfigure)
	if !ok {
		t.Fatal("bigquery credential resource does not implement ResourceWithConfigure")
	}
	configurable.Configure(
		context.Background(),
		resource.ConfigureRequest{ProviderData: client},
		&resource.ConfigureResponse{},
	)

	buildValue := func(dataset string, numThreads, connectionID int, authType, workloadPoolProviderPath, serviceAccountImpersonationURL tftypes.Value) tftypes.Value {
		return tftypes.NewValue(
			tftypes.Object{AttributeTypes: credentialAttributeTypes()},
			map[string]tftypes.Value{
				"id":                                tftypes.NewValue(tftypes.String, fmt.Sprintf("%d:%d", projectID, credentialID)),
				"is_active":                         tftypes.NewValue(tftypes.Bool, true),
				"project_id":                        tftypes.NewValue(tftypes.Number, projectID),
				"credential_id":                     tftypes.NewValue(tftypes.Number, credentialID),
				"dataset":                           tftypes.NewValue(tftypes.String, dataset),
				"num_threads":                       tftypes.NewValue(tftypes.Number, numThreads),
				"connection_id":                     tftypes.NewValue(tftypes.Number, connectionID),
				"auth_type":                         authType,
				"workload_pool_provider_path":       workloadPoolProviderPath,
				"service_account_impersonation_url": serviceAccountImpersonationURL,
			},
		)
	}

	const (
		newWorkloadPoolProviderPath       = "projects/1/locations/global/workloadIdentityPools/p/providers/pr"
		newServiceAccountImpersonationURL = "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/svc:generateAccessToken"
	)

	stateRaw := buildValue(
		"my_dataset", 4, 456,
		tftypes.NewValue(tftypes.String, "service-account-json"),
		tftypes.NewValue(tftypes.String, nil),
		tftypes.NewValue(tftypes.String, nil),
	)
	// Only the WIF fields change; dataset/num_threads stay the same, so before the fix
	// this update would be a no-op against the API.
	planRaw := buildValue(
		"my_dataset", 4, 456,
		tftypes.NewValue(tftypes.String, "external-oauth-wif"),
		tftypes.NewValue(tftypes.String, newWorkloadPoolProviderPath),
		tftypes.NewValue(tftypes.String, newServiceAccountImpersonationURL),
	)

	updateResp := &resource.UpdateResponse{
		State: tfsdk.State{Schema: bqcredential.BigQueryResourceSchema},
	}
	credResource.Update(context.Background(), resource.UpdateRequest{
		Plan:  tfsdk.Plan{Schema: bqcredential.BigQueryResourceSchema, Raw: planRaw},
		State: tfsdk.State{Schema: bqcredential.BigQueryResourceSchema, Raw: stateRaw},
	}, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %s", updateResp.Diagnostics.Errors())
	}
	if updateCalls != 1 {
		t.Fatalf("expected exactly 1 UpdateBigQueryCredential call, got %d", updateCalls)
	}

	details, ok := lastUpdateBody["credential_details"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected credential_details in update request body, got: %v", lastUpdateBody)
	}
	fields, ok := details["fields"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected credential_details.fields in update request body, got: %v", details)
	}

	assertFieldValue := func(name, want string) {
		field, ok := fields[name].(map[string]interface{})
		if !ok {
			t.Fatalf("expected field %q in credential_details.fields, got: %v", name, fields)
		}
		got, _ := field["value"].(string)
		if got != want {
			t.Errorf("credential_details.fields[%q].value = %q, want %q", name, got, want)
		}
	}

	assertFieldValue("auth_type", "external-oauth-wif")
	assertFieldValue("workload_pool_provider_path", newWorkloadPoolProviderPath)
	assertFieldValue("service_account_impersonation_url", newServiceAccountImpersonationURL)
}
