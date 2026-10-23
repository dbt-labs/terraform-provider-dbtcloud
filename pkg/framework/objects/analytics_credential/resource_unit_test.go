package analytics_credential_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/analytics_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const testAccountID = 1

var (
	collectionPath = fmt.Sprintf("/v3/accounts/%d/analytics-credentials/", testAccountID)
	itemPath       = func(id int) string {
		return fmt.Sprintf("/v3/accounts/%d/analytics-credentials/%d/", testAccountID, id)
	}
	connectionPath = func(id int) string {
		return fmt.Sprintf("/v3/accounts/%d/connections/%d/", testAccountID, id)
	}
)

const credentialID = 7

var secretKeys = []string{"password", "private_key", "private_key_passphrase", "token"}

// fakeAPI is a stateful stand-in for the analytics credentials endpoints. Like the real
// API it never stores or returns secrets, unless leakSecrets is set, which makes the GET
// response carry secret values to prove the provider ignores them.
type fakeAPI struct {
	mu          sync.Mutex
	connections map[int]string
	config      map[string]interface{}
	connection  int
	leakSecrets bool
}

func newFakeAPI(connections map[int]string) *fakeAPI {
	return &fakeAPI{connections: connections, config: map[string]interface{}{}}
}

func (f *fakeAPI) credential() map[string]interface{} {
	config := map[string]interface{}{}
	for k, v := range f.config {
		config[k] = v
	}
	// the API returns the nullable fields as explicit nulls
	for _, nullable := range []string{"role", "catalog", "execution_project", "service_account_impersonation_url"} {
		if _, ok := config[nullable]; !ok {
			config[nullable] = nil
		}
	}
	if f.leakSecrets {
		config["password"] = "leaked-password"
		config["private_key"] = "leaked-private-key"
		config["private_key_passphrase"] = "leaked-passphrase"
		config["token"] = "leaked-token"
	}
	return map[string]interface{}{
		"status": map[string]interface{}{"code": 200, "is_success": true},
		"data": map[string]interface{}{
			"id":              credentialID,
			"account_id":      testAccountID,
			"connection_id":   f.connection,
			"adapter_version": f.connections[f.connection],
			"config":          config,
			"created_at":      "2026-10-23T10:00:00Z",
			"updated_at":      "2026-10-23T10:00:00Z",
		},
	}
}

func (f *fakeAPI) store(r *http.Request, replace bool) {
	body := decodeBody(r)
	if connectionID, ok := body["connection_id"].(float64); ok {
		f.connection = int(connectionID)
	}
	config, _ := body["config"].(map[string]interface{})
	if replace {
		f.config = map[string]interface{}{}
		// like the server, a v1 BigQuery credential defaults to a service account
		if _, ok := config["auth_type"]; !ok && f.connections[f.connection] == "bigquery_v1" {
			f.config["auth_type"] = "service-account-json"
		}
	}
	for k, v := range config {
		isSecret := false
		for _, secret := range secretKeys {
			if k == secret {
				isSecret = true
			}
		}
		if !isSecret {
			f.config[k] = v
		}
	}
}

func (f *fakeAPI) handlers() map[string]testhelpers.MockEndpointHandler {
	handlers := map[string]testhelpers.MockEndpointHandler{
		"POST " + collectionPath: func(r *http.Request) (int, interface{}, error) {
			f.mu.Lock()
			defer f.mu.Unlock()
			f.store(r, true)
			return http.StatusCreated, f.credential(), nil
		},
		"GET " + itemPath(credentialID): func(r *http.Request) (int, interface{}, error) {
			f.mu.Lock()
			defer f.mu.Unlock()
			return http.StatusOK, f.credential(), nil
		},
		"PATCH " + itemPath(credentialID): func(r *http.Request) (int, interface{}, error) {
			f.mu.Lock()
			defer f.mu.Unlock()
			f.store(r, false)
			return http.StatusOK, f.credential(), nil
		},
		"DELETE " + itemPath(credentialID): func(r *http.Request) (int, interface{}, error) {
			return http.StatusOK, map[string]interface{}{
				"status": map[string]interface{}{"code": 200, "is_success": true},
				"data":   map[string]interface{}{"id": credentialID},
			}, nil
		},
	}
	for id, adapter := range f.connections {
		handlers["GET "+connectionPath(id)] = connectionHandler(id, adapter)
	}
	return handlers
}

func connectionHandler(id int, adapter string) testhelpers.MockEndpointHandler {
	return func(r *http.Request) (int, interface{}, error) {
		return http.StatusOK, map[string]interface{}{
			"status": map[string]interface{}{"code": 200, "is_success": true},
			"data":   map[string]interface{}{"id": id, "adapter_version": adapter},
		}, nil
	}
}

func decodeBody(r *http.Request) map[string]interface{} {
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	return body
}

// calls returns the captured request bodies of one method and path
func calls(srv *testhelpers.MockServer, t *testing.T, method, path string) []map[string]interface{} {
	t.Helper()
	var out []map[string]interface{}
	for _, call := range srv.GetCapturedCalls(path) {
		if call.Method == method {
			out = append(out, call.BodyAsMap(t))
		}
	}
	return out
}

func countCalls(srv *testhelpers.MockServer, method string) int {
	n := 0
	for _, path := range []string{collectionPath, itemPath(credentialID)} {
		for _, call := range srv.GetCapturedCalls(path) {
			if call.Method == method {
				n++
			}
		}
	}
	return n
}

func providerBlock(hostURL string) string {
	return fmt.Sprintf(`
provider "dbtcloud" {
  host_url   = "%s"
  token      = "dummy-token"
  account_id = %d
}
`, hostURL, testAccountID)
}

