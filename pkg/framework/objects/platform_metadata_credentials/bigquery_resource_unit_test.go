package platform_metadata_credentials_test

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const (
	bqAccountID    = 1
	bqConnectionID = 42
	bqResourceName = "dbtcloud_bigquery_platform_metadata_credential.test"
	bqClientEmail  = "sa@my-project.iam.gserviceaccount.com"
	// what dbt Cloud fills in from the connection's impersonate_service_account
	bqImpersonationURL = "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/impersonated@my-project.iam.gserviceaccount.com:generateAccessToken"
	bqWorkloadPoolPath = "//iam.googleapis.com/projects/1/locations/global/workloadIdentityPools/pool/providers/provider"
)

var (
	bqCollectionPath       = fmt.Sprintf("/v3/accounts/%d/platform-metadata-credentials/", bqAccountID)
	bqServiceAccountFields = []string{
		"private_key_id", "client_email", "client_id", "auth_uri", "token_uri",
		"auth_provider_x509_cert_url", "client_x509_cert_url",
	}
	bqWIFFields = []string{"workload_pool_provider_path", "service_account_impersonation_url"}
)

func bqItemPath(id int) string {
	return fmt.Sprintf("%s%d/", bqCollectionPath, id)
}

// fakeBigQueryAPI keeps one platform metadata credential and behaves like the dbt Cloud API:
//   - bigquery_v0 has no auth_type: it drops auth_type and the WIF fields, and only accepts a
//     service account key;
//   - PATCH merges the sent config over the stored one, and replaces it when auth_type changes;
//   - for WIF on a v1 connection with impersonate_service_account, the impersonation URL is filled
//     in when it is missing;
//   - GET returns the private key scrubbed.
type fakeBigQueryAPI struct {
	t                     *testing.T
	adapterVersion        string
	connectionImpersonate bool

	nextID int
	stored map[string]interface{}
	srv    *testhelpers.MockServer
}

func newFakeBigQueryAPI(t *testing.T, adapterVersion string, connectionImpersonate bool) *fakeBigQueryAPI {
	t.Helper()
	t.Setenv("TF_ACC", "1")

	api := &fakeBigQueryAPI{
		t:                     t,
		adapterVersion:        adapterVersion,
		connectionImpersonate: connectionImpersonate,
		nextID:                100,
	}
	// a credential keeps its ID for its whole life, so the item path is known up front
	itemPath := bqItemPath(api.nextID)
	api.srv = testhelpers.SetupMockServer(t, map[string]testhelpers.MockEndpointHandler{
		"POST " + bqCollectionPath: api.create,
		"GET " + itemPath:          api.get,
		"PATCH " + itemPath:        api.patch,
		"DELETE " + itemPath:       api.delete,
	})
	t.Cleanup(api.srv.Close)
	return api
}

func (a *fakeBigQueryAPI) requestBody(r *http.Request) map[string]interface{} {
	a.t.Helper()
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		a.t.Fatalf("reading request body: %v", err)
	}
	body := map[string]interface{}{}
	if err := json.Unmarshal(raw, &body); err != nil {
		a.t.Fatalf("decoding request body %q: %v", raw, err)
	}
	return body
}

