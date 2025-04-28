package runs_test

import (
	"fmt"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDbtCloudUsersDataSource(t *testing.T) {

	environmentId, _, _ := acctest_helper.GetSemanticLayerConfigTestingConfigurations()
	if environmentId == 0 {
		t.Skip("Skipping test because no environment ID is set")
	}

	config := runs(environmentId)

	check := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet("data.dbtcloud_runs.all", "runs.0.id"),
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

func runs(envId int) string {
	return fmt.Sprintf(`

data "dbtcloud_runs" "all" {
  filter = {
    	environment_id = %d
	}
}
`, envId)
}
