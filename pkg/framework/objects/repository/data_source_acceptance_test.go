package repository_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudRepositoryDataSource(t *testing.T) {
	randomProjectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
	repoUrl := "git@github.com:dbt-labs/terraform-provider-dbtcloud.git"

	config := repositoryDataSourceConfig(randomProjectName, repoUrl)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.dbtcloud_repository.test", "remote_url", repoUrl),
		resource.TestCheckResourceAttrSet("data.dbtcloud_repository.test", "repository_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_repository.test", "project_id"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_repository.test", "is_active"),
	)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  check,
			},
		},
	})
}

func repositoryDataSourceConfig(projectName, repositoryUrl string) string {
	return fmt.Sprintf(`
    resource "dbtcloud_project" "test_project" {
        name = "%s"
    }

    resource "dbtcloud_repository" "test_repository" {
        project_id = dbtcloud_project.test_project.id
        remote_url = "%s"
        is_active = true
        depends_on = [
            dbtcloud_project.test_project
        ]
    }

    data "dbtcloud_repository" "test" {
        project_id = dbtcloud_project.test_project.id
        repository_id = dbtcloud_repository.test_repository.repository_id
    }
    `, projectName, repositoryUrl)
}