// validate returns the config as dbt Cloud would store it, or the 400 response body.
func (a *fakeBigQueryAPI) validate(config map[string]interface{}) (map[string]interface{}, map[string]interface{}) {
	config = maps.Clone(config)
	if a.adapterVersion == "bigquery_v0" {
		delete(config, "auth_type")
		for _, field := range bqWIFFields {
			delete(config, field)
		}
	}

	if config["auth_type"] == "external-oauth-wif" {
		if config["service_account_impersonation_url"] == nil && a.connectionImpersonate {
			config["service_account_impersonation_url"] = bqImpersonationURL
		}
		return config, nil
	}

	var missing []string
	for _, field := range append([]string{"private_key"}, bqServiceAccountFields...) {
		if config[field] == nil {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		errs := map[string]interface{}{}
		for _, field := range missing {
			errs[field] = []string{"field required"}
		}
		return nil, map[string]interface{}{
			"status": map[string]interface{}{"code": 400, "is_success": false, "user_message": "Invalid request"},
			"data":   errs,
		}
	}
	return config, nil
}

func (a *fakeBigQueryAPI) response(status int) (int, interface{}, error) {
	config := maps.Clone(a.stored["config"].(map[string]interface{}))
	if _, ok := config["private_key"]; ok {
		config["private_key"] = "**********"
	}
	data := maps.Clone(a.stored)
	data["config"] = config
	return status, map[string]interface{}{
		"status": map[string]interface{}{"code": status, "is_success": true},
		"data":   data,
	}, nil
}

func (a *fakeBigQueryAPI) create(r *http.Request) (int, interface{}, error) {
	body := a.requestBody(r)
	config, errBody := a.validate(body["config"].(map[string]interface{}))
	if errBody != nil {
		return http.StatusBadRequest, errBody, nil
	}
	a.stored = map[string]interface{}{
		"id":                        a.nextID,
		"account_id":                bqAccountID,
		"connection_id":             body["connection_id"],
		"adapter_version":           a.adapterVersion,
		"catalog_ingestion_enabled": body["catalog_ingestion_enabled"],
		"cost_optimization_enabled": body["cost_optimization_enabled"],
		"cost_insights_enabled":     body["cost_insights_enabled"],
		"config":                    config,
	}
	return a.response(http.StatusCreated)
}

func (a *fakeBigQueryAPI) get(_ *http.Request) (int, interface{}, error) {
	if a.stored == nil {
		return http.StatusNotFound, map[string]interface{}{
			"status": map[string]interface{}{"code": 404, "is_success": false, "user_message": "Not found"},
		}, nil
	}
	return a.response(http.StatusOK)
}

func (a *fakeBigQueryAPI) patch(r *http.Request) (int, interface{}, error) {
	body := a.requestBody(r)
	sent := body["config"].(map[string]interface{})
	stored := a.stored["config"].(map[string]interface{})

	merged := maps.Clone(stored)
	maps.Copy(merged, sent)
	config, errBody := a.validate(merged)
	if errBody != nil {
		return http.StatusBadRequest, errBody, nil
	}
	if sent["auth_type"] != stored["auth_type"] && a.adapterVersion != "bigquery_v0" {
		// the stored config is replaced, so the stale fields of the old auth type go away
		replaced := map[string]interface{}{}
		for key := range sent {
			replaced[key] = config[key]
		}
		if url, ok := config["service_account_impersonation_url"]; ok && sent["auth_type"] == "external-oauth-wif" {
			replaced["service_account_impersonation_url"] = url
		}
		config = replaced
	}

	for _, flag := range []string{"catalog_ingestion_enabled", "cost_optimization_enabled", "cost_insights_enabled"} {
		a.stored[flag] = body[flag]
	}
	a.stored["config"] = config
	return a.response(http.StatusOK)
}

func (a *fakeBigQueryAPI) delete(_ *http.Request) (int, interface{}, error) {
	a.stored = nil
	return http.StatusOK, map[string]interface{}{
		"status": map[string]interface{}{"code": 200, "is_success": true},
	}, nil
}

// sentConfigs returns the config of every request with the given method, in order.
func (a *fakeBigQueryAPI) sentConfigs(method string) []map[string]interface{} {
	a.t.Helper()
	path := bqCollectionPath
	if method != http.MethodPost {
		path = bqItemPath(100)
	}
	var configs []map[string]interface{}
	for _, call := range a.srv.GetCapturedCalls(path) {
		if call.Method == method {
			config, _ := call.BodyAsMap(a.t)["config"].(map[string]interface{})
			configs = append(configs, config)
		}
	}
	return configs
}

// checkSent checks the config of the index-th request with the given method: the fields in want
// must have that value, and the fields in absent must not be sent.
func (a *fakeBigQueryAPI) checkSent(method string, index int, want map[string]string, absent ...string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		configs := a.sentConfigs(method)
		if len(configs) <= index {
			return fmt.Errorf("expected at least %d %s request(s), got %d", index+1, method, len(configs))
		}
		config := configs[index]
		for field, value := range want {
			if got, _ := config[field].(string); got != value {
				return fmt.Errorf("%s #%d: expected %s = %q, got %q (config: %v)", method, index, field, value, got, config)
			}
		}
		for _, field := range absent {
			if _, ok := config[field]; ok {
				return fmt.Errorf("%s #%d: expected no %s, got %v", method, index, field, config[field])
			}
		}
		return nil
	}
}

