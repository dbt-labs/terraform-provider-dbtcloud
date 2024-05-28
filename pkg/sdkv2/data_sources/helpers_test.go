package data_sources_test

import (
	"os"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	DBT_CLOUD_VERSION = "1.6.0-latest"
)

func providers() map[string]*schema.Provider {
	p := provider.SDKProvider("test")()
	return map[string]*schema.Provider{
		"dbtcloud": p,
	}
}

func isDbtCloudPR() bool {
	return os.Getenv("DBT_CLOUD_ACCOUNT_ID") == "1"
}