// snowflakeConfig renders a credential with the given attributes inside the snowflake block
func snowflakeConfig(hostURL string, connectionID int, attributes string) string {
	return providerBlock(hostURL) + fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "test" {
  connection_id = %d
  snowflake = {
    %s
  }
}
`, connectionID, attributes)
}

const passwordAttributes = `auth_type = "password"
    user = "analytics_user"
    warehouse = "analytics_wh"
    password_wo = "super-secret-password"
    password_wo_version = 1`

const keypairAttributes = `auth_type = "keypair"
    user = "analytics_user"
    warehouse = "analytics_wh"
    private_key_wo = "super-secret-key"
    private_key_wo_version = 1
    private_key_passphrase_wo = "super-secret-passphrase"
    private_key_passphrase_wo_version = 1`

func unitTestCase(steps []tfresource.TestStep) tfresource.TestCase {
	return tfresource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		// write-only attributes require Terraform >= 1.11
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_11_0),
		},
		Steps: steps,
	}
}

func setupAcc(t *testing.T) {
	t.Helper()
	t.Setenv("TF_ACC", "1")
}

// --- direct ModifyPlan helpers -------------------------------------------------------

func newTestResource(t *testing.T, hostURL string) resource.ResourceWithModifyPlan {
	t.Helper()
	t.Setenv("TF_ACC", "1") // skips credential validation in NewClient
	accountID := int64(testAccountID)
	token := "dummy-token"
	maxRetries := 0
	retryInterval := 0
	timeout := 5
	skipValidation := true
	client, err := dbt_cloud.NewClient(
		&accountID, &token, &hostURL, &maxRetries, &retryInterval, nil, skipValidation, &timeout,
	)
	if err != nil {
		t.Fatalf("could not build the client: %v", err)
	}
	r := analytics_credential.AnalyticsCredentialResource()
	configureResp := &resource.ConfigureResponse{}
	r.(resource.ResourceWithConfigure).Configure(
		context.Background(), resource.ConfigureRequest{ProviderData: client}, configureResp,
	)
	if configureResp.Diagnostics.HasError() {
		t.Fatalf("configure failed: %v", configureResp.Diagnostics)
	}
	return r.(resource.ResourceWithModifyPlan)
}

func resourceSchema(t *testing.T, r resource.Resource) resource.SchemaResponse {
	t.Helper()
	resp := resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	return resp
}

func snowflakeModel(authType string) *analytics_credential.SnowflakeModel {
	return &analytics_credential.SnowflakeModel{
		AuthType:                      types.StringValue(authType),
		User:                          types.StringValue("analytics_user"),
		Role:                          types.StringNull(),
		Warehouse:                     types.StringValue("analytics_wh"),
		PasswordWo:                    types.StringNull(),
		PasswordWoVersion:             types.Int64Null(),
		PrivateKeyWo:                  types.StringNull(),
		PrivateKeyWoVersion:           types.Int64Null(),
		PrivateKeyPassphraseWo:        types.StringNull(),
		PrivateKeyPassphraseWoVersion: types.Int64Null(),
	}
}

func credentialModel(
	connectionID int64,
	snowflake *analytics_credential.SnowflakeModel,
) analytics_credential.AnalyticsCredentialResourceModel {
	return analytics_credential.AnalyticsCredentialResourceModel{
		ID:             types.Int64Value(credentialID),
		ConnectionID:   types.Int64Value(connectionID),
		AdapterVersion: types.StringValue("snowflake_v0"),
		Snowflake:      snowflake,
	}
}

func buildPlan(t *testing.T, s resource.SchemaResponse, model analytics_credential.AnalyticsCredentialResourceModel) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: s.Schema}
	if diags := plan.Set(context.Background(), &model); diags.HasError() {
		t.Fatalf("could not build the plan: %v", diags)
	}
	return plan
}

func buildState(t *testing.T, s resource.SchemaResponse, model analytics_credential.AnalyticsCredentialResourceModel) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: s.Schema}
	if diags := state.Set(context.Background(), &model); diags.HasError() {
		t.Fatalf("could not build the state: %v", diags)
	}
	return state
}

func nullState(s resource.SchemaResponse) tfsdk.State {
	return tfsdk.State{Schema: s.Schema, Raw: tftypes.NewValue(s.Schema.Type().TerraformType(context.Background()), nil)}
}

func nullPlan(s resource.SchemaResponse) tfsdk.Plan {
	return tfsdk.Plan{Schema: s.Schema, Raw: tftypes.NewValue(s.Schema.Type().TerraformType(context.Background()), nil)}
}

func runModifyPlan(
	t *testing.T,
	r resource.ResourceWithModifyPlan,
	plan tfsdk.Plan,
	state tfsdk.State,
) *resource.ModifyPlanResponse {
	t.Helper()
	resp := &resource.ModifyPlanResponse{Plan: plan}
	r.ModifyPlan(context.Background(), resource.ModifyPlanRequest{Plan: plan, State: state}, resp)
	return resp
}

func hasErrorSummary(resp *resource.ModifyPlanResponse, summary string) bool {
	for _, d := range resp.Diagnostics.Errors() {
		if strings.Contains(d.Summary(), summary) {
			return true
		}
	}
	return false
}

func requiresReplace(resp *resource.ModifyPlanResponse, p path.Path) bool {
	for _, candidate := range resp.RequiresReplace {
		if candidate.Equal(p) {
			return true
		}
	}
	return false
}

// --- R02 -------------------------------------------------------------------------------

// Rule: a credential has exactly one adapter block.
// given: a credential with no adapter block
// when: the configuration is validated
// then: the plan fails with the ExactlyOneOf error
func TestAnalyticsCredential_NoAdapterBlockFailsValidation(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "snowflake_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
		Config: providerBlock(srv.URL) + `
resource "dbtcloud_analytics_credential" "test" {
  connection_id = 42
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile(`Missing Attribute Configuration`),
	}}))
}

