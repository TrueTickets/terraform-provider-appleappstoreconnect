// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSubscriptionPriceResource(t *testing.T) {
	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	suffix := time.Now().Unix()
	productID := fmt.Sprintf("com.truetickets.test.subprice%d", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionPriceResourceConfig(appID, suffix, productID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("appleappstoreconnect_subscription_price.test", "id"),
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription_price.test", "territory", "USA"),
					resource.TestCheckResourceAttrPair(
						"appleappstoreconnect_subscription_price.test", "subscription_price_point_id",
						"data.appleappstoreconnect_subscription_price_point.usd_499", "id",
					),
				),
			},
			// ImportState testing with composite "subscription_id:price_id".
			{
				ResourceName:      "appleappstoreconnect_subscription_price.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources["appleappstoreconnect_subscription_price.test"]
					if rs == nil {
						return "", fmt.Errorf("resource not found in state")
					}
					return fmt.Sprintf("%s:%s", rs.Primary.Attributes["subscription_id"], rs.Primary.Attributes["id"]), nil
				},
			},
		},
	})
}

func testAccSubscriptionPriceResourceConfig(appID string, suffix int64, productID string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_subscription_group" "test" {
  app_id         = %[1]q
  reference_name = "tf_test_subprice_group_%[2]d"
}

resource "appleappstoreconnect_subscription" "test" {
  subscription_group_id = appleappstoreconnect_subscription_group.test.id
  product_id            = %[3]q
  name                  = "Price Test"
  subscription_period   = "ONE_MONTH"
}

data "appleappstoreconnect_subscription_price_point" "usd_499" {
  subscription_id = appleappstoreconnect_subscription.test.id
  territory       = "USA"
  customer_price  = "4.99"
}

resource "appleappstoreconnect_subscription_price" "test" {
  subscription_id             = appleappstoreconnect_subscription.test.id
  subscription_price_point_id = data.appleappstoreconnect_subscription_price_point.usd_499.id
  territory                   = "USA"
}
`, appID, suffix, productID)
}