func (a *fakeBigQueryAPI) checkRequestCount(method string, want int) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if got := len(a.sentConfigs(method)); got != want {
			return fmt.Errorf("expected %d %s request(s), got %d", want, method, got)
		}
		return nil
	}
}

func bqConfig(hostURL, body string) string {
	return fmt.Sprintf(`
provider "dbtcloud" {
  host_url   = %q
  token      = "dummy-token"
  account_id = %d
}

resource "dbtcloud_bigquery_platform_metadata_credential" "test" {
  connection_id = %d
%s
}
`, hostURL, bqAccountID, bqConnectionID, body)
}

// bqServiceAccountBody returns service account attributes. keyLines sets the private key, and
// extra is added as is.
func bqServiceAccountBody(keyLines, extra string) string {
	return fmt.Sprintf(`
  private_key_id              = "key-id-1"
  client_email                = %q
  client_id                   = "1234567890"
  auth_uri                    = "https://accounts.google.com/o/oauth2/auth"
  token_uri                   = "https://oauth2.googleapis.com/token"
  auth_provider_x509_cert_url = "https://www.googleapis.com/oauth2/v1/certs"
  client_x509_cert_url        = "https://www.googleapis.com/robot/v1/metadata/x509/sa"
  %s
  %s
`, bqClientEmail, keyLines, extra)
}

func bqWIFBody(extra string) string {
	return fmt.Sprintf(`
  auth_type                   = "external-oauth-wif"
  workload_pool_provider_path = %q
  %s
`, bqWorkloadPoolPath, extra)
}

var bqWriteOnlyVersionCheck = []tfversion.TerraformVersionCheck{
	tfversion.SkipBelow(tfversion.Version1_11_0),
}

func bqExpectURL(value knownvalue.Check) plancheck.PlanCheck {
	return plancheck.ExpectKnownValue(bqResourceName, tfjsonpath.New("service_account_impersonation_url"), value)
}

// A service account key on bigquery_v0 with auth_type left out. v0 never stores auth_type, so
// reading it back must not show a diff (the framework re-plans after each step).
func TestBigQueryPMC_ServiceAccountOnV0(t *testing.T) {
	api := newFakeBigQueryAPI(t, "bigquery_v0", false)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bqConfig(api.srv.URL, bqServiceAccountBody(`private_key = "first-key"`, "")),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{bqExpectURL(knownvalue.Null())},
				},
				Check: resource.ComposeTestCheckFunc(
					api.checkSent(http.MethodPost, 0, map[string]string{
						"auth_type":                   "service-account-json",
						"private_key":                 "first-key",
						"private_key_id":              "key-id-1",
						"client_email":                bqClientEmail,
						"client_id":                   "1234567890",
						"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
						"token_uri":                   "https://oauth2.googleapis.com/token",
						"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
						"client_x509_cert_url":        "https://www.googleapis.com/robot/v1/metadata/x509/sa",
					}, bqWIFFields...),
					func(*terraform.State) error {
						if _, ok := api.stored["config"].(map[string]interface{})["auth_type"]; ok {
							return fmt.Errorf("the v0 fake must not store auth_type")
						}
						return nil
					},
					resource.TestCheckResourceAttr(bqResourceName, "id", "1:100"),
					resource.TestCheckResourceAttr(bqResourceName, "credential_id", "100"),
					resource.TestCheckResourceAttr(bqResourceName, "auth_type", "service-account-json"),
					resource.TestCheckResourceAttr(bqResourceName, "adapter_version", "bigquery_v0"),
					resource.TestCheckResourceAttr(bqResourceName, "private_key", "first-key"),
					resource.TestCheckNoResourceAttr(bqResourceName, "service_account_impersonation_url"),
					resource.TestCheckNoResourceAttr(bqResourceName, "workload_pool_provider_path"),
				),
			},
			{
				// refresh reads the v0 response without auth_type
				RefreshState: true,
				Check:        resource.TestCheckResourceAttr(bqResourceName, "auth_type", "service-account-json"),
			},
		},
	})
}

