package acctest_helper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/provider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helperTestResource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func SharedClient() (*dbt_cloud.Client, error) {

	accountIDString := os.Getenv("DBT_CLOUD_ACCOUNT_ID")
	accountID, _ := strconv.Atoi(accountIDString)
	token := os.Getenv("DBT_CLOUD_TOKEN")
	hostURL := os.Getenv("DBT_CLOUD_HOST_URL")

	if hostURL == "" {
		hostURL = "https://cloud.getdbt.com/api"
	}

	parsedURL, err := url.Parse(hostURL)
	if err != nil {
		panic(fmt.Sprintf("failed to parse serverURL: %s, error: %v", hostURL, err))
	}
	client := dbt_cloud.Client{
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
		HostURL:      parsedURL,
		Token:        token,
		AccountID:    accountID,
		DisableRetry: true,
	}

	return &client, nil
}

var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"dbtcloud": func() (tfprotov6.ProviderServer, error) {
		return providerserver.NewProtocol6(provider.New())(), nil
	},
}

func TestAccPreCheck(t *testing.T) {
	if v := os.Getenv("DBT_CLOUD_ACCOUNT_ID"); v == "" {
		t.Fatal("DBT_CLOUD_ACCOUNT_ID must be set for acceptance tests")
	}
	if v := os.Getenv("DBT_CLOUD_TOKEN"); v == "" {
		t.Fatal("DBT_CLOUD_TOKEN must be set for acceptance tests")
	}
}

func HelperTestResourceSchema[R resource.Resource](t *testing.T, r R) {
	ctx := context.Background()

	req := resource.SchemaRequest{}
	res := resource.SchemaResponse{}

	r.Schema(ctx, req, &res)

	if res.Diagnostics.HasError() {
		t.Fatalf("Error in schema: %v", res.Diagnostics)
	}

	diags := res.Schema.ValidateImplementation(ctx)

	if diags.HasError() {
		t.Fatalf("Error in schema validation: %v", diags)
	}
}

func HelperTestDataSourceSchema[DS datasource.DataSource](t *testing.T, ds DS) {
	ctx := context.Background()

	req := datasource.SchemaRequest{}
	res := datasource.SchemaResponse{}

	ds.Schema(ctx, req, &res)

	if res.Diagnostics.HasError() {
		t.Fatalf("Error in schema: %v", res.Diagnostics)
	}

	diags := res.Schema.ValidateImplementation(ctx)

	if diags.HasError() {
		t.Fatalf("Error in schema validation: %v", diags)
	}
}

func MakeExternalProviderTestStep(ts helperTestResource.TestStep, frameworkVersion string) helperTestResource.TestStep {
	return helperTestResource.TestStep{
		ExternalProviders: map[string]helperTestResource.ExternalProvider{
			"dbtcloud": {
				VersionConstraint: frameworkVersion,
				Source:            "dbt-labs/dbtcloud",
			},
		},
		Config: ts.Config,
		Check:  ts.Check,
	}
}

func MakeCurrentProviderNoOpTestStep(ts helperTestResource.TestStep) helperTestResource.TestStep {
	return helperTestResource.TestStep{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Config:                   ts.Config,
		ConfigPlanChecks: helperTestResource.ConfigPlanChecks{
			PreApply: []plancheck.PlanCheck{
				plancheck.ExpectEmptyPlan(),
			},
		},
		Check: nil,
	}
}

// This is used to test the acceptance tests against the current provider version using
// a real environment, as oppossed to a mocked one. This is useful when we want to test
// real envs, but using the mocked one simplifies testing scenarios.
func GetSemanticLayerConfigTestingConfigurations() (int, int, int) {
	environmentId := os.Getenv("DBT_CLOUD_ENVIRONMENT_ID_1")
	environmentId2 := os.Getenv("DBT_CLOUD_ENVIRONMENT_ID_2")
	projectId := os.Getenv("DBT_CLOUD_PROJECT_ID")

	if environmentId == "" || environmentId2 == "" || projectId == "" {
		return 0, 0, 0
	}

	envId, _ := strconv.Atoi(environmentId)
	envId2, _ := strconv.Atoi(environmentId2)
	projectIdInt, _ := strconv.Atoi(projectId)

	return envId, envId2, projectIdInt
}

// PlatformMetadataCredentialConfig holds the configuration for platform metadata credential tests
type PlatformMetadataCredentialConfig struct {
	// Snowflake connection details (for creating the global connection)
	SnowflakeAccount   string
	SnowflakeDatabase  string
	SnowflakeWarehouse string

	// Snowflake auth credentials (for the platform metadata credential)
	User     string
	Password string
	Role     string
}

// GetPlatformMetadataCredentialTestingConfigurations returns the configuration needed to test
// platform metadata credentials. Returns nil if required environment variables are not set.
// Required env vars:
//   - ACC_TEST_SNOWFLAKE_ACCOUNT: Snowflake account identifier
//   - ACC_TEST_SNOWFLAKE_DATABASE: Database name
//   - ACC_TEST_SNOWFLAKE_WAREHOUSE: Warehouse name
//   - ACC_TEST_SNOWFLAKE_USER: User for metadata credential auth
//   - ACC_TEST_SNOWFLAKE_PASSWORD: Password for metadata credential auth
//   - ACC_TEST_SNOWFLAKE_ROLE: Role for metadata credential auth
func GetPlatformMetadataCredentialTestingConfigurations() *PlatformMetadataCredentialConfig {
	account := os.Getenv("ACC_TEST_SNOWFLAKE_ACCOUNT")
	database := os.Getenv("ACC_TEST_SNOWFLAKE_DATABASE")
	warehouse := os.Getenv("ACC_TEST_SNOWFLAKE_WAREHOUSE")
	user := os.Getenv("ACC_TEST_SNOWFLAKE_USER")
	password := os.Getenv("ACC_TEST_SNOWFLAKE_PASSWORD")
	role := os.Getenv("ACC_TEST_SNOWFLAKE_ROLE")

	if account == "" || database == "" || warehouse == "" || user == "" || password == "" || role == "" {
		return nil
	}

	return &PlatformMetadataCredentialConfig{
		SnowflakeAccount:   account,
		SnowflakeDatabase:  database,
		SnowflakeWarehouse: warehouse,
		User:               user,
		Password:           password,
		Role:               role,
	}
}
