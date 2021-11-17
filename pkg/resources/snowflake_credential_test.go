package resources_test

//
// import (
// 	"fmt"
// 	"testing"
//
// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
// )
//
// func TestDbtCloudSnowflakeCredentialResource(t *testing.T) {
//
// 	config := fmt.Sprintf(`
// 			resource "dbt_cloud_snowflake_credential" "test" {
// 				is_active = true
// 				project_id = 123
// 				auth_type = "password"
// 				schema = "moo"
// 				user = "test_user"
// 				password = "test-password"
// 				num_threads = 3
// 			}
// 		`)
//
// 	check := resource.ComposeAggregateTestCheckFunc(
// 		resource.TestCheckResourceAttrSet("dbt_cloud_snowflake_credential.test", "credential_id"),
// 		resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test", "is_active", "true"),
// 		resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test", "project_id", "123"),
// 		resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test", "auth_type", "password"),
// 		resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test", "schema", "moo"),
// 		resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test", "user", "test_user"),
// 		resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test", "password", "test-password"),
// 		resource.TestCheckResourceAttr("dbt_cloud_snowflake_credential.test", "num_threads", "3"),
// 	)
//
// 	resource.ParallelTest(t, resource.TestCase{
// 		Providers: providers(),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: config,
// 				Check:  check,
// 			},
// 		},
// 	})
// }