func TestBigQueryPMC_Create_WIF(t *testing.T) {
	for name, tc := range map[string]struct {
		connectionImpersonate bool
		// empty means no URL
		wantURL string
	}{
		"filled from the connection":       {true, bqImpersonationURL},
		"connection without impersonation": {false, ""},
	} {
		t.Run(name, func(t *testing.T) {
			api := newFakeBigQueryAPI(t, "bigquery_v1", tc.connectionImpersonate)

			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: bqConfig(api.srv.URL, bqWIFBody("")),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectUnknownValue(bqResourceName, tfjsonpath.New("service_account_impersonation_url")),
							},
							PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
						},
						Check: resource.ComposeTestCheckFunc(
							api.checkSent(http.MethodPost, 0, map[string]string{
								"auth_type":                   "external-oauth-wif",
								"workload_pool_provider_path": bqWorkloadPoolPath,
							}, append([]string{"private_key", "service_account_impersonation_url"}, bqServiceAccountFields...)...),
							resource.TestCheckNoResourceAttr(bqResourceName, "private_key_id"),
							func(state *terraform.State) error {
								if tc.wantURL == "" {
									return resource.TestCheckNoResourceAttr(bqResourceName, "service_account_impersonation_url")(state)
								}
								return resource.TestCheckResourceAttr(bqResourceName, "service_account_impersonation_url", tc.wantURL)(state)
							},
						),
					},
				},
			})
		})
	}
}

// Write-only values are null in the plan, so the key has to come from the config: it must be
// in the POST body on create and in the PATCH body when the version is bumped.
func TestBigQueryPMC_WriteOnlyKey_CreateAndRotate(t *testing.T) {
	api := newFakeBigQueryAPI(t, "bigquery_v1", false)
	config := func(key string, version int) string {
		return bqConfig(api.srv.URL, bqServiceAccountBody(
			fmt.Sprintf("private_key_wo = %q\n  private_key_wo_version = %d", key, version), "",
		))
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		TerraformVersionChecks:   bqWriteOnlyVersionCheck,
		Steps: []resource.TestStep{
			{
				Config: config("first-key", 1),
				Check: resource.ComposeTestCheckFunc(
					api.checkSent(http.MethodPost, 0, map[string]string{"private_key": "first-key"}),
					resource.TestCheckNoResourceAttr(bqResourceName, "private_key_wo"),
					resource.TestCheckNoResourceAttr(bqResourceName, "private_key"),
					resource.TestCheckResourceAttr(bqResourceName, "private_key_wo_version", "1"),
				),
			},
			{
				// same version, new value: the version is what triggers an update
				Config:   config("ignored-key", 1),
				PlanOnly: true,
			},
			{
				Config: config("second-key", 2),
				Check: resource.ComposeTestCheckFunc(
					api.checkSent(http.MethodPatch, 0, map[string]string{"private_key": "second-key"}),
					resource.TestCheckNoResourceAttr(bqResourceName, "private_key_wo"),
					resource.TestCheckResourceAttr(bqResourceName, "private_key_wo_version", "2"),
					resource.TestCheckResourceAttr(bqResourceName, "credential_id", "100"),
				),
			},
		},
	})
}

// The validators on private_key point at private_key_wo with a relative path. They only run when
// the plain key is set, so a plan with it catches path expression errors.
func TestBigQueryPMC_PlainSecret_PlanOnly(t *testing.T) {
	api := newFakeBigQueryAPI(t, "bigquery_v1", false)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             bqConfig(api.srv.URL, bqServiceAccountBody(`private_key = "plain-key"`, "")),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestBigQueryPMC_Update_KeyUnchanged(t *testing.T) {
	api := newFakeBigQueryAPI(t, "bigquery_v1", false)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bqConfig(api.srv.URL, bqServiceAccountBody(`private_key = "first-key"`, "")),
			},
			{
				Config: bqConfig(api.srv.URL, bqServiceAccountBody(`private_key = "first-key"`, "cost_insights_enabled = true")),
				Check: resource.ComposeTestCheckFunc(
					api.checkSent(http.MethodPatch, 0, map[string]string{
						"auth_type":      "service-account-json",
						"private_key_id": "key-id-1",
					}, "private_key"),
					// the stored key is kept by the PATCH merge
					func(*terraform.State) error {
						if got := api.stored["config"].(map[string]interface{})["private_key"]; got != "first-key" {
							return fmt.Errorf("expected the stored key to be kept, got %v", got)
						}
						return nil
					},
					resource.TestCheckResourceAttr(bqResourceName, "cost_insights_enabled", "true"),
				),
			},
		},
	})
}

