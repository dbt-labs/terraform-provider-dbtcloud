package dbt_cloud

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// projectServer answers every request with the given project payload and records
// the bodies of the POST requests
func projectServer(t *testing.T, project string) (*httptest.Server, *[]map[string]any) {
	t.Helper()
	var posts []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			parsed := map[string]any{}
			_ = json.Unmarshal(body, &parsed)
			posts = append(posts, parsed)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":` + project + `,"status":{"code":200,"is_success":true}}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &posts
}

// Rule: updating a project keeps its analytics connection.
// given: a project read with analytics_connection_id set
// when: another attribute is updated and the project is posted back
// then: the POST body still carries the same analytics_connection_id.
func TestUpdateProject_KeepsAnalyticsConnection(t *testing.T) {
	srv, posts := projectServer(t, `{"id":123,"account_id":123,"name":"p","type":0,"state":1,"analytics_connection_id":99}`)
	c := newRetryTestClient(t, srv.URL, 1, nil)

	project, err := c.GetProject("123")
	require.NoError(t, err)
	project.Name = "renamed"
	_, err = c.UpdateProject("123", *project)
	require.NoError(t, err)

	require.Len(t, *posts, 1)
	assert.EqualValues(t, 99, (*posts)[0]["analytics_connection_id"])
}

// Rule: a server that does not know the analytics connection is not sent one.
// given: a project response without analytics_connection_id
// when: the project is updated
// then: the POST body has no analytics_connection_id key.
func TestUpdateProject_OmitsAnalyticsConnectionWhenServerSentNone(t *testing.T) {
	srv, posts := projectServer(t, `{"id":123,"account_id":123,"name":"p","type":0,"state":1}`)
	c := newRetryTestClient(t, srv.URL, 1, nil)

	project, err := c.GetProject("123")
	require.NoError(t, err)
	_, err = c.UpdateProject("123", *project)
	require.NoError(t, err)

	require.Len(t, *posts, 1)
	_, exists := (*posts)[0]["analytics_connection_id"]
	assert.False(t, exists, "analytics_connection_id should not be sent")
}

// Rule: updating a project keeps its analytics connection.
// given: a project read with analytics_connection_id null
// when: the project is posted back
// then: the POST body carries an explicit null, as the server read it.
func TestUpdateProject_KeepsNullAnalyticsConnection(t *testing.T) {
	srv, posts := projectServer(t, `{"id":123,"account_id":123,"name":"p","type":0,"state":1,"analytics_connection_id":null}`)
	c := newRetryTestClient(t, srv.URL, 1, nil)

	project, err := c.GetProject("123")
	require.NoError(t, err)
	_, err = c.UpdateProject("123", *project)
	require.NoError(t, err)

	require.Len(t, *posts, 1)
	value, exists := (*posts)[0]["analytics_connection_id"]
	assert.True(t, exists)
	assert.Nil(t, value)
}

// Rule: a project can be created with an analytics connection.
// given: a create call with and without an analytics connection
// when: the body is sent
// then: it carries the id when given and has no key otherwise.
func TestCreateProject_AnalyticsConnection(t *testing.T) {
	srv, posts := projectServer(t, `{"id":123,"account_id":123,"name":"p","type":0,"state":1}`)
	c := newRetryTestClient(t, srv.URL, 1, nil)

	connectionID := int64(55)
	_, err := c.CreateProject("p", "", "", 0, &connectionID)
	require.NoError(t, err)
	_, err = c.CreateProject("p", "", "", 0, nil)
	require.NoError(t, err)

	require.Len(t, *posts, 2)
	assert.EqualValues(t, 55, (*posts)[0]["analytics_connection_id"])
	_, exists := (*posts)[1]["analytics_connection_id"]
	assert.False(t, exists)
}

// Rule: the project list exposes the analytics connection.
// given: a project list response with analytics_connection_id
// when: the projects are listed
// then: the value is decoded.
func TestGetAllProjects_DecodesAnalyticsConnection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":1,"name":"p","analytics_connection_id":77},{"id":2,"name":"q","analytics_connection_id":null}],"extra":{"pagination":{"count":2,"total_count":2}},"status":{"code":200}}`))
	}))
	defer srv.Close()
	c := newRetryTestClient(t, srv.URL, 1, nil)

	projects, err := c.GetAllProjects("")
	require.NoError(t, err)
	require.Len(t, projects, 2)
	require.NotNil(t, projects[0].AnalyticsConnectionID)
	assert.EqualValues(t, 77, *projects[0].AnalyticsConnectionID)
	assert.Nil(t, projects[1].AnalyticsConnectionID)
}