// Rule: the adapter block must match the adapter of the connection.
// given: a snowflake block on a connection whose adapter is databricks_v0
// when: the plan is made
// then: the plan fails with the mismatch error
func TestAnalyticsCredential_AdapterBlockMismatchFailsPlan(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "databricks_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
		Config:      snowflakeConfig(srv.URL, 42, passwordAttributes),
		PlanOnly:    true,
		ExpectError: regexp.MustCompile(`Adapter mismatch`),
	}}))
}

// Rule: the adapter block must match the adapter of the connection.
// given: a connection_id that is unknown at plan time
// when: the plan is made
// then: the adapter check is skipped and no connection is read
func TestAnalyticsCredential_UnknownConnectionSkipsAdapterCheck(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "databricks_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
		Config: providerBlock(srv.URL) + fmt.Sprintf(`
resource "terraform_data" "connection" {
  input = "42"
}

resource "dbtcloud_analytics_credential" "test" {
  connection_id = tonumber(terraform_data.connection.id)
  snowflake = {
    %s
  }
}`, passwordAttributes),
		PlanOnly:           true,
		ExpectNonEmptyPlan: true,
		Check: func(*terraform.State) error {
			if n := len(srv.GetCapturedCalls(connectionPath(42))); n != 0 {
				return fmt.Errorf("expected no connection read, got %d", n)
			}
			return nil
		},
	}}))
}

// Rule: the adapter block must match the adapter of the connection.
// given: connections of each known adapter version, an unsupported one, and API failures
// when: a snowflake block is planned against them
// then: only snowflake_v0 is accepted; the others fail as mismatch, unsupported adapter or API error
func TestAnalyticsCredential_AdapterLookupOutcomes(t *testing.T) {
	cases := []struct {
		name    string
		adapter string
		status  int
		summary string
	}{
		{"snowflake matches", "snowflake_v0", 200, ""},
		{"databricks mismatch", "databricks_v0", 200, "Adapter mismatch"},
		{"redshift mismatch", "redshift_v0", 200, "Adapter mismatch"},
		{"bigquery v0 mismatch", "bigquery_v0", 200, "Adapter mismatch"},
		{"bigquery v1 mismatch", "bigquery_v1", 200, "Adapter mismatch"},
		{"postgres unsupported", "postgres_v0", 200, "Unsupported adapter"},
		{"connection not found", "", http.StatusNotFound, "Error reading connection"},
		{"connection forbidden", "", http.StatusForbidden, "Error reading connection"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
				"GET " + connectionPath(42): func(r *http.Request) (int, interface{}, error) {
					if tc.status != 200 {
						return tc.status, map[string]interface{}{
							"status": map[string]interface{}{"code": tc.status, "is_success": false},
						}, nil
					}
					return connectionHandler(42, tc.adapter)(r)
				},
			})
			defer srv.Close()
			r := newTestResource(t, srv.URL)
			s := resourceSchema(t, r)

			resp := runModifyPlan(t, r, buildPlan(t, s, credentialModel(42, snowflakeModel("password"))), nullState(s))

			if tc.summary == "" {
				if resp.Diagnostics.HasError() {
					t.Fatalf("expected no error, got %v", resp.Diagnostics)
				}
				return
			}
			if !hasErrorSummary(resp, tc.summary) {
				t.Fatalf("expected %q, got %v", tc.summary, resp.Diagnostics)
			}
		})
	}
}

// Rule: the adapter block must match the adapter of the connection.
// given: a destroy plan
// when: ModifyPlan runs
// then: the connection is not read
func TestAnalyticsCredential_DestroyPlanReadsNoConnection(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	resp := runModifyPlan(t, r, nullPlan(s), buildState(t, s, credentialModel(42, snowflakeModel("password"))))

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
	if n := len(srv.GetCapturedCalls(connectionPath(42))); n != 0 {
		t.Fatalf("expected no connection read, got %d", n)
	}
}

// --- R01 -------------------------------------------------------------------------------

// Rule: the adapter of a credential cannot change in place.
// given: a state with a snowflake block, or with none
// when: the block is removed, or a block is added where there was none
// then: ModifyPlan appends replace for the block; a connection_id change is replace
func TestAnalyticsCredential_BlockPresenceChangeRequiresReplace(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "snowflake_v0"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)
	snowflakePath := path.Root("snowflake")

	t.Run("set to null", func(t *testing.T) {
		resp := runModifyPlan(t, r,
			buildPlan(t, s, credentialModel(42, nil)),
			buildState(t, s, credentialModel(42, snowflakeModel("password"))))
		if !requiresReplace(resp, snowflakePath) {
			t.Fatalf("expected replace on snowflake, got %v", resp.RequiresReplace)
		}
	})
	t.Run("null to set", func(t *testing.T) {
		resp := runModifyPlan(t, r,
			buildPlan(t, s, credentialModel(42, snowflakeModel("password"))),
			buildState(t, s, credentialModel(42, nil)))
		if !requiresReplace(resp, snowflakePath) {
			t.Fatalf("expected replace on snowflake, got %v", resp.RequiresReplace)
		}
	})
	t.Run("unchanged block", func(t *testing.T) {
		resp := runModifyPlan(t, r,
			buildPlan(t, s, credentialModel(42, snowflakeModel("password"))),
			buildState(t, s, credentialModel(42, snowflakeModel("password"))))
		if len(resp.RequiresReplace) != 0 {
			t.Fatalf("expected no replace, got %v", resp.RequiresReplace)
		}
	})
	t.Run("connection_id change", func(t *testing.T) {
		attribute, ok := s.Schema.GetAttributes()["connection_id"].(schema.Int64Attribute)
		if !ok {
			t.Fatal("connection_id is not an Int64 attribute in the schema")
		}
		for _, modifier := range attribute.PlanModifiers {
			if strings.Contains(modifier.Description(context.Background()), "destroy and recreate") {
				return
			}
		}
		t.Fatal("connection_id has no replace plan modifier")
	})
}

