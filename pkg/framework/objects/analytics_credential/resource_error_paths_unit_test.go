package analytics_credential_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/analytics_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func errorResponse(code int) testhelpers.MockEndpointHandler {
	return func(r *http.Request) (int, interface{}, error) {
		return code, map[string]interface{}{
			"status": map[string]interface{}{"code": code, "is_success": false, "user_message": "failed"},
		}, nil
	}
}

func hasErrorWith(diags diag.Diagnostics, summary, detail string) bool {
	for _, d := range diags.Errors() {
		if strings.Contains(d.Summary(), summary) && strings.Contains(d.Detail(), detail) {
			return true
		}
	}
	return false
}

func planAsConfig(plan tfsdk.Plan) tfsdk.Config {
	return tfsdk.Config{Schema: plan.Schema, Raw: plan.Raw}
}

// --- auth switches may drop optional attributes ------------------------------------------

// Rule: switching the authentication type replaces the credential, so the optional attributes of the old type may be dropped.
// given: a Snowflake credential with a role, switched to keypair authentication without a role
// when: the configuration is planned
// then: the credential is replaced and no "cannot be removed" error is raised.
func TestAnalyticsCredential_AuthSwitchMayDropOptionalSnowflakeField(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "snowflake_v0"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	withRole := snowflakeModel("password")
	withRole.Role = types.StringValue("ANALYST")

	resp := runModifyPlan(t, r,
		buildPlan(t, s, credentialModel(42, snowflakeModel("keypair"))),
		buildState(t, s, credentialModel(42, withRole)))

	if hasErrorSummary(resp, "Optional attribute cannot be removed") {
		t.Fatalf("a replacement must not raise the clearing error, got %v", resp.Diagnostics)
	}
	if len(resp.RequiresReplace) == 0 {
		t.Fatal("expected the auth type change to require a replacement")
	}
}

// Rule: switching a BigQuery credential from workload identity federation to a service account replaces it, so the impersonation URL may be dropped.
// given: a workload identity credential with an impersonation URL, switched to a service account without it
// when: the configuration is planned
// then: the credential is replaced and no "cannot be removed" error is raised.
func TestAnalyticsCredential_AuthSwitchMayDropBigQueryImpersonationURL(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "bigquery_v1"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	wif := bigqueryCredentialModel(42, "bigquery_v1", strPtr("external-oauth-wif"))
	wif.BigQuery.ServiceAccountImpersonationURL = types.StringValue("https://iamcredentials.googleapis.com/x")

	resp := runModifyPlan(t, r,
		buildPlan(t, s, bigqueryCredentialModel(42, "bigquery_v1", nil)),
		buildState(t, s, wif))

	if hasErrorSummary(resp, "Optional attribute cannot be removed") {
		t.Fatalf("a replacement must not raise the clearing error, got %v", resp.Diagnostics)
	}
	if len(resp.RequiresReplace) == 0 {
		t.Fatal("expected the auth type change to require a replacement")
	}
}

// Rule: moving a credential to another connection replaces it, so optional attributes may be dropped.
// given: a Snowflake credential with a role, moved to another Snowflake connection without a role
// when: the configuration is planned
// then: no "cannot be removed" error is raised.
func TestAnalyticsCredential_ConnectionChangeMayDropOptionalField(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "snowflake_v0"),
		"GET " + connectionPath(43): connectionHandler(43, "snowflake_v0"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	withRole := snowflakeModel("password")
	withRole.Role = types.StringValue("ANALYST")

	resp := runModifyPlan(t, r,
		buildPlan(t, s, credentialModel(43, snowflakeModel("password"))),
		buildState(t, s, credentialModel(42, withRole)))

	if hasErrorSummary(resp, "Optional attribute cannot be removed") {
		t.Fatalf("a replacement must not raise the clearing error, got %v", resp.Diagnostics)
	}
}

// --- error paths -------------------------------------------------------------------------

// Rule: a credential that no longer exists is removed from the state on refresh.
// given: the API answers 404 for the credential
// when: the resource is read
// then: it is removed from the state without an error.
func TestAnalyticsCredential_ReadNotFoundRemovesFromState(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + itemPath(credentialID): errorResponse(http.StatusNotFound),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	state := buildState(t, s, credentialModel(42, snowflakeModel("password")))
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatal("expected the resource to be removed from the state")
	}
}

// Rule: a credential that is already gone can be destroyed.
// given: the API answers 404 when the credential is deleted
// when: the resource is destroyed
// then: no error is raised.
func TestAnalyticsCredential_DeleteNotFoundSucceeds(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"DELETE " + itemPath(credentialID): errorResponse(http.StatusNotFound),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	state := buildState(t, s, credentialModel(42, snowflakeModel("password")))
	resp := &resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
}

