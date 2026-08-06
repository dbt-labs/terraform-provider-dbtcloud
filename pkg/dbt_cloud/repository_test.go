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

	backend := "github"
	fullName := "test/repo"
	server.SetCreateResponse(&dbt_cloud.RepositoryResponse{
		Data: dbt_cloud.Repository{
			ID:                        testutil.IntPtr(repositoryID),
			RemoteBackend:             &backend,
			FullName:                  &fullName,
			DeployKeyID:               testutil.IntPtr(deployKeyID),
			RepositoryCredentialsID:   testutil.IntPtr(credentialsID),
		},
	})

	server.SetUpdateResponse(&dbt_cloud.RepositoryResponse{
		Data: dbt_cloud.Repository{
			ID:                        testutil.IntPtr(repositoryID),
			RemoteBackend:             &backend,
			FullName:                  &fullName,
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

	if updateReq.RemoteBackend == nil || *updateReq.RemoteBackend != backend {
		t.Errorf("Expected update request to have RemoteBackend '%s', got '%v'", backend, updateReq.RemoteBackend)
	}

	if updateReq.FullName == nil || *updateReq.FullName != fullName {
		t.Errorf("Expected update request to have FullName '%s', got '%v'", fullName, updateReq.FullName)
	}

	if *updateReq.DeployKeyID != deployKeyID {
		t.Errorf("Expected update request to have DeployKeyID %d, got %d", deployKeyID, *updateReq.DeployKeyID)
	}

	if *updateReq.RepositoryCredentialsID != credentialsID {
		t.Errorf("Expected update request to have RepositoryCredentialsID %d, got %d", credentialsID, *updateReq.RepositoryCredentialsID)
	}
}

// TestCreateRepository_ADO_StripsAzureFieldsOnUpdate ensures that the Azure
// DevOps fields, which are create-only (RequiresReplace) and accepted only by
// the collection-create endpoint, are NOT sent on the follow-up update call.
// The detail/update endpoint rejects them with a 400 ("Additional properties
// are not allowed"), so UpdateRepository must strip them.
func TestCreateRepository_ADO_StripsAzureFieldsOnUpdate(t *testing.T) {
	const (
		accountID    = 123
		projectID    = 456
		repositoryID = 789
	)

	server := testutil.NewMockRepositoryServer(accountID, projectID, repositoryID)
	defer server.Close()

	backend := "ado"
	azureProjectID := "12345678-1234-1234-1234-1234567890ab"
	azureRepositoryID := "87654321-4321-abcd-abcd-464327678642"
	bypass := false

	adoResponse := &dbt_cloud.RepositoryResponse{
		Data: dbt_cloud.Repository{
			ID:                                    testutil.IntPtr(repositoryID),
			RemoteBackend:                         &backend,
			AzureActiveDirectoryProjectID:         &azureProjectID,
			AzureActiveDirectoryRepositoryID:      &azureRepositoryID,
			AzureBypassWebhookRegistrationFailure: &bypass,
		},
	}
	server.SetCreateResponse(adoResponse)
	server.SetUpdateResponse(adoResponse)

	client := testutil.CreateTestClient(server.URL(), accountID)

	// A pull_request_url_template forces the follow-up UpdateRepository call.
	_, err := client.CreateRepository(
		projectID,
		"https://abc@dev.azure.com/abc/def/_git/my_repo",
		true,
		"azure_active_directory_app",
		0,
		0,
		"",
		azureProjectID,
		azureRepositoryID,
		bypass,
		"https://dev.azure.com/abc/def/pullrequestcreate?sourceRef={{source}}&targetRef={{destination}}",
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updateReq := server.GetLastUpdateRequest()
	if updateReq == nil {
		t.Fatal("Expected update request to be made")
	}

	if updateReq.AzureActiveDirectoryProjectID != nil {
		t.Errorf("Expected AzureActiveDirectoryProjectID to be stripped from update, got '%v'", *updateReq.AzureActiveDirectoryProjectID)
	}
	if updateReq.AzureActiveDirectoryRepositoryID != nil {
		t.Errorf("Expected AzureActiveDirectoryRepositoryID to be stripped from update, got '%v'", *updateReq.AzureActiveDirectoryRepositoryID)
	}
	if updateReq.AzureBypassWebhookRegistrationFailure != nil {
		t.Errorf("Expected AzureBypassWebhookRegistrationFailure to be stripped from update, got '%v'", *updateReq.AzureBypassWebhookRegistrationFailure)
	}
}