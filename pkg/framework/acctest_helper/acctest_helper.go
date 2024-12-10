package acctest_helper

import (
	"context"
	"fmt"
	"log"
	"net/http"
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
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
)

func SharedClient() (*dbt_cloud.Client, error) {

	accountIDString := os.Getenv("DBT_CLOUD_ACCOUNT_ID")
	accountID, _ := strconv.Atoi(accountIDString)
	token := os.Getenv("DBT_CLOUD_TOKEN")
	hostURL := os.Getenv("DBT_CLOUD_HOST_URL")

	if hostURL == "" {
		hostURL = "https://cloud.getdbt.com/api"
	}

	client := dbt_cloud.Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		HostURL:    hostURL,
		Token:      token,
		AccountID:  accountID,
	}

	return &client, nil
}

const (
	DBT_CLOUD_VERSION = "latest"
)

var TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"dbtcloud": func() (tfprotov6.ProviderServer, error) {
		upgradedSdkProvider, err := tf5to6server.UpgradeServer(
			context.Background(),
			provider.SDKProvider("test")().GRPCProvider,
		)
		if err != nil {
			log.Fatal(err)
		}
		providers := []func() tfprotov6.ProviderServer{
			func() tfprotov6.ProviderServer {
				return upgradedSdkProvider
			},
			providerserver.NewProtocol6(provider.New()),
		}

		return tf6muxserver.NewMuxServer(context.Background(), providers...)
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

func IsDbtCloudPR() bool {
	return os.Getenv("DBT_CLOUD_ACCOUNT_ID") == "1"
}

// GetDbtCloudUserId returns the user ID to use for acceptance tests.
// Currently, this utilizes some legacy logic to determine the user ID.
// If the DBT_CLOUD_USER_ID environment variable is fully adopted, this
// function can be simplified.
func GetDbtCloudUserId() int {
	if IsDbtCloudPR() {
		return 1
	} else if os.Getenv("CI") != "" {
		return 54461
	} else {
		id, err := strconv.Atoi(os.Getenv("DBT_CLOUD_USER_ID"))
		if err != nil {
			log.Fatalf("Unable to determine UserID for test: %v", err)
		}
		return id
	}
}

// GetDbtCloudUserEmail returns the user email to use for acceptance tests.
// Currently, this utilizes some legacy logic to determine the user email.
// If the DBT_CLOUD_USER_EMAIL environment variable is fully adopted, this
// function can be simplified.
func GetDbtCloudUserEmail() string {
	if IsDbtCloudPR() {
		return "d" + "ev@" + "db" + "tla" + "bs.c" + "om"
	} else if os.Getenv("CI") != "" {
		return "beno" + "it" + ".per" + "igaud" + "@" + "fisht" + "ownanalytics" + "." + "com"
	} else {
		email := os.Getenv("DBT_CLOUD_USER_EMAIL")
		if email == "" {
			log.Fatalf("Unable to determine GroupIds for test")
		}
		return email
	}
}

// GetDbtCloudGroupIds returns the group IDs to use for acceptance tests.
// Currently, this utilizes some legacy logic to determine the group IDs.
// If the DBT_CLOUD_GROUP_IDS environment variable is fully adopted, this
// function can be simplified.
func GetDbtCloudGroupIds() string {
	var groupIds string
	if IsDbtCloudPR() {
		groupIds = "1,2,3"
	} else if os.Getenv("CI") != "" {
		groupIds = "531585,531584,531583"
	} else {
		groupIds = os.Getenv("DBT_CLOUD_GROUP_IDS")
		if groupIds == "" {
			log.Fatalf("Unable to determine GroupIds for test")
		}
	}
	return fmt.Sprintf("[%s]", groupIds)
}

// GetGitHubRepoUrl returns the GitHub repository URL to use for acceptance tests.
// Currently, this utilizes some legacy logic to determine the GitHub repository URL.
// If the ACC_TEST_GITHUB_REPO_URL environment variable is fully adopted, this
// function can be simplified.
func GetGitHubRepoUrl() string {
	if IsDbtCloudPR() || os.Getenv("CI") != "" {
		return "git://github.com/dbt-labs/jaffle_shop.git"
	} else {
		url := os.Getenv("ACC_TEST_GITHUB_REPO_URL")
		if url == "" {
			log.Fatalf("Unable to determine GitHub repository url for test")
		}
		return url
	}
}

// GetGitHubAppInstallationId returns the GitHub app installation ID to use for acceptance tests.
// Currently, this utilizes some legacy logic to determine the GitHub app installation ID.
// If the ACC_TEST_GITHUB_APP_INSTALLATION_ID environment variable is fully adopted, this
// function can be simplified.
func GetGitHubAppInstallationId() int {
	if IsDbtCloudPR() || os.Getenv("CI") != "" {
		return 28374841
	} else {
		id, err := strconv.Atoi(os.Getenv("ACC_TEST_GITHUB_APP_INSTALLATION_ID"))
		if err != nil {
			log.Fatalf("Unable to determine GitHub app installation id for test: %v", err)
		}
		return id
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
