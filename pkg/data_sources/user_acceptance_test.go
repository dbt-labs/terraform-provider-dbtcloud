package data_sources_test

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
// )

// func TestAccDbtCloudUserDataSource(t *testing.T) {

// 	userEmail := "test@email.com"

// 	config := user(userEmail)

// 	check := resource.ComposeAggregateTestCheckFunc(
// 		resource.TestCheckResourceAttr("data.dbt_cloud_user.test_user_read", "email", userEmail),
// 		resource.TestCheckResourceAttrSet("data.dbt_cloud_user.test_user_read", "id"),
// 	)

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

// func user(userEmail string) string {
// 	return fmt.Sprintf(`
// data "dbt_cloud_user" "test_user_read" {
//     email = "%s"
// }
// `, userEmail)
// }
