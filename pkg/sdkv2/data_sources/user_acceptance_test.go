package data_sources_test

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
// )

// func TestAccDbtCloudUserDataSource(t *testing.T) {

// 	userEmail := "test@email.com"

// 	config := user(userEmail)

// 	check := resource.ComposeAggregateTestCheckFunc(
// 		resource.TestCheckResourceAttr("data.dbtcloud_user.test_user_read", "email", userEmail),
// 		resource.TestCheckResourceAttrSet("data.dbtcloud_user.test_user_read", "id"),
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
// data "dbtcloud_user" "test_user_read" {
//     email = "%s"
// }
// `, userEmail)
// }