func TestBigQueryPMC_Update_RotateKey(t *testing.T) {
	api := newFakeBigQueryAPI(t, "bigquery_v0", false)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bqConfig(api.srv.URL, bqServiceAccountBody(`private_key = "first-key"`, "")),
			},
			{
				Config: bqConfig(api.srv.URL, strings.Replace(
					bqServiceAccountBody(`private_key = "second-key"`, ""), "key-id-1", "key-id-2", 1,
				)),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(bqResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					api.checkSent(http.MethodPatch, 0, map[string]string{
						"private_key":    "second-key",
						"private_key_id": "key-id-2",
					}),
					api.checkRequestCount(http.MethodPost, 1),
					resource.TestCheckResourceAttr(bqResourceName, "credential_id", "100"),
					resource.TestCheckResourceAttr(bqResourceName, "private_key_id", "key-id-2"),
				),
			},
		},
	})
}

func TestBigQueryPMC_Update_SwitchToWIF(t *testing.T) {
	api := newFakeBigQueryAPI(t, "bigquery_v1", true)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bqConfig(api.srv.URL, bqServiceAccountBody(`private_key = "first-key"`, "")),
			},
			{
				Config: bqConfig(api.srv.URL, bqWIFBody("")),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(bqResourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue(bqResourceName, tfjsonpath.New("service_account_impersonation_url")),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					api.checkSent(http.MethodPatch, 0, map[string]string{
						"auth_type":                   "external-oauth-wif",
						"workload_pool_provider_path": bqWorkloadPoolPath,
					}, append([]string{"private_key", "service_account_impersonation_url"}, bqServiceAccountFields...)...),
					resource.TestCheckResourceAttr(bqResourceName, "credential_id", "100"),
					resource.TestCheckResourceAttr(bqResourceName, "service_account_impersonation_url", bqImpersonationURL),
					resource.TestCheckNoResourceAttr(bqResourceName, "private_key_id"),
				),
			},
		},
	})
}

// The stored WIF config has no key for the PATCH merge to keep, so the key must be sent, and the
// URL the server filled in must be planned away.
func TestBigQueryPMC_Update_SwitchToSA(t *testing.T) {
	api := newFakeBigQueryAPI(t, "bigquery_v1", true)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		TerraformVersionChecks:   bqWriteOnlyVersionCheck,
		Steps: []resource.TestStep{
			{
				Config: bqConfig(api.srv.URL, bqWIFBody("")),
				Check:  resource.TestCheckResourceAttr(bqResourceName, "service_account_impersonation_url", bqImpersonationURL),
			},
			{
				Config: bqConfig(api.srv.URL, bqServiceAccountBody(
					"private_key_wo = \"wo-key\"\n  private_key_wo_version = 1", `auth_type = "service-account-json"`,
				)),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(bqResourceName, plancheck.ResourceActionUpdate),
						bqExpectURL(knownvalue.Null()),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					api.checkSent(http.MethodPatch, 0, map[string]string{
						"auth_type":   "service-account-json",
						"private_key": "wo-key",
					}, bqWIFFields...),
					resource.TestCheckResourceAttr(bqResourceName, "credential_id", "100"),
					resource.TestCheckNoResourceAttr(bqResourceName, "service_account_impersonation_url"),
				),
			},
		},
	})
}

// The same switch with the plain key: the key is unchanged in Terraform, but the stored WIF config
// has none, so it still has to be sent.
func TestBigQueryPMC_Update_SwitchToSA_PlainKeyUnchanged(t *testing.T) {
	api := newFakeBigQueryAPI(t, "bigquery_v1", false)
	saConfig := bqConfig(api.srv.URL, bqServiceAccountBody(`private_key = "plain-key"`, ""))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: saConfig},
			{Config: bqConfig(api.srv.URL, bqWIFBody(""))},
			{
				Config: saConfig,
				Check: api.checkSent(http.MethodPatch, 1, map[string]string{
					"auth_type":   "service-account-json",
					"private_key": "plain-key",
				}),
			},
		},
	})
}

