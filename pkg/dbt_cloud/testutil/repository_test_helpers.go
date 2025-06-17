package testutil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
)

type MockRepositoryServer struct {
	server *httptest.Server
	createResponse *dbt_cloud.RepositoryResponse
	updateResponse *dbt_cloud.RepositoryResponse
	accountID int
	projectID int
	repositoryID int
	lastUpdateRequest *dbt_cloud.Repository
}

func (m *MockRepositoryServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	createPath := fmt.Sprintf("/v3/accounts/%d/projects/%d/repositories/", m.accountID, m.projectID)
	updatePath := fmt.Sprintf("/v3/accounts/%d/projects/%d/repositories/%d/", m.accountID, m.projectID, m.repositoryID)

	if r.Method == "POST" && r.URL.Path == createPath {
		if m.createResponse != nil {
			json.NewEncoder(w).Encode(m.createResponse)
			return
		}
	}

	if r.Method == "POST" && r.URL.Path == updatePath {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var updateReq dbt_cloud.Repository
		if err := json.Unmarshal(body, &updateReq); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.lastUpdateRequest = &updateReq

		if m.updateResponse != nil {
			json.NewEncoder(w).Encode(m.updateResponse)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func (m *MockRepositoryServer) GetLastUpdateRequest() *dbt_cloud.Repository {
	return m.lastUpdateRequest
}

func NewMockRepositoryServer(accountID, projectID, repositoryID int) *MockRepositoryServer {
	mock := &MockRepositoryServer{
		accountID: accountID,
		projectID: projectID,
		repositoryID: repositoryID,
	}
	mock.server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))
	return mock
}

func (m *MockRepositoryServer) SetCreateResponse(response *dbt_cloud.RepositoryResponse) {
	m.createResponse = response
}

func (m *MockRepositoryServer) SetUpdateResponse(response *dbt_cloud.RepositoryResponse) {
	m.updateResponse = response
}

func (m *MockRepositoryServer) Close() {
	m.server.Close()
}

func (m *MockRepositoryServer) URL() string {
	return m.server.URL
}

func CreateTestClient(serverURL string, accountID int) *dbt_cloud.Client {
	return &dbt_cloud.Client{
		HostURL:    serverURL,
		HTTPClient: &http.Client{},
		AccountID:  accountID,
	}
}

func IntPtr(i int) *int {
	return &i
} 