package ip_restrictions_rule_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudIPRestrictionsRuleResource(t *testing.T) {
	ruleName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	ruleName2 := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create basic IP restrictions rule
			{
				Config: testAccDbtCloudIPRestrictionsRuleResourceBasicConfig(ruleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_ip_restrictions_rule.test",
						"name",
						ruleName,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_ip_restrictions_rule.test",
						"type",
						"allow",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_ip_restrictions_rule.test",
						"description",
						"Test IP restriction rule",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_ip_restrictions_rule.test",
						"cidrs.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_ip_restrictions_rule.test",
						"rule_set_enabled",
						"false",
					),
				),
			},
			// Update rule name and description
			{
				Config: testAccDbtCloudIPRestrictionsRuleResourceModifiedConfig(ruleName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_ip_restrictions_rule.test",
						"name",
						ruleName2,
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_ip_restrictions_rule.test",
						"description",
						"Modified test IP restriction rule",
					),
				),
			},
			// Add more CIDRs
			{
				Config: testAccDbtCloudIPRestrictionsRuleResourceMoreCIDRsConfig(ruleName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_ip_restrictions_rule.test",
						"cidrs.#",
						"4",
					),
				),
			},
			// Remove CIDRs and change type to deny
			{
				Config: testAccDbtCloudIPRestrictionsRuleResourceLessCIDRsConfig(ruleName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_ip_restrictions_rule.test",
						"cidrs.#",
						"1",
					),
				),
			},
			// Import test
			{
				ResourceName:      "dbtcloud_ip_restrictions_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDbtCloudIPRestrictionsRuleResourceBasicConfig(name string) string {
	return fmt.Sprintf(`
resource "dbtcloud_ip_restrictions_rule" "test" {
    name            = "%s"
    type            = "allow"
    description     = "Test IP restriction rule"
    rule_set_enabled = false
    
    cidrs  = [
		{
			cidr = "10.0.0.0/24"
		},  
		{
			cidr = "192.168.1.0/24"
		}
	]
}
`, name)
}

func testAccDbtCloudIPRestrictionsRuleResourceModifiedConfig(name string) string {
	return fmt.Sprintf(`
resource "dbtcloud_ip_restrictions_rule" "test" {
    name            = "%s"
    type            = "allow"
    description     = "Modified test IP restriction rule"
    rule_set_enabled = false
    
    cidrs  = [
		{
			cidr = "10.0.0.0/24"
		},  
		{
			cidr = "192.168.1.0/24"
		}
	]
}
`, name)
}

func testAccDbtCloudIPRestrictionsRuleResourceMoreCIDRsConfig(name string) string {
	return fmt.Sprintf(`
resource "dbtcloud_ip_restrictions_rule" "test" {
    name            = "%s"
    type            = "allow"
    description     = "Modified test IP restriction rule"
    rule_set_enabled = false
    
	cidrs  = [
		{
			cidr = "10.0.0.0/24"
		},  
		{
			cidr = "192.168.1.0/24"
		},  
		{
			cidr = "72.16.0.0/24"
		},  
		{
			cidr = "192.168.2.0/24"
		}
	]
}
`, name)
}

func testAccDbtCloudIPRestrictionsRuleResourceLessCIDRsConfig(name string) string {
	return fmt.Sprintf(`
resource "dbtcloud_ip_restrictions_rule" "test" {
    name            = "%s"
    type            = "deny"
    description     = "Modified test IP restriction rule" 
    rule_set_enabled = false
    
    cidrs = [{
        cidr = "10.0.0.0/24"
    }]
}
`, name)
}