// Switched to WIF outside Terraform: the refresh keeps the key in state, so the key looks
// unchanged, but the stored WIF config has none for the PATCH merge to keep.
func TestBigQueryPMC_Update_SwitchedToWIFOutsideTerraform(t *testing.T) {
	api := newFakeBigQueryAPI(t, "bigquery_v1", false)
	config := bqConfig(api.srv.URL, bqServiceAccountBody(`private_key = "plain-key"`, ""))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config},
			{
				PreConfig: func() {
					api.stored["config"] = map[string]interface{}{
						"auth_type":                   "external-oauth-wif",
						"workload_pool_provider_path": bqWorkloadPoolPath,
					}
				},
				Config: config,
				Check: api.checkSent(http.MethodPatch, 0, map[string]string{
					"auth_type":   "service-account-json",
					"private_key": "plain-key",
				}),
			},
		},
	})
}

func TestBigQueryPMC_Update_WIFKeepsURL(t *testing.T) {
	api := newFakeBigQueryAPI(t, "bigquery_v1", true)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: bqConfig(api.srv.URL, bqWIFBody("")),
			},
			{
				Config: bqConfig(api.srv.URL, bqWIFBody("cost_insights_enabled = true")),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(bqResourceName, plancheck.ResourceActionUpdate),
						bqExpectURL(knownvalue.StringExact(bqImpersonationURL)),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(bqResourceName, "cost_insights_enabled", "true"),
					resource.TestCheckResourceAttr(bqResourceName, "service_account_impersonation_url", bqImpersonationURL),
				),
			},
		},
	})
}

func TestBigQueryPMC_Create_WIFOnV0Hint(t *testing.T) {
	api := newFakeBigQueryAPI(t, "bigquery_v0", false)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      bqConfig(api.srv.URL, bqWIFBody("")),
				ExpectError: regexp.MustCompile(`external-oauth-wif requires a bigquery_v1 connection`),
			},
		},
	})
}

func TestBigQueryPMC_Read_NotFound(t *testing.T) {
	api := newFakeBigQueryAPI(t, "bigquery_v1", false)
	config := bqConfig(api.srv.URL, bqServiceAccountBody(`private_key = "first-key"`, ""))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config},
			{
				// deleted outside Terraform: the refresh drops it from state, and the apply creates it again
				PreConfig: func() { api.stored = nil },
				Config:    config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(bqResourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: api.checkRequestCount(http.MethodPost, 2),
			},
		},
	})
}

func TestBigQueryPMC_ImportState(t *testing.T) {
	api := newFakeBigQueryAPI(t, "bigquery_v1", true)
	config := bqConfig(api.srv.URL, bqWIFBody(""))

	importStep := func(id string) resource.TestStep {
		return resource.TestStep{
			Config:            config,
			ResourceName:      bqResourceName,
			ImportState:       true,
			ImportStateId:     id,
			ImportStateVerify: true,
		}
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config},
			importStep("100"),
			importStep("1:100"),
			func() resource.TestStep {
				step := importStep("1:2:100")
				step.ExpectError = regexp.MustCompile(`Expected format: credential_id or account_id:credential_id`)
				return step
			}(),
			func() resource.TestStep {
				step := importStep("abc")
				step.ExpectError = regexp.MustCompile(`Could not parse credential_id as integer`)
				return step
			}(),
		},
	})
}

// After importing a service account credential the key is not in state, so the first apply sends
// it again. The other attributes match.
func TestBigQueryPMC_ImportState_ServiceAccount(t *testing.T) {
	api := newFakeBigQueryAPI(t, "bigquery_v0", false)
	config := bqConfig(api.srv.URL, bqServiceAccountBody(`private_key = "first-key"`, ""))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config},
			{
				Config:                  config,
				ResourceName:            bqResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key"},
			},
		},
	})
}
