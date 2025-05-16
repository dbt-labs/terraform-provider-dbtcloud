package data_sources_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudPostgresCredentialDataSource(t *testing.T) {

	randomProjectName := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	config := postgres_credential(randomProjectName, "moo", "baa", "maa", 64)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet(
			"data.dbtcloud_postgres_credential.test",
			"credential_id",
		),
		resource.TestCheckResourceAttrSet("data.dbtcloud_postgres_credential.test", "project_id"),
		resource.TestCheckResourceAttrSet(
			"data.dbtcloud_postgres_credential.test",
			"default_schema",
		),
		resource.TestCheckResourceAttrSet("data.dbtcloud_postgres_credential.test", "username"),
		resource.TestCheckResourceAttrSet("data.dbtcloud_postgres_credential.test", "num_threads"),
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

func postgres_credential(
	projectName string,
	defaultSchema string,
	username string,
	password string,
	numThreads int,
) string {
	return fmt.Sprintf(`
    resource "dbtcloud_project" "test_credential_project" {
        name = "%s"
    }

    resource "dbtcloud_postgres_credential" "test_cred" {
        project_id = dbtcloud_project.test_credential_project.id
        num_threads = 64
		type = "postgres"
        username = "baa"
        password = "maa"
        default_schema = "moo"
    }

    data "dbtcloud_postgres_credential" "test" {
        project_id = dbtcloud_project.test_credential_project.id
        credential_id = dbtcloud_postgres_credential.test_cred.credential_id
    }
    `, projectName)
}
