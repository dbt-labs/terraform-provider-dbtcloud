package data_sources_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDbtCloudSnowflakeCredentialDataSource(t *testing.T) {

	randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	randomIDInt, _ := strconv.Atoi(randomID)

	config := fmt.Sprintf(`
			data "dbt_cloud_snowflake_credential" "test" {
				project_id = 123
				credential_id = %d
			}
		`, randomIDInt)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbt_cloud_snowflake_credential.test", "credential_id", randomID),
		resource.TestCheckResourceAttr("data.dbt_cloud_snowflake_credential.test", "project_id", "123"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_snowflake_credential.test", "auth_type"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_snowflake_credential.test", "is_active"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_snowflake_credential.test", "schema"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_snowflake_credential.test", "user"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_snowflake_credential.test", "password"),
		resource.TestCheckResourceAttrSet("data.dbt_cloud_snowflake_credential.test", "num_threads"),
	)

	resource.ParallelTest(t, resource.TestCase{
		Providers: providers(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  check,
			},
		},
	})
}