// Rule: an unexpected failure when deleting is reported with the credential and connection.
// given: the API answers 500 when the credential is deleted
// when: the resource is destroyed
// then: an error names the credential and its connection.
func TestAnalyticsCredential_DeleteFailureNamesCredential(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"DELETE " + itemPath(credentialID): errorResponse(http.StatusInternalServerError),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	state := buildState(t, s, credentialModel(42, snowflakeModel("password")))
	resp := &resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, resp)

	if !hasErrorWith(resp.Diagnostics, "Error deleting analytics credential", "credential 7 (connection 42)") {
		t.Fatalf("expected an error naming the credential and connection, got %v", resp.Diagnostics)
	}
}

// Rule: a failed create is reported with the connection it was meant for.
// given: the API answers 500 when the credential is created
// when: the resource is created
// then: an error names the connection.
func TestAnalyticsCredential_CreateFailureNamesConnection(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "snowflake_v0"),
		"POST " + collectionPath:    errorResponse(http.StatusInternalServerError),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	plan := buildPlan(t, s, credentialModel(42, snowflakeModel("password")))
	resp := &resource.CreateResponse{State: nullState(s)}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan, Config: planAsConfig(plan)}, resp)

	if !hasErrorWith(resp.Diagnostics, "Error creating analytics credential", "connection 42") {
		t.Fatalf("expected an error naming the connection, got %v", resp.Diagnostics)
	}
}

// Rule: a create response without an id tells the user that a credential may exist.
// given: the API accepts the create but returns no id
// when: the resource is created
// then: the error explains that a credential may exist and how to import it.
func TestAnalyticsCredential_CreateWithoutIDExplainsOrphan(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "snowflake_v0"),
		"POST " + collectionPath: func(r *http.Request) (int, interface{}, error) {
			return http.StatusCreated, map[string]interface{}{
				"status": map[string]interface{}{"code": 201, "is_success": true},
				"data":   map[string]interface{}{"adapter_version": "snowflake_v0"},
			}, nil
		},
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	plan := buildPlan(t, s, credentialModel(42, snowflakeModel("password")))
	resp := &resource.CreateResponse{State: nullState(s)}
	r.Create(context.Background(), resource.CreateRequest{Plan: plan, Config: planAsConfig(plan)}, resp)

	if !hasErrorWith(resp.Diagnostics, "Error creating analytics credential", "may already exist") {
		t.Fatalf("expected the orphan explanation, got %v", resp.Diagnostics)
	}
}

// Rule: updating a credential that no longer exists tells the user how to recover.
// given: the API answers 404 when the credential is updated
// when: the resource is updated
// then: the error names the credential and suggests refreshing or replacing it.
func TestAnalyticsCredential_UpdateNotFoundSuggestsRecovery(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42):       connectionHandler(42, "snowflake_v0"),
		"PATCH " + itemPath(credentialID): errorResponse(http.StatusNotFound),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	plan := buildPlan(t, s, credentialModel(42, snowflakeModel("password")))
	state := buildState(t, s, credentialModel(42, snowflakeModel("password")))
	resp := &resource.UpdateResponse{State: state}
	r.Update(context.Background(), resource.UpdateRequest{Plan: plan, Config: planAsConfig(plan), State: state}, resp)

	if !hasErrorWith(resp.Diagnostics, "Error updating analytics credential", "terraform apply -replace") {
		t.Fatalf("expected a recovery hint, got %v", resp.Diagnostics)
	}
}

// Rule: a credential of an adapter the provider does not support is reported with a way out.
// given: the API returns a credential of a PostgreSQL connection
// when: the resource is read
// then: the error lists the supported adapters and suggests removing it from the state.
func TestAnalyticsCredential_ReadUnsupportedAdapterSuggestsStateRm(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + itemPath(credentialID): func(r *http.Request) (int, interface{}, error) {
			return http.StatusOK, map[string]interface{}{
				"status": map[string]interface{}{"code": 200, "is_success": true},
				"data": map[string]interface{}{
					"id": credentialID, "connection_id": 42, "adapter_version": "postgres_v0", "config": map[string]interface{}{},
				},
			}, nil
		},
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	state := buildState(t, s, credentialModel(42, snowflakeModel("password")))
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, resp)

	if !hasErrorWith(resp.Diagnostics, "Unsupported adapter", "supported: bigquery_v0, bigquery_v1, databricks_v0, redshift_v0, snowflake_v0") {
		t.Fatalf("expected the supported adapters, got %v", resp.Diagnostics)
	}
	if !hasErrorWith(resp.Diagnostics, "Unsupported adapter", "terraform state rm") {
		t.Fatalf("expected the state rm hint, got %v", resp.Diagnostics)
	}
}

