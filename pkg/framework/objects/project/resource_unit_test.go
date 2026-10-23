package project_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/testhelpers"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	unitAccountID = 1
	unitProjectID = 5
	// an analytics connection of another account: the server answers 404
	foreignConnectionID = 404
)

var (
	projectsPath    = fmt.Sprintf("/v3/accounts/%d/projects/", unitAccountID)
	projectItemPath = fmt.Sprintf("/v3/accounts/%d/projects/%d/", unitAccountID, unitProjectID)
)

// fakeProjects mimics the project endpoints, including the server behaviour that an
// update without analytics_connection_id clears it
type fakeProjects struct {
	mu                  sync.Mutex
	name                string
	description         string
	analyticsConnection any
	state               int
}

func (f *fakeProjects) project() map[string]interface{} {
	return map[string]interface{}{
		"id":                       unitProjectID,
		"account_id":               unitAccountID,
		"name":                     f.name,
		"description":              f.description,
		"type":                     0,
		"state":                    f.state,
		"dbt_project_subdirectory": nil,
		"connection_id":            nil,
		"repository_id":            nil,
		"freshness_job_id":         nil,
		"docs_job_id":              nil,
		"analytics_connection_id":  f.analyticsConnection,
	}
}

func envelope(code int, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"status": map[string]interface{}{"code": code, "is_success": code < 400},
		"data":   data,
	}
}

func (f *fakeProjects) handlers() map[string]testhelpers.MockEndpointHandler {
	apply := func(r *http.Request) (int, interface{}, error) {
		body := map[string]interface{}{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if name, ok := body["name"].(string); ok {
			f.name = name
		}
		if description, ok := body["description"].(string); ok {
			f.description = description
		}
		if state, ok := body["state"].(float64); ok {
			f.state = int(state)
		}
		f.analyticsConnection = body["analytics_connection_id"]
		if id, ok := body["analytics_connection_id"].(float64); ok && int(id) == foreignConnectionID && f.state != 2 {
			return http.StatusNotFound, envelope(404, nil), nil
		}
		return http.StatusOK, envelope(200, f.project()), nil
	}
	return map[string]testhelpers.MockEndpointHandler{
		"POST " + projectsPath: func(r *http.Request) (int, interface{}, error) {
			f.mu.Lock()
			defer f.mu.Unlock()
			code, response, err := apply(r)
			if code == http.StatusOK {
				code = http.StatusCreated
			}
			return code, response, err
		},
		"GET " + projectItemPath: func(r *http.Request) (int, interface{}, error) {
			f.mu.Lock()
			defer f.mu.Unlock()
			return http.StatusOK, envelope(200, f.project()), nil
		},
		"POST " + projectItemPath: func(r *http.Request) (int, interface{}, error) {
			f.mu.Lock()
			defer f.mu.Unlock()
			return apply(r)
		},
		"GET " + projectsPath: func(r *http.Request) (int, interface{}, error) {
			f.mu.Lock()
			defer f.mu.Unlock()
			response := envelope(200, []interface{}{f.project()})
			response["extra"] = map[string]interface{}{"pagination": map[string]interface{}{"count": 1, "total_count": 1}}
			return http.StatusOK, response, nil
		},
	}
}

func unitProjectConfig(hostURL, attributes string) string {
	return fmt.Sprintf(`
provider "dbtcloud" {
  host_url   = "%s"
  token      = "dummy-token"
  account_id = %d
}

resource "dbtcloud_project" "test" {
  name = "test-project"
  %s
}
`, hostURL, unitAccountID, attributes)
}

func unitProjectCase(t *testing.T, steps []resource.TestStep) resource.TestCase {
	t.Setenv("TF_ACC", "1")
	return resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps:                    steps,
	}
}

func lastItemPost(srv *testhelpers.MockServer, t *testing.T) map[string]interface{} {
	t.Helper()
	var last map[string]interface{}
	for _, call := range srv.GetCapturedCalls(projectItemPath) {
		if call.Method == http.MethodPost {
			last = call.BodyAsMap(t)
		}
	}
	return last
}