// Rule: the adapter of a credential cannot change in place.
// given: a credential on connection 42 whose config keeps the snowflake block
// when: connection_id moves to connection 43, whose adapter is databricks_v0
// then: the plan fails and the server records no DELETE and no POST
func TestAnalyticsCredential_MoveToMismatchingConnectionFailsWithoutDestroy(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(43): connectionHandler(43, "databricks_v0"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	resp := runModifyPlan(t, r,
		buildPlan(t, s, credentialModel(43, snowflakeModel("password"))),
		buildState(t, s, credentialModel(42, snowflakeModel("password"))))

	if !hasErrorSummary(resp, "Adapter mismatch") {
		t.Fatalf("expected the mismatch error, got %v", resp.Diagnostics)
	}
	if countCalls(srv, "DELETE") != 0 || countCalls(srv, "POST") != 0 {
		t.Fatalf("expected no DELETE and no POST")
	}
}

// Rule: the adapter of a credential cannot change in place.
// given: auth_type password in the state
// when: the configuration says keypair
// then: the plan is replace
func TestAnalyticsCredential_AuthTypeChangeRequiresReplace(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "snowflake_v0"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	resp := runModifyPlan(t, r,
		buildPlan(t, s, credentialModel(42, snowflakeModel("keypair"))),
		buildState(t, s, credentialModel(42, snowflakeModel("password"))))

	if !requiresReplace(resp, path.Root("snowflake").AtName("auth_type")) {
		t.Fatalf("expected replace on snowflake.auth_type, got %v", resp.RequiresReplace)
	}
}

// --- optional fields -------------------------------------------------------------------

// Rule: an optional field cannot be cleared once set.
// given: role set in the state
// when: role is removed from the configuration
// then: the plan fails with the documented error
func TestAnalyticsCredential_RemovingOptionalFieldFailsPlan(t *testing.T) {
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

	if !hasErrorSummary(resp, "Optional attribute cannot be removed") {
		t.Fatalf("expected the clearing error, got %v", resp.Diagnostics)
	}
}

// --- validators ------------------------------------------------------------------------

// Rule: the snowflake auth type decides which attributes are required and which are forbidden.
// given: snowflake blocks that miss a required attribute or mix the auth types
// when: the configuration is validated at plan time
// then: each case fails with the expected error
func TestAnalyticsCredential_SnowflakeValidators(t *testing.T) {
	cases := []struct {
		name       string
		attributes string
		expect     string
	}{
		{"missing user", `auth_type = "password"
    warehouse = "w"
    password_wo = "x"
    password_wo_version = 1`, `(?s)Incorrect attribute value type.*"user"`},
		{"missing warehouse", `auth_type = "password"
    user = "u"
    password_wo = "x"
    password_wo_version = 1`, `(?s)Incorrect attribute value type.*"warehouse"`},
		{"missing auth_type", `user = "u"
    warehouse = "w"
    password_wo = "x"
    password_wo_version = 1`, `(?s)Incorrect attribute value type.*"auth_type"`},
		{"password without password_wo", `auth_type = "password"
    user = "u"
    warehouse = "w"`, `Missing required attribute .password_wo.`},
		{"password with private_key_wo", `auth_type = "password"
    user = "u"
    warehouse = "w"
    password_wo = "x"
    password_wo_version = 1
    private_key_wo = "k"
    private_key_wo_version = 1`, `Attribute .private_key_wo. not allowed`},
		{"password with private_key_passphrase_wo", `auth_type = "password"
    user = "u"
    warehouse = "w"
    password_wo = "x"
    password_wo_version = 1
    private_key_passphrase_wo = "p"
    private_key_passphrase_wo_version = 1`, `Attribute .private_key_passphrase_wo. not allowed`},
		{"keypair without private_key_wo", `auth_type = "keypair"
    user = "u"
    warehouse = "w"`, `Missing required attribute .private_key_wo.`},
		{"keypair with password_wo", `auth_type = "keypair"
    user = "u"
    warehouse = "w"
    private_key_wo = "k"
    private_key_wo_version = 1
    password_wo = "x"
    password_wo_version = 1`, `Attribute .password_wo. not allowed`},
		{"secret without version", `auth_type = "password"
    user = "u"
    warehouse = "w"
    password_wo = "x"`, `Missing required attribute .password_wo_version.`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setupAcc(t)
			api := newFakeAPI(map[int]string{42: "snowflake_v0"})
			srv := testhelpers.SetupMockServer(t, api.handlers())
			defer srv.Close()

			tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
				Config:      snowflakeConfig(srv.URL, 42, tc.attributes),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(tc.expect),
			}}))
		})
	}
}

// Rule: the snowflake auth type decides which attributes are required and which are forbidden.
// given: valid password and keypair blocks that use write-only values inside the nested block
// when: the configuration is planned
// then: the validators resolve their relative paths and accept both
func TestAnalyticsCredential_ValidSnowflakeBlocksPlan(t *testing.T) {
	for name, attributes := range map[string]string{"password": passwordAttributes, "keypair": keypairAttributes} {
		t.Run(name, func(t *testing.T) {
			setupAcc(t)
			api := newFakeAPI(map[int]string{42: "snowflake_v0"})
			srv := testhelpers.SetupMockServer(t, api.handlers())
			defer srv.Close()

			tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
				Config:             snowflakeConfig(srv.URL, 42, attributes),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			}}))
		})
	}
}

// --- import ----------------------------------------------------------------------------

