// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSubscriptionLocalizationResource(t *testing.T) {
	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	suffix := time.Now().Unix()
	productID := fmt.Sprintf("com.truetickets.test.subloc%d", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: testAccSubscriptionLocalizationResourceConfig(appID, suffix, productID, "Premium", "All premium features."),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("appleappstoreconnect_subscription_localization.test", "id"),
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription_localization.test", "locale", "en-US"),
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription_localization.test", "name", "Premium"),
				),
			},
			// Update in place.
			{
				Config: testAccSubscriptionLocalizationResourceConfig(appID, suffix, productID, "Premium Plus", "Everything included."),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription_localization.test", "name", "Premium Plus"),
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription_localization.test", "description", "Everything included."),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "appleappstoreconnect_subscription_localization.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSubscriptionLocalizationResourceConfig(appID string, suffix int64, productID, name, description string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_subscription_group" "test" {
  app_id         = %[1]q
  reference_name = "tf_test_subloc_group_%[2]d"
}

resource "appleappstoreconnect_subscription" "test" {
  subscription_group_id = appleappstoreconnect_subscription_group.test.id
  product_id            = %[3]q
  name                  = "Localization Test"
  subscription_period   = "ONE_MONTH"
}

resource "appleappstoreconnect_subscription_localization" "test" {
  subscription_id = appleappstoreconnect_subscription.test.id
  locale          = "en-US"
  name            = %[4]q
  description     = %[5]q
}
`, appID, suffix, productID, name, description)
}
