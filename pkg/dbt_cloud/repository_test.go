package dbt_cloud_test

import (
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud/testutil"
)

func TestCreateRepository_PreservesRemoteBackendAndFullName(t *testing.T) {
	const (
		accountID = 123
		projectID = 456
		repositoryID = 789
		deployKeyID = 101
		credentialsID = 202
	)

	server := testutil.NewMockRepositoryServer(accountID, projectID, repositoryID)
	defer server.Close()

	server.SetCreateResponse(&dbt_cloud.RepositoryResponse{
		Data: dbt_cloud.Repository{
			ID:                        testutil.IntPtr(repositoryID),
			RemoteBackend:             "github",
			FullName:                  "test/repo",
			DeployKeyID:               testutil.IntPtr(deployKeyID),
			RepositoryCredentialsID:   testutil.IntPtr(credentialsID),
		},
	})

	server.SetUpdateResponse(&dbt_cloud.RepositoryResponse{
		Data: dbt_cloud.Repository{
			ID:                        testutil.IntPtr(repositoryID),
			RemoteBackend:             "github",
			FullName:                  "test/repo",
			DeployKeyID:               testutil.IntPtr(deployKeyID),
			RepositoryCredentialsID:   testutil.IntPtr(credentialsID),
		},
	})

	client := testutil.CreateTestClient(server.URL(), accountID)

	_, err := client.CreateRepository(
		projectID,
		"git@github.com:test/repo.git",
		true,
		"github_app",
		0,
		0,
		"",
		"",
		false,
		"https://github.com/test/repo/compare/{{destination}}...{{source}}",
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updateReq := server.GetLastUpdateRequest()

	if updateReq == nil {
		t.Fatal("Expected update request to be made")
	}

	if updateReq.RemoteBackend != "github" {
		t.Errorf("Expected update request to have RemoteBackend 'github', got %s", updateReq.RemoteBackend)
	}

	if updateReq.FullName != "test/repo" {
		t.Errorf("Expected update request to have FullName 'test/repo', got %s", updateReq.FullName)
	}

	if *updateReq.DeployKeyID != deployKeyID {
		t.Errorf("Expected update request to have DeployKeyID %d, got %d", deployKeyID, *updateReq.DeployKeyID)
	}

	if *updateReq.RepositoryCredentialsID != credentialsID {
		t.Errorf("Expected update request to have RepositoryCredentialsID %d, got %d", credentialsID, *updateReq.RepositoryCredentialsID)
	}
}