package partial_notification_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_config"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Requires an account with the account-level Slack app linked, so it only runs when ACC_TEST_DBT_CLOUD_SLACK_V2 is set.
func TestAccDbtCloudPartialNotificationResourceSlackV2(t *testing.T) {
	if os.Getenv("ACC_TEST_DBT_CLOUD_SLACK_V2") == "" {
		t.Skip("Skipping Slack V2 partial notification test: ACC_TEST_DBT_CLOUD_SLACK_V2 is not set")
	}

	userID := acctest_config.AcceptanceTestConfig.DbtCloudUserId
	projectName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	suffix := strings.ToUpper(acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	channelA := "C0A" + suffix
	channelB := "C0B" + suffix

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudPartialNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudPartialNotificationSlackV2Config(projectName, userID, channelA, channelB, "channel-b", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudPartialNotificationSlackChannel("dbtcloud_partial_notification.slack_a", channelA, "channel-a"),
				),
			},
			// a second channel must get its own notification and leave the first one untouched
			{
				Config: testAccDbtCloudPartialNotificationSlackV2Config(projectName, userID, channelA, channelB, "channel-b", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudPartialNotificationSlackChannel("dbtcloud_partial_notification.slack_a", channelA, "channel-a"),
					testAccCheckDbtCloudPartialNotificationSlackChannel("dbtcloud_partial_notification.slack_b", channelB, "channel-b"),
					testAccCheckDbtCloudPartialNotificationDifferentIDs(
						"dbtcloud_partial_notification.slack_a",
						"dbtcloud_partial_notification.slack_b",
					),
				),
			},
			// renaming the channel updates the notification in place instead of replacing it
			{
				Config: testAccDbtCloudPartialNotificationSlackV2Config(projectName, userID, channelA, channelB, "channel-b-renamed", true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dbtcloud_partial_notification.slack_b", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbtCloudPartialNotificationSlackChannel("dbtcloud_partial_notification.slack_b", channelB, "channel-b-renamed"),
				),
			},
		},
	})
}

func testAccDbtCloudPartialNotificationSlackV2Config(
	projectName string,
	userID int,
	channelA string,
	channelB string,
	channelBName string,
	withB bool,
) string {
	config := fmt.Sprintf(`
resource "dbtcloud_partial_notification" "slack_a" {
	user_id            = %d
	notification_type  = 5
	slack_channel_id   = "%s"
	slack_channel_name = "channel-a"
	on_failure         = [dbtcloud_job.test_notification_job_1.id]
}
`, userID, channelA)
	if withB {
		config += fmt.Sprintf(`
resource "dbtcloud_partial_notification" "slack_b" {
	user_id            = %d
	notification_type  = 5
	slack_channel_id   = "%s"
	slack_channel_name = "%s"
	on_failure         = [dbtcloud_job.test_notification_job_2.id]
	depends_on         = [dbtcloud_partial_notification.slack_a]
}
`, userID, channelB, channelBName)
	}
	return testAccDbtCloudPartialNotificationResourceBasicConfig(projectName) + "\n" + config
}

func testAccCheckDbtCloudPartialNotificationSlackChannel(
	resourceName string,
	channelID string,
	channelName string,
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		apiClient, err := acctest_helper.SharedClient()
		if err != nil {
			return fmt.Errorf("Issue getting the client")
		}
		notification, err := apiClient.GetNotification(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching item with resource %s. %s", resourceName, err)
		}
		if notification.SlackChannelID == nil || *notification.SlackChannelID != channelID {
			return fmt.Errorf("%s: expected slack_channel_id %s, got %v", resourceName, channelID, notification.SlackChannelID)
		}
		if notification.SlackChannelName == nil || *notification.SlackChannelName != channelName {
			return fmt.Errorf("%s: expected slack_channel_name %s, got %v", resourceName, channelName, notification.SlackChannelName)
		}
		return nil
	}
}

func testAccCheckDbtCloudPartialNotificationDifferentIDs(first string, second string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		a, okA := state.RootModule().Resources[first]
		b, okB := state.RootModule().Resources[second]
		if !okA || !okB {
			return fmt.Errorf("Not found: %s or %s", first, second)
		}
		if a.Primary.ID == b.Primary.ID {
			return fmt.Errorf("%s and %s share notification %s", first, second, a.Primary.ID)
		}
		return nil
	}
}
