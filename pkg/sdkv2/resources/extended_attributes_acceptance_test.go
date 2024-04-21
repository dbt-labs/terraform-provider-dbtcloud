package resources_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDbtCloudExtendedAttributesResource(t *testing.T) {

	projectName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDbtCloudExtendedAttributesDestroy,
		Steps: []resource.TestStep{
			// CREATE
			{
				Config: testAccDbtCloudExtendedAttributesResourceConfig(projectName, "step1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudExtendedAttributesExists(
						"dbtcloud_extended_attributes.test_extended_attributes",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_extended_attributes.test_extended_attributes",
						"extended_attributes_id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_extended_attributes.test_extended_attributes",
						"project_id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_extended_attributes.test_extended_attributes",
						"extended_attributes",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_extended_attributes.test_extended_attributes",
						"extended_attributes",
						"{\"catalog\":\"dbt_catalog\",\"http_path\":\"/sql/your/http/path\",\"my_nested_field\":{\"subfield\":\"my_value\"},\"type\":\"databricks\"}",
					),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudExtendedAttributesResourceConfig(projectName, "step2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudExtendedAttributesExists(
						"dbtcloud_extended_attributes.test_extended_attributes",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_extended_attributes.test_extended_attributes",
						"extended_attributes_id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_extended_attributes.test_extended_attributes",
						"project_id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_extended_attributes.test_extended_attributes",
						"extended_attributes",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_extended_attributes.test_extended_attributes",
						"extended_attributes",
						"{\"catalog\":\"dbt_catalog_new\",\"type\":\"databricks\"}",
					),
				),
			},
			// REMOVE FROM ENVIRONMENT
			{
				Config: testAccDbtCloudExtendedAttributesResourceUnlinked(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudExtendedAttributesExists(
						"dbtcloud_extended_attributes.test_extended_attributes",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_extended_attributes.test_extended_attributes",
						"extended_attributes_id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_extended_attributes.test_extended_attributes",
						"project_id",
					),
					resource.TestCheckResourceAttrSet(
						"dbtcloud_extended_attributes.test_extended_attributes",
						"extended_attributes",
					),
					resource.TestCheckResourceAttr(
						"dbtcloud_extended_attributes.test_extended_attributes",
						"extended_attributes",
						"{\"catalog\":\"dbt_catalog_new\",\"type\":\"databricks\"}",
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_extended_attributes.test_extended_attributes",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccDbtCloudExtendedAttributesResourceConfig(projectName, step string) string {

	var extendedAttributes string

	if step == "step1" {
		extendedAttributes = `jsonencode(
			{
			  type      = "databricks"
			  catalog   = "dbt_catalog"
			  http_path = "/sql/your/http/path"
			  my_nested_field = {
				subfield = "my_value"
			  }
			}
		  )
		`
	} else if step == "step2" {
		// try the "Heredoc" syntax instead of the jsonencode function
		extendedAttributes = `<<EOF
		{
		  "catalog": "dbt_catalog_new",
		  "type": "databricks"
		}
		EOF`
	}

	return fmt.Sprintf(`
	resource "dbtcloud_project" "test_project" {
        name        = "%s"
    }

    resource "dbtcloud_environment" "test_environment" {
        project_id = dbtcloud_project.test_project.id
        name = "extended_attributes_test_env"
        dbt_version = "%s"
        type = "development"
		extended_attributes_id = dbtcloud_extended_attributes.test_extended_attributes.extended_attributes_id
    }

    resource "dbtcloud_extended_attributes" "test_extended_attributes" {
        extended_attributes = %s
        project_id = dbtcloud_project.test_project.id
      }

`, projectName, DBT_CLOUD_VERSION, extendedAttributes)
}

func testAccDbtCloudExtendedAttributesResourceUnlinked(projectName string) string {
	return fmt.Sprintf(`
	resource "dbtcloud_project" "test_project" {
        name        = "%s"
    }

    resource "dbtcloud_environment" "test_environment" {
        project_id = dbtcloud_project.test_project.id
        name = "extended_attributes_test_env"
        dbt_version = "%s"
        type = "development"
    }

    resource "dbtcloud_extended_attributes" "test_extended_attributes" {
        extended_attributes = jsonencode(
			{
			  type      = "databricks"
			  catalog   = "dbt_catalog_new"
			}
		  )
        project_id = dbtcloud_project.test_project.id
      }
`, projectName, DBT_CLOUD_VERSION)
}

func testAccCheckDbtCloudExtendedAttributesExists(resource string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}
		apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get projectID")
		}
		extendedAttributesID, err := strconv.Atoi(
			strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1],
		)
		if err != nil {
			return fmt.Errorf("Can't get extendedAttributesID")
		}

		_, err = apiClient.GetExtendedAttributes(projectId, extendedAttributesID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resource, err)
		}
		return nil
	}
}

func testAccCheckDbtCloudExtendedAttributesDestroy(s *terraform.State) error {
	apiClient := testAccProvider.Meta().(*dbt_cloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_extended_attributes" {
			continue
		}
		projectId, err := strconv.Atoi(strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[0])
		if err != nil {
			return fmt.Errorf("Can't get projectID")
		}
		extendedAttributesID, err := strconv.Atoi(
			strings.Split(rs.Primary.ID, dbt_cloud.ID_DELIMITER)[1],
		)
		if err != nil {
			return fmt.Errorf("Can't get extendedAttributesID")
		}

		_, err = apiClient.GetExtendedAttributes(projectId, extendedAttributesID)
		if err == nil {
			return fmt.Errorf("Extended attributes still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