// --- import ids --------------------------------------------------------------------------

// Rule: an import id has at most an account and a credential.
// given: import ids with three parts, or with another account than the provider's
// when: the credential is imported
// then: each is rejected with an explanation.
func TestAnalyticsCredential_ImportRejectsMalformedIDs(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	cases := map[string]string{
		"1:2:3": "Expected `credential_id` or `account_id:credential_id`",
		"5:7":   "which is 1",
	}
	for id, detail := range cases {
		t.Run(id, func(t *testing.T) {
			resp := &resource.ImportStateResponse{State: nullState(s)}
			r.(resource.ResourceWithImportState).ImportState(
				context.Background(), resource.ImportStateRequest{ID: id}, resp,
			)
			if !hasErrorWith(resp.Diagnostics, "Invalid Import ID", detail) {
				t.Fatalf("expected an error containing %q, got %v", detail, resp.Diagnostics)
			}
		})
	}
}

// Rule: the error for a field that cannot be cleared names the credential and its connection.
// given: a Snowflake credential with a role in the state
// when: the role is removed from the configuration and the plan is run
// then: the error names the credential id and the connection id.
func TestAnalyticsCredential_RemovalErrorNamesCredentialAndConnection(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "snowflake_v0"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	withRole := snowflakeModel("password")
	withRole.Role = types.StringValue("ANALYST")

	resp := runModifyPlan(t, r,
		buildPlan(t, s, credentialModel(42, snowflakeModel("password"))),
		buildState(t, s, credentialModel(42, withRole)))

	if !hasErrorWith(resp.Diagnostics, "Optional attribute cannot be removed", "credential 7") ||
		!hasErrorWith(resp.Diagnostics, "Optional attribute cannot be removed", "connection 42") {
		t.Fatalf("expected the credential and connection ids in the error, got %v", resp.Diagnostics)
	}
	// a forced replacement is planned after the provider rejects the removal, so only a destroy of the existing configuration helps
	if !hasErrorWith(resp.Diagnostics, "Optional attribute cannot be removed", "terraform destroy -target") ||
		hasErrorWith(resp.Diagnostics, "Optional attribute cannot be removed", "-replace") {
		t.Fatalf("expected a destroy hint and no -replace hint, got %v", resp.Diagnostics)
	}
}

// Rule: an imported BigQuery credential keeps the auth type the server returns.
// given: a latest-adapter BigQuery credential with a service account, read with no block in the state (import)
// when: the resource is read
// then: the auth type in the state is the server's, so a configuration that sets it has no diff.
func TestAnalyticsCredential_ImportStoresBigQueryAuthType(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + itemPath(credentialID): func(r *http.Request) (int, interface{}, error) {
			return http.StatusOK, map[string]interface{}{
				"status": map[string]interface{}{"code": 200, "is_success": true},
				"data": map[string]interface{}{
					"id": credentialID, "connection_id": 42, "adapter_version": "bigquery_v1",
					"config": map[string]interface{}{"auth_type": "service-account-json", "execution_project": "billing"},
				},
			}, nil
		},
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	imported := analytics_credential.AnalyticsCredentialResourceModel{
		ID:           types.Int64Value(credentialID),
		ConnectionID: types.Int64Value(42),
	}
	state := buildState(t, s, imported)
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), resource.ReadRequest{State: state}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}

	var got analytics_credential.AnalyticsCredentialResourceModel
	if diags := resp.State.Get(context.Background(), &got); diags.HasError() {
		t.Fatalf("could not decode the state: %v", diags)
	}
	if got.BigQuery == nil || got.BigQuery.AuthType.ValueString() != "service-account-json" {
		t.Fatalf("expected the server auth type after an import, got %+v", got.BigQuery)
	}
}