// Rule: the adapter of a credential cannot change in place.
// given: an applied snowflake credential
// when: it is imported by "<id>" and by "<account_id>:<id>", or with a malformed id
// then: the state is populated from the API for the valid ids and the bad id is an error
func TestAnalyticsCredential_Import(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "snowflake_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	ignore := []string{
		"snowflake.password_wo_version",
		"snowflake.private_key_wo_version",
		"snowflake.private_key_passphrase_wo_version",
	}
	tfresource.Test(t, unitTestCase([]tfresource.TestStep{
		{Config: snowflakeConfig(srv.URL, 42, passwordAttributes)},
		{
			Config:                  snowflakeConfig(srv.URL, 42, passwordAttributes),
			ResourceName:            "dbtcloud_analytics_credential.test",
			ImportState:             true,
			ImportStateId:           fmt.Sprintf("%d", credentialID),
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: ignore,
		},
		{
			Config:                  snowflakeConfig(srv.URL, 42, passwordAttributes),
			ResourceName:            "dbtcloud_analytics_credential.test",
			ImportState:             true,
			ImportStateId:           fmt.Sprintf("%d:%d", testAccountID, credentialID),
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: ignore,
		},
		{
			Config:        snowflakeConfig(srv.URL, 42, passwordAttributes),
			ResourceName:  "dbtcloud_analytics_credential.test",
			ImportState:   true,
			ImportStateId: "not-a-number",
			ExpectError:   regexp.MustCompile(`Invalid Import ID`),
		},
	}))
}

// --- databricks block ------------------------------------------------------------------

func databricksConfig(hostURL string, connectionID int, attributes string) string {
	return providerBlock(hostURL) + fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "test" {
  connection_id = %d
  databricks = {
    %s
  }
}
`, connectionID, attributes)
}

const databricksAttributes = `token_wo = "super-secret-token"
    token_wo_version = 1
    catalog = "main"`

func databricksModel() *analytics_credential.DatabricksModel {
	return &analytics_credential.DatabricksModel{
		TokenWo:        types.StringNull(),
		TokenWoVersion: types.Int64Value(1),
		Catalog:        types.StringNull(),
	}
}

func databricksCredentialModel(connectionID int64) analytics_credential.AnalyticsCredentialResourceModel {
	return analytics_credential.AnalyticsCredentialResourceModel{
		ID:             types.Int64Value(credentialID),
		ConnectionID:   types.Int64Value(connectionID),
		AdapterVersion: types.StringValue("databricks_v0"),
		Databricks:     databricksModel(),
	}
}

// Rule: a credential has exactly one adapter block.
// given: a databricks block together with a snowflake block
// when: the configuration is validated
// then: the plan fails with the ExactlyOneOf error; a databricks block alone is valid
func TestAnalyticsCredential_DatabricksAndSnowflakeBlocksTogetherFail(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "databricks_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	both := providerBlock(srv.URL) + fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "test" {
  connection_id = 42
  snowflake = {
    %s
  }
  databricks = {
    %s
  }
}`, passwordAttributes, databricksAttributes)

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{
		{
			Config:      both,
			PlanOnly:    true,
			ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
		},
		{
			Config:             databricksConfig(srv.URL, 42, databricksAttributes),
			PlanOnly:           true,
			ExpectNonEmptyPlan: true,
		},
	}))
}

// Rule: a credential has exactly one adapter block.
// given: a databricks block without token_wo or without its version
// when: the configuration is validated at plan time
// then: the plan fails
func TestAnalyticsCredential_DatabricksTokenIsRequired(t *testing.T) {
	cases := map[string]struct{ attributes, expect string }{
		"missing token": {
			`token_wo_version = 1
    catalog = "main"`, `(?s)Incorrect attribute value type.*"token_wo"`,
		},
		"missing version": {
			`token_wo = "x"
    catalog = "main"`, `(?s)Incorrect attribute value type.*"token_wo_version"`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			setupAcc(t)
			api := newFakeAPI(map[int]string{42: "databricks_v0"})
			srv := testhelpers.SetupMockServer(t, api.handlers())
			defer srv.Close()

			tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
				Config:      databricksConfig(srv.URL, 42, tc.attributes),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(tc.expect),
			}}))
		})
	}
}

// Rule: the adapter of a credential cannot change in place.
// given: a snowflake state
// when: the configuration moves to a databricks block on the same connection
// then: the plan fails at plan time with the adapter mismatch and the server records no DELETE
func TestAnalyticsCredential_SwapToDatabricksOnSameConnectionFailsWithoutDestroy(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "snowflake_v0"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	resp := runModifyPlan(t, r,
		buildPlan(t, s, databricksCredentialModel(42)),
		buildState(t, s, credentialModel(42, snowflakeModel("password"))))

	if !hasErrorSummary(resp, "Adapter mismatch") {
		t.Fatalf("expected the mismatch error, got %v", resp.Diagnostics)
	}
	if countCalls(srv, "DELETE") != 0 {
		t.Fatalf("expected no DELETE")
	}
}

// Rule: the adapter of a credential cannot change in place.
// given: a snowflake state on a Snowflake connection
// when: the configuration moves to a databricks block on a Databricks connection
// then: the plan is replace for both blocks
func TestAnalyticsCredential_SwapToDatabricksOnDatabricksConnectionRequiresReplace(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(43): connectionHandler(43, "databricks_v0"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	resp := runModifyPlan(t, r,
		buildPlan(t, s, databricksCredentialModel(43)),
		buildState(t, s, credentialModel(42, snowflakeModel("password"))))

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
	for _, block := range []string{"snowflake", "databricks"} {
		if !requiresReplace(resp, path.Root(block)) {
			t.Fatalf("expected replace on %s, got %v", block, resp.RequiresReplace)
		}
	}
}

// Rule: the adapter block must match the adapter of the connection.
// given: a databricks block
// when: it is planned against a Databricks connection and against a Snowflake one
// then: the first is accepted and the second is a mismatch
func TestAnalyticsCredential_DatabricksAdapterLookup(t *testing.T) {
	for adapter, summary := range map[string]string{"databricks_v0": "", "snowflake_v0": "Adapter mismatch"} {
		t.Run(adapter, func(t *testing.T) {
			srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
				"GET " + connectionPath(42): connectionHandler(42, adapter),
			})
			defer srv.Close()
			r := newTestResource(t, srv.URL)
			s := resourceSchema(t, r)

			resp := runModifyPlan(t, r, buildPlan(t, s, databricksCredentialModel(42)), nullState(s))

			if summary == "" && resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error: %v", resp.Diagnostics)
			}
			if summary != "" && !hasErrorSummary(resp, summary) {
				t.Fatalf("expected %q, got %v", summary, resp.Diagnostics)
			}
		})
	}
}