// Rule: a project can have an analytics connection, which is sent when set and cleared when removed.
// given: a project without an analytics connection
// when: analytics_connection_id is set, another attribute is updated, and the attribute is removed
// then: the POST carries the id, keeps it on the unrelated update, and sends an explicit null on removal.
func TestProjectResource_AnalyticsConnectionLifecycle(t *testing.T) {
	api := &fakeProjects{name: "test-project", state: 1}
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	resource.Test(t, unitProjectCase(t, []resource.TestStep{
		{
			Config: unitProjectConfig(srv.URL, ""),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckNoResourceAttr("dbtcloud_project.test", "analytics_connection_id"),
				func(*terraform.State) error {
					for _, call := range srv.GetCapturedCalls(projectsPath) {
						if call.Method == http.MethodPost {
							if _, ok := call.BodyAsMap(t)["analytics_connection_id"]; ok {
								return fmt.Errorf("the create body carries analytics_connection_id")
							}
						}
					}
					return nil
				},
			),
		},
		{
			Config: unitProjectConfig(srv.URL, "analytics_connection_id = 42"),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("dbtcloud_project.test", "analytics_connection_id", "42"),
				func(*terraform.State) error {
					if got := lastItemPost(srv, t)["analytics_connection_id"]; got != float64(42) {
						return fmt.Errorf("expected analytics_connection_id 42, got %v", got)
					}
					return nil
				},
			),
		},
		{
			Config: unitProjectConfig(srv.URL, "analytics_connection_id = 42\n  description = \"new description\""),
			Check: func(*terraform.State) error {
				if got := lastItemPost(srv, t)["analytics_connection_id"]; got != float64(42) {
					return fmt.Errorf("an unrelated update lost analytics_connection_id, got %v", got)
				}
				return nil
			},
		},
		{
			Config: unitProjectConfig(srv.URL, "description = \"new description\""),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckNoResourceAttr("dbtcloud_project.test", "analytics_connection_id"),
				func(*terraform.State) error {
					body := lastItemPost(srv, t)
					value, ok := body["analytics_connection_id"]
					if !ok || value != nil {
						return fmt.Errorf("expected an explicit null, got %v (present: %v)", value, ok)
					}
					return nil
				},
			),
		},
	}))
}

// Rule: a project can be created with an analytics connection.
// given: a configuration that sets analytics_connection_id
// when: the project is created
// then: the create body carries the id.
func TestProjectResource_CreateWithAnalyticsConnection(t *testing.T) {
	api := &fakeProjects{name: "test-project", state: 1}
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	resource.Test(t, unitProjectCase(t, []resource.TestStep{{
		Config: unitProjectConfig(srv.URL, "analytics_connection_id = 42"),
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("dbtcloud_project.test", "analytics_connection_id", "42"),
			func(*terraform.State) error {
				for _, call := range srv.GetCapturedCalls(projectsPath) {
					if call.Method == http.MethodPost {
						if got := call.BodyAsMap(t)["analytics_connection_id"]; got != float64(42) {
							return fmt.Errorf("expected analytics_connection_id 42 on create, got %v", got)
						}
						return nil
					}
				}
				return fmt.Errorf("no create call")
			},
		),
	}}))
}

// Rule: a connection of another account cannot be the analytics connection.
// given: a project and a connection the server answers 404 for
// when: analytics_connection_id points at it
// then: the 404 is surfaced as an error.
func TestProjectResource_AnalyticsConnectionOfAnotherAccountFails(t *testing.T) {
	api := &fakeProjects{name: "test-project", state: 1}
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	resource.Test(t, unitProjectCase(t, []resource.TestStep{
		{Config: unitProjectConfig(srv.URL, "")},
		{
			Config:      unitProjectConfig(srv.URL, fmt.Sprintf("analytics_connection_id = %d", foreignConnectionID)),
			ExpectError: regexp.MustCompile(`Error updating project`),
		},
	}))
}

// Rule: the project data sources expose the analytics connection.
// given: a project with an analytics connection
// when: it is read through the project data source and the projects data source
// then: both expose analytics_connection_id.
func TestProjectDataSources_AnalyticsConnection(t *testing.T) {
	api := &fakeProjects{name: "test-project", state: 1, analyticsConnection: float64(42)}
	srv := testhelpers.SetupMockServer(t, api.handlers())
	defer srv.Close()

	resource.Test(t, unitProjectCase(t, []resource.TestStep{{
		Config: fmt.Sprintf(`
provider "dbtcloud" {
  host_url   = "%s"
  token      = "dummy-token"
  account_id = %d
}

data "dbtcloud_project" "one" {
  id = %d
}

data "dbtcloud_projects" "all" {}
`, srv.URL, unitAccountID, unitProjectID),
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr("data.dbtcloud_project.one", "analytics_connection_id", "42"),
			resource.TestCheckTypeSetElemNestedAttrs("data.dbtcloud_projects.all", "projects.*", map[string]string{
				"id":                      "5",
				"analytics_connection_id": "42",
			}),
		),
	}}))
}
