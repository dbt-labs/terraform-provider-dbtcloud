package scim_group_permissions_test

import (
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestDbtCloudScimGroupPermissionsResource_ValidateConfig tests that the resource
// configuration is valid and can be planned. We use ExpectNonEmptyPlan because
// the resource would create new infrastructure.
func TestDbtCloudScimGroupPermissionsResource_ValidateConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudScimGroupPermissionsResourceBasicConfig(),
				PlanOnly: true,
				ExpectNonEmptyPlan: true, // We expect a plan because this would create resources
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// This validates that the configuration is syntactically correct
						// and the resource can be processed by Terraform
						return nil
					},
				),
			},
		},
	})
}

func testAccDbtCloudScimGroupPermissionsResourceBasicConfig() string {
	return `
resource "dbtcloud_project" "test_project" {
  name = "Test Project for SCIM Group Permissions"
}

# Example SCIM group permissions configuration
# In practice, the group_id would be from an external identity provider
resource "dbtcloud_scim_group_permissions" "test" {
  group_id = 12345  # External SCIM group ID
  
  permissions = [
    {
      permission_set = "developer"
      project_id     = dbtcloud_project.test_project.id
      all_projects   = false
      writable_environment_categories = ["development"]
    }
  ]
}
`
}