// --- redshift block --------------------------------------------------------------------

func redshiftConfig(hostURL string, connectionID int, attributes string) string {
	return providerBlock(hostURL) + fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "test" {
  connection_id = %d
  redshift = {
    %s
  }
}
`, connectionID, attributes)
}

const redshiftAttributes = `username = "analytics_user"
    password_wo = "super-secret-redshift"
    password_wo_version = 1`

func redshiftCredentialModel(connectionID int64) analytics_credential.AnalyticsCredentialResourceModel {
	return analytics_credential.AnalyticsCredentialResourceModel{
		ID:             types.Int64Value(credentialID),
		ConnectionID:   types.Int64Value(connectionID),
		AdapterVersion: types.StringValue("redshift_v0"),
		Redshift: &analytics_credential.RedshiftModel{
			Username:          types.StringValue("analytics_user"),
			PasswordWo:        types.StringNull(),
			PasswordWoVersion: types.Int64Value(1),
		},
	}
}

// Rule: a credential has exactly one adapter block.
// given: a redshift block together with another block
// when: the configuration is validated
// then: the plan fails with the ExactlyOneOf error
func TestAnalyticsCredential_RedshiftAndAnotherBlockFail(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "redshift_v0"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
		Config: providerBlock(srv.URL) + fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "test" {
  connection_id = 42
  redshift = {
    %s
  }
  databricks = {
    %s
  }
}`, redshiftAttributes, databricksAttributes),
		PlanOnly:    true,
		ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
	}}))
}

// Rule: the adapter block must match the adapter of the connection.
// given: a redshift block
// when: it is planned against a Redshift connection and against a Snowflake one
// then: the first is accepted and the second is a mismatch
func TestAnalyticsCredential_RedshiftAdapterLookup(t *testing.T) {
	for adapter, summary := range map[string]string{"redshift_v0": "", "snowflake_v0": "Adapter mismatch"} {
		t.Run(adapter, func(t *testing.T) {
			srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
				"GET " + connectionPath(42): connectionHandler(42, adapter),
			})
			defer srv.Close()
			r := newTestResource(t, srv.URL)
			s := resourceSchema(t, r)

			resp := runModifyPlan(t, r, buildPlan(t, s, redshiftCredentialModel(42)), nullState(s))

			if summary == "" && resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error: %v", resp.Diagnostics)
			}
			if summary != "" && !hasErrorSummary(resp, summary) {
				t.Fatalf("expected %q, got %v", summary, resp.Diagnostics)
			}
		})
	}
}

// Rule: the adapter of a credential cannot change in place.
// given: a redshift state
// when: the configuration moves to another block on the same connection, or to a connection of another adapter with that block
// then: the first fails at plan time and sends no DELETE; the second is replace
func TestAnalyticsCredential_RedshiftSwapRules(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "redshift_v0"),
		"GET " + connectionPath(43): connectionHandler(43, "databricks_v0"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	t.Run("same connection fails", func(t *testing.T) {
		resp := runModifyPlan(t, r,
			buildPlan(t, s, credentialModel(42, snowflakeModel("password"))),
			buildState(t, s, redshiftCredentialModel(42)))
		if !hasErrorSummary(resp, "Adapter mismatch") {
			t.Fatalf("expected the mismatch error, got %v", resp.Diagnostics)
		}
		if countCalls(srv, "DELETE") != 0 {
			t.Fatalf("expected no DELETE")
		}
	})
	t.Run("other adapter connection replaces", func(t *testing.T) {
		resp := runModifyPlan(t, r,
			buildPlan(t, s, databricksCredentialModel(43)),
			buildState(t, s, redshiftCredentialModel(42)))
		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %v", resp.Diagnostics)
		}
		for _, block := range []string{"redshift", "databricks"} {
			if !requiresReplace(resp, path.Root(block)) {
				t.Fatalf("expected replace on %s, got %v", block, resp.RequiresReplace)
			}
		}
	})
}

// Rule: the redshift block needs a username and a password.
// given: a redshift block without username, without password_wo, or without the version
// when: the configuration is validated at plan time
// then: the plan fails
func TestAnalyticsCredential_RedshiftValidators(t *testing.T) {
	cases := map[string]struct{ attributes, expect string }{
		"missing username": {
			`password_wo = "x"
    password_wo_version = 1`, `(?s)Incorrect attribute value type.*"username"`,
		},
		"missing password": {
			`username = "u"
    password_wo_version = 1`, `(?s)Incorrect attribute value type.*"password_wo"`,
		},
		"missing version": {
			`username = "u"
    password_wo = "x"`, `(?s)Incorrect attribute value type.*"password_wo_version"`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			setupAcc(t)
			api := newFakeAPI(map[int]string{42: "redshift_v0"})
			srv := testhelpers.SetupMockServer(t, api.handlers())
			defer srv.Close()

			tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
				Config:      redshiftConfig(srv.URL, 42, tc.attributes),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(tc.expect),
			}}))
		})
	}
}

// --- bigquery block --------------------------------------------------------------------

