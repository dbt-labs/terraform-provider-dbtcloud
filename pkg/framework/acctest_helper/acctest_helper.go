package acctest_helper

import (
	"context"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
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

	client, err := dbt_cloud.NewClient(
		&accountID,
		&token,
		&hostURL,
	)

	if err != nil {
		return client, err
	}

	return client, nil
}

const (
	DBT_CLOUD_VERSION = "versionless"
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
