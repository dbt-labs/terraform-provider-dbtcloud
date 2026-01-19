package account_features_test

import (
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudAccountFeaturesResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDbtCloudAccountFeaturesResourceBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_account_features.test",
						"advanced_ci",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_account_features.test",
						"partial_parsing",
						"false",
					),
				),
			},
			// Update testing
			{
				Config: testAccDbtCloudAccountFeaturesResourceFullConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_account_features.test",
						"advanced_ci",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_account_features.test",
						"partial_parsing",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_account_features.test",
						"repo_caching",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_account_features.test",
						"ai_features",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_account_features.test",
						"catalog_ingestion",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_account_features.test",
						"explorer_account_ui",
						"true",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_account_features.test",
						"fusion_migration_permissions",
						"false",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_account_features.test",
						"cost_insights",
						"true",
					),
				),
			},
		},
	})
}

func testAccDbtCloudAccountFeaturesResourceBasicConfig() string {
	return `
resource "dbtcloud_account_features" "test" {
    advanced_ci     = true
    partial_parsing = false
}
`
}

func testAccDbtCloudAccountFeaturesResourceFullConfig() string {
	return `
resource "dbtcloud_account_features" "test" {
    advanced_ci                  = true
    partial_parsing              = true
    repo_caching                 = true
    ai_features                  = true
    catalog_ingestion            = true
    explorer_account_ui          = true
    fusion_migration_permissions = false
    cost_insights                = true
}
`
}