func bigqueryConfig(hostURL string, connectionID int, attributes string) string {
	return providerBlock(hostURL) + fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "test" {
  connection_id = %d
  bigquery = {
    %s
  }
}
`, connectionID, attributes)
}

const serviceAccountFields = `private_key_id = "key-id"
    private_key_wo = "super-secret-bq-key"
    private_key_wo_version = 1
    client_email = "sa@example.iam.gserviceaccount.com"
    client_id = "123"
    auth_uri = "https://accounts.google.com/o/oauth2/auth"
    token_uri = "https://oauth2.googleapis.com/token"
    auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
    client_x509_cert_url = "https://www.googleapis.com/robot/v1/metadata/x509/sa"`

const wifFields = `workload_pool_provider_path = "//iam.googleapis.com/projects/1/locations/global/workloadIdentityPools/p/providers/q"`

func bigqueryCredentialModel(
	connectionID int64,
	adapterVersion string,
	authType *string,
) analytics_credential.AnalyticsCredentialResourceModel {
	bigquery := &analytics_credential.BigQueryModel{
		AuthType:                       types.StringNull(),
		ExecutionProject:               types.StringNull(),
		PrivateKeyID:                   types.StringValue("key-id"),
		PrivateKeyWo:                   types.StringNull(),
		PrivateKeyWoVersion:            types.Int64Value(1),
		ClientEmail:                    types.StringValue("sa@example.iam.gserviceaccount.com"),
		ClientID:                       types.StringValue("123"),
		AuthURI:                        types.StringValue("https://accounts.google.com/o/oauth2/auth"),
		TokenURI:                       types.StringValue("https://oauth2.googleapis.com/token"),
		AuthProviderX509CertURL:        types.StringValue("https://www.googleapis.com/oauth2/v1/certs"),
		ClientX509CertURL:              types.StringValue("https://www.googleapis.com/robot/v1/metadata/x509/sa"),
		WorkloadPoolProviderPath:       types.StringNull(),
		ServiceAccountImpersonationURL: types.StringNull(),
	}
	if authType != nil {
		bigquery.AuthType = types.StringValue(*authType)
	}
	return analytics_credential.AnalyticsCredentialResourceModel{
		ID:             types.Int64Value(credentialID),
		ConnectionID:   types.Int64Value(connectionID),
		AdapterVersion: types.StringValue(adapterVersion),
		BigQuery:       bigquery,
	}
}

func strPtr(s string) *string { return &s }

// Rule: a credential has exactly one adapter block.
// given: a bigquery block together with a redshift block
// when: the configuration is validated
// then: the plan fails with the ExactlyOneOf error
func TestAnalyticsCredential_BigQueryAndAnotherBlockFail(t *testing.T) {
	setupAcc(t)
	api := newFakeAPI(map[int]string{42: "bigquery_v1"})
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
		Config: providerBlock(srv.URL) + fmt.Sprintf(`
resource "dbtcloud_analytics_credential" "test" {
  connection_id = 42
  bigquery = {
    %s
  }
  redshift = {
    %s
  }
}`, serviceAccountFields, redshiftAttributes),
		PlanOnly:    true,
		ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
	}}))
}

// Rule: the adapter block must match the adapter of the connection.
// given: a bigquery block
// when: it is planned against bigquery_v0 and bigquery_v1 connections and against a Snowflake one
// then: both BigQuery versions are accepted and the Snowflake connection is a mismatch
func TestAnalyticsCredential_BigQueryAdapterLookup(t *testing.T) {
	for adapter, summary := range map[string]string{"bigquery_v1": "", "snowflake_v0": "Adapter mismatch"} {
		t.Run(adapter, func(t *testing.T) {
			srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
				"GET " + connectionPath(42): connectionHandler(42, adapter),
			})
			defer srv.Close()
			r := newTestResource(t, srv.URL)
			s := resourceSchema(t, r)

			resp := runModifyPlan(t, r, buildPlan(t, s, bigqueryCredentialModel(42, "bigquery_v1", nil)), nullState(s))

			if summary == "" && resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error: %v", resp.Diagnostics)
			}
			if summary != "" && !hasErrorSummary(resp, summary) {
				t.Fatalf("expected %q, got %v", summary, resp.Diagnostics)
			}
		})
	}
}

// Rule: a bigquery_v0 connection has no auth type.
// given: a bigquery block with auth_type on a bigquery_v0 connection
// when: the plan is made
// then: the plan fails; without auth_type it is accepted
func TestAnalyticsCredential_BigQueryV0RejectsAuthType(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "bigquery_v0"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	withAuthType := runModifyPlan(t, r,
		buildPlan(t, s, bigqueryCredentialModel(42, "bigquery_v0", strPtr("service-account-json"))), nullState(s))
	if !hasErrorSummary(withAuthType, "Unsupported attribute for adapter") {
		t.Fatalf("expected the auth_type error, got %v", withAuthType.Diagnostics)
	}

	without := runModifyPlan(t, r, buildPlan(t, s, bigqueryCredentialModel(42, "bigquery_v0", nil)), nullState(s))
	if without.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", without.Diagnostics)
	}
}

// Rule: the bigquery auth type decides which fields are required and which are forbidden.
// given: bigquery blocks that miss a required field or mix the two auth types
// when: the configuration is validated at plan time
// then: each case fails with the expected error
func TestAnalyticsCredential_BigQueryValidators(t *testing.T) {
	without := func(field string) string {
		var kept []string
		for _, line := range strings.Split(serviceAccountFields, "\n") {
			if !strings.HasPrefix(strings.TrimSpace(line), field+" ") && !strings.HasPrefix(strings.TrimSpace(line), field+"=") {
				kept = append(kept, line)
			}
		}
		return strings.Join(kept, "\n")
	}
	cases := []struct{ name, attributes, expect string }{
		{"service account missing private_key_wo", without("private_key_wo"), "Missing required attribute .private_key_wo."},
		{"service account missing private_key_id", without("private_key_id"), "Missing required attribute .private_key_id."},
		{"service account missing client_email", without("client_email"), "Missing required attribute .client_email."},
		{"service account missing client_id", without("client_id"), "Missing required attribute .client_id."},
		{"service account missing auth_uri", without("auth_uri"), "Missing required attribute .auth_uri."},
		{"service account missing token_uri", without("token_uri"), "Missing required attribute .token_uri."},
		{"service account missing auth_provider_x509_cert_url", without("auth_provider_x509_cert_url"), "Missing required attribute .auth_provider_x509_cert_url."},
		{"service account missing client_x509_cert_url", without("client_x509_cert_url"), "Missing required attribute .client_x509_cert_url."},
		{"service account missing key version", without("private_key_wo_version"), "Missing required attribute .private_key_wo_version."},
		{"service account with wif fields", serviceAccountFields + "\n    " + wifFields, "Attribute .workload_pool_provider_path. not allowed"},
		{"wif missing path", `auth_type = "external-oauth-wif"`, "Missing required attribute .workload_pool_provider_path."},
		{"wif with service account fields", `auth_type = "external-oauth-wif"
    ` + wifFields + `
    client_email = "sa@example.com"`, "Attribute .client_email. not allowed"},
		{"wif with private key", `auth_type = "external-oauth-wif"
    ` + wifFields + `
    private_key_wo = "k"
    private_key_wo_version = 1`, "Attribute .private_key_wo. not allowed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setupAcc(t)
			api := newFakeAPI(map[int]string{42: "bigquery_v1"})
			srv := testhelpers.SetupMockServer(t, api.handlers())
			defer srv.Close()

			tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
				Config:      bigqueryConfig(srv.URL, 42, tc.attributes),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(tc.expect),
			}}))
		})
	}
}

// Rule: the bigquery auth type decides which fields are required and which are forbidden.
// given: valid service account and workload identity federation blocks
// when: the configuration is planned
// then: both are accepted
func TestAnalyticsCredential_ValidBigQueryBlocksPlan(t *testing.T) {
	for name, attributes := range map[string]string{
		"service account default": serviceAccountFields,
		"service account explicit": `auth_type = "service-account-json"
    ` + serviceAccountFields,
		"wif": `auth_type = "external-oauth-wif"
    ` + wifFields + `
    service_account_impersonation_url = "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/sa@example.com:generateAccessToken"`,
	} {
		t.Run(name, func(t *testing.T) {
			setupAcc(t)
			api := newFakeAPI(map[int]string{42: "bigquery_v1"})
			srv := testhelpers.SetupMockServer(t, api.handlers())
			defer srv.Close()

			tfresource.Test(t, unitTestCase([]tfresource.TestStep{{
				Config:             bigqueryConfig(srv.URL, 42, attributes),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			}}))
		})
	}
}

// Rule: the adapter of a credential cannot change in place.
// given: a bigquery state
// when: the configuration moves to another block on the same connection, to another BigQuery connection with another block, or flips auth_type
// then: the first fails at plan time without a DELETE and the others are replace
func TestAnalyticsCredential_BigQuerySwapRules(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "bigquery_v1"),
		"GET " + connectionPath(43): connectionHandler(43, "databricks_v0"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)
	state := buildState(t, s, bigqueryCredentialModel(42, "bigquery_v1", strPtr("service-account-json")))

	t.Run("same connection fails", func(t *testing.T) {
		resp := runModifyPlan(t, r, buildPlan(t, s, redshiftCredentialModel(42)), state)
		if !hasErrorSummary(resp, "Adapter mismatch") {
			t.Fatalf("expected the mismatch error, got %v", resp.Diagnostics)
		}
		if countCalls(srv, "DELETE") != 0 {
			t.Fatalf("expected no DELETE")
		}
	})
	t.Run("other adapter connection replaces", func(t *testing.T) {
		resp := runModifyPlan(t, r, buildPlan(t, s, databricksCredentialModel(43)), state)
		for _, block := range []string{"bigquery", "databricks"} {
			if !requiresReplace(resp, path.Root(block)) {
				t.Fatalf("expected replace on %s, got %v", block, resp.RequiresReplace)
			}
		}
	})
	t.Run("auth_type flip replaces", func(t *testing.T) {
		flipped := bigqueryCredentialModel(42, "bigquery_v1", strPtr("external-oauth-wif"))
		resp := runModifyPlan(t, r, buildPlan(t, s, flipped), state)
		if !requiresReplace(resp, path.Root("bigquery").AtName("auth_type")) {
			t.Fatalf("expected replace on bigquery.auth_type, got %v", resp.RequiresReplace)
		}
	})
	t.Run("default auth_type is not a flip", func(t *testing.T) {
		resp := runModifyPlan(t, r, buildPlan(t, s, bigqueryCredentialModel(42, "bigquery_v1", nil)), state)
		if len(resp.RequiresReplace) != 0 || resp.Diagnostics.HasError() {
			t.Fatalf("expected no replace and no error, got %v %v", resp.RequiresReplace, resp.Diagnostics)
		}
	})
}

// Rule: an optional field cannot be cleared once set.
// given: execution_project set in the state
// when: it is removed from the configuration
// then: the plan fails with the documented error
func TestAnalyticsCredential_RemovingExecutionProjectFailsPlan(t *testing.T) {
	srv := testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"GET " + connectionPath(42): connectionHandler(42, "bigquery_v1"),
	})
	defer srv.Close()
	r := newTestResource(t, srv.URL)
	s := resourceSchema(t, r)

	withProject := bigqueryCredentialModel(42, "bigquery_v1", nil)
	withProject.BigQuery.ExecutionProject = types.StringValue("billing-project")

	resp := runModifyPlan(t, r,
		buildPlan(t, s, bigqueryCredentialModel(42, "bigquery_v1", nil)),
		buildState(t, s, withProject))

	if !hasErrorSummary(resp, "Optional attribute cannot be removed") {
		t.Fatalf("expected the clearing error, got %v", resp.Diagnostics)
	}
}
