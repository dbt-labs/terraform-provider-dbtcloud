package oauth_configuration_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/acctest_helper"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDbtCloudOAuthConfigurationOktaResource(t *testing.T) {

	oAuthType := "okta"
	oAuthConfigurationName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	oauthClientId := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	oauthClientSecret := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	authorizeUrl := "https://" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + ".com"
	tokenUrl := "https://" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + ".com"
	redirectUri := "https://" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + ".com"

	oAuthConfigurationName2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	oauthClientId2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	oauthClientSecret2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	authorizeUrl2 := "https://" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + ".com"
	tokenUrl2 := "https://" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + ".com"
	redirectUri2 := "https://" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + ".com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudOAuthConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudOAuthConfigurationResourceBasicConfig(
					oAuthType,
					oAuthConfigurationName,
					oauthClientId,
					oauthClientSecret,
					authorizeUrl,
					tokenUrl,
					redirectUri,
					"",
				),
				// we just need to check the ones that have special logics or are computed
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_oauth_configuration.test_oauth_configuration",
						"client_secret",
						oauthClientSecret,
					),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudOAuthConfigurationResourceBasicConfig(
					oAuthType,
					oAuthConfigurationName2,
					oauthClientId2,
					oauthClientSecret2,
					authorizeUrl2,
					tokenUrl2,
					redirectUri2,
					"",
				),
				// we just need to check the ones that have special logics or are computed
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_oauth_configuration.test_oauth_configuration",
						"client_secret",
						oauthClientSecret2,
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_oauth_configuration.test_oauth_configuration",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_secret"},
			},
		},
	})
}

func TestAccDbtCloudOAuthConfigurationEntraResource(t *testing.T) {

	oAuthType := "entra"
	oAuthConfigurationName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	oauthClientId := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	oauthClientSecret := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	authorizeUrl := "https://" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + ".com"
	tokenUrl := "https://" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + ".com"
	redirectUri := "https://" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + ".com"
	applicationIdUri := "https://" + acctest.RandStringFromCharSet(
		10,
		acctest.CharSetAlpha,
	) + ".com"

	oAuthConfigurationName2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	oauthClientId2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	oauthClientSecret2 := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	authorizeUrl2 := "https://" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + ".com"
	tokenUrl2 := "https://" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + ".com"
	redirectUri2 := "https://" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + ".com"
	applicationIdUri2 := "https://" + acctest.RandStringFromCharSet(
		10,
		acctest.CharSetAlpha,
	) + ".com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest_helper.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest_helper.TestAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDbtCloudOAuthConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDbtCloudOAuthConfigurationResourceBasicConfig(
					oAuthType,
					oAuthConfigurationName,
					oauthClientId,
					oauthClientSecret,
					authorizeUrl,
					tokenUrl,
					redirectUri,
					applicationIdUri,
				),
				// we just need to check the ones that have special logics or are computed
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_oauth_configuration.test_oauth_configuration",
						"client_secret",
						oauthClientSecret,
					),
				),
			},
			// MODIFY
			{
				Config: testAccDbtCloudOAuthConfigurationResourceBasicConfig(
					oAuthType,
					oAuthConfigurationName2,
					oauthClientId2,
					oauthClientSecret2,
					authorizeUrl2,
					tokenUrl2,
					redirectUri2,
					applicationIdUri2,
				),
				// we just need to check the ones that have special logics or are computed
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"dbtcloud_oauth_configuration.test_oauth_configuration",
						"client_secret",
						oauthClientSecret2,
					),
				),
			},
			// IMPORT
			{
				ResourceName:            "dbtcloud_oauth_configuration.test_oauth_configuration",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"client_secret"},
			},
		},
	})
}

func testAccDbtCloudOAuthConfigurationResourceBasicConfig(
	oAuthType,
	oAuthConfigurationName,
	oauth_client_id,
	oauth_client_secret,
	authorize_url,
	token_url,
	redirect_uri,
	application_id_uri string,
) string {

	application_id_uri_config := ""
	if oAuthType == "entra" {
		application_id_uri_config = fmt.Sprintf(`
		application_id_uri = "%s"
		`, application_id_uri)
	}

	return fmt.Sprintf(`
resource "dbtcloud_oauth_configuration" "test_oauth_configuration" {
    type = "%s"
    name = "%s"
	client_id = "%s"
	client_secret = "%s"
	authorize_url = "%s"
	token_url = "%s"
	redirect_uri = "%s"
	%s
}
	`, oAuthType,
		oAuthConfigurationName,
		oauth_client_id,
		oauth_client_secret,
		authorize_url,
		token_url,
		redirect_uri,
		application_id_uri_config)
}

func testAccCheckDbtCloudOAuthConfigurationDestroy(s *terraform.State) error {
	apiClient, err := acctest_helper.SharedClient()
	if err != nil {
		return fmt.Errorf("Issue getting the client")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "dbtcloud_oauth_configuration" {
			continue
		}
		oAuthConfigurationID, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Can't get oAuthConfigurationID")
		}
		_, err = apiClient.GetOAuthConfiguration(int64(oAuthConfigurationID))
		if err == nil {
			return fmt.Errorf("OAuthConfiguration still exists")
		}
		notFoundErr := "resource-not-found"
		expectedErr := regexp.MustCompile(notFoundErr)
		if !expectedErr.Match([]byte(err.Error())) {
			return fmt.Errorf("expected %s, got %s", notFoundErr, err)
		}
	}

	return nil
}
