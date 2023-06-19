package resources_test

import (
	"os"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	DBT_CLOUD_VERSION = "1.0.0"
)

func providers() map[string]*schema.Provider {
	p := provider.Provider()
	return map[string]*schema.Provider{
		"dbtcloud": p,
	}
}

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = provider.Provider()
	testAccProviders = map[string]*schema.Provider{
		"dbtcloud": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("DBT_CLOUD_ACCOUNT_ID"); v == "" {
		t.Fatal("DBT_CLOUD_ACCOUNT_ID must be set for acceptance tests")
	}
	if v := os.Getenv("DBT_CLOUD_TOKEN"); v == "" {
		t.Fatal("DBT_CLOUD_TOKEN must be set for acceptance tests")
	}
}
