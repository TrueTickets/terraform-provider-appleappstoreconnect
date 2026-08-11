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

func TestAccSubscriptionResource(t *testing.T) {
	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	suffix := time.Now().Unix()
	productID := fmt.Sprintf("com.truetickets.test.sub%d", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: testAccSubscriptionResourceConfig(appID, suffix, productID, "Premium Monthly"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("appleappstoreconnect_subscription.test", "id"),
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription.test", "product_id", productID),
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription.test", "name", "Premium Monthly"),
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription.test", "subscription_period", "ONE_MONTH"),
					resource.TestCheckResourceAttrPair(
						"appleappstoreconnect_subscription.test", "subscription_group_id",
						"appleappstoreconnect_subscription_group.test", "id",
					),
				),
			},
			// Update name in place.
			{
				Config: testAccSubscriptionResourceConfig(appID, suffix, productID, "Premium Monthly v2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription.test", "name", "Premium Monthly v2"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "appleappstoreconnect_subscription.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSubscriptionResourceConfig(appID string, suffix int64, productID, name string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_subscription_group" "test" {
  app_id         = %[1]q
  reference_name = "tf_test_sub_group_%[2]d"
}

resource "appleappstoreconnect_subscription" "test" {
  subscription_group_id = appleappstoreconnect_subscription_group.test.id
  product_id            = %[3]q
  name                  = %[4]q
  subscription_period   = "ONE_MONTH"
}
`, appID, suffix, productID, name)
}