// Rule: the version of an optional secret cannot be removed while the credential is updated in place.
// given: a keypair credential whose state has a passphrase version
// when: the passphrase version is removed from the configuration and the plan is run
// then: the plan fails naming the credential and its connection, and a replacement raises no such error.
func TestAnalyticsCredential_PassphraseVersionRemovalFailsPlan(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "snowflake_v0"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	withPassphrase := snowflakeModel("keypair")
	withPassphrase.PrivateKeyPassphraseWoVersion = types.Int64Value(1)

	resp := runModifyPlan(t, r,
		buildPlan(t, s, credentialModel(42, snowflakeModel("keypair"))),
		buildState(t, s, credentialModel(42, withPassphrase)))
	if !hasErrorWith(resp.Diagnostics, "Optional attribute cannot be removed", "private_key_passphrase_wo_version") ||
		!hasErrorWith(resp.Diagnostics, "Optional attribute cannot be removed", "credential 7 (connection 42)") {
		t.Fatalf("expected the removal error naming the credential, got %v", resp.Diagnostics)
	}

	// switching to password auth replaces the credential, so the passphrase version may go
	replaced := runModifyPlan(t, r,
		buildPlan(t, s, credentialModel(42, snowflakeModel("password"))),
		buildState(t, s, credentialModel(42, withPassphrase)))
	if hasErrorSummary(replaced, "Optional attribute cannot be removed") {
		t.Fatalf("a replacement must not raise the removal error, got %v", replaced.Diagnostics)
	}
}

// Rule: an auth type that is not known yet plans a replacement, so the final plan cannot change from an update to a replacement.
// given: an existing Snowflake or BigQuery credential, and a plan whose auth type is unknown
// when: the plan is run
// then: a replacement is required for the auth type.
func TestAnalyticsCredential_UnknownAuthTypeRequiresReplace(t *testing.T) {
	cases := map[string]struct {
		adapter string
		state   func() analytics_credential.AnalyticsCredentialResourceModel
		plan    func() analytics_credential.AnalyticsCredentialResourceModel
	}{
		"snowflake": {
			adapter: "snowflake_v0",
			state: func() analytics_credential.AnalyticsCredentialResourceModel {
				return credentialModel(42, snowflakeModel("password"))
			},
			plan: func() analytics_credential.AnalyticsCredentialResourceModel {
				m := credentialModel(42, snowflakeModel("password"))
				m.Snowflake.AuthType = types.StringUnknown()
				return m
			},
		},
		"bigquery": {
			adapter: "bigquery_v1",
			state: func() analytics_credential.AnalyticsCredentialResourceModel {
				return bigqueryCredentialModel(42, "bigquery_v1", nil)
			},
			plan: func() analytics_credential.AnalyticsCredentialResourceModel {
				m := bigqueryCredentialModel(42, "bigquery_v1", nil)
				m.BigQuery.AuthType = types.StringUnknown()
				return m
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
				"GET " + connectionPath(42): connectionHandler(42, tc.adapter),
			})
			defer srv.Close()
			r := newTestResource(t, srv.URL)
			s := resourceSchema(t, r)

			resp := runModifyPlan(t, r, buildPlan(t, s, tc.plan()), buildState(t, s, tc.state()))
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error: %v", resp.Diagnostics)
			}
			if len(resp.RequiresReplace) == 0 {
				t.Fatal("expected an unknown auth type to require a replacement")
			}
		})
	}
}

// Rule: the latest BigQuery adapter has no server side default auth type, so the provider sends the documented one.
// given: a bigquery credential without `auth_type` on a bigquery_v1 connection, one with an explicit WIF type, and one on bigquery_v0
// when: each is created
// then: the request carries service-account-json, external-oauth-wif and no auth type respectively.
func TestAnalyticsCredential_CreateSendsDefaultBigQueryAuthType(t *testing.T) {
	cases := []struct {
		name, adapter string
		authType      *string
		want          string
	}{
		{"v1 without auth_type", "bigquery_v1", nil, "service-account-json"},
		{"v1 explicit wif", "bigquery_v1", strPtr("external-oauth-wif"), "external-oauth-wif"},
		{"v0 has no auth_type", "bigquery_v0", nil, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sent map[string]interface{}
			srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
				"GET " + connectionPath(42): connectionHandler(42, tc.adapter),
				"POST " + collectionPath: func(r *http.Request) (int, interface{}, error) {
					if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
						return http.StatusBadRequest, nil, err
					}
					return http.StatusCreated, map[string]interface{}{
						"status": map[string]interface{}{"code": 201, "is_success": true},
						"data":   map[string]interface{}{"id": 7, "adapter_version": tc.adapter},
					}, nil
				},
			})
			defer srv.Close()
			r := newTestResource(t, srv.URL)
			s := resourceSchema(t, r)

			plan := buildPlan(t, s, bigqueryCredentialModel(42, tc.adapter, tc.authType))
			resp := &resource.CreateResponse{State: nullState(s)}
			r.Create(context.Background(), resource.CreateRequest{Plan: plan, Config: planAsConfig(plan)}, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error: %v", resp.Diagnostics)
			}
			config, _ := sent["config"].(map[string]interface{})
			got, _ := config["auth_type"].(string)
			if got != tc.want {
				t.Fatalf("auth_type in the request = %q, want %q", got, tc.want)
			}
		})
	}
}
