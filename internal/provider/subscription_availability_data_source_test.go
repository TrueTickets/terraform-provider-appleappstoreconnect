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

func TestAccSubscriptionAvailabilityDataSource(t *testing.T) {
	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	suffix := time.Now().Unix()
	productID := fmt.Sprintf("com.truetickets.test.subavail%d", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionAvailabilityDataSourceConfig(appID, suffix, productID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_subscription_availability.test", "id"),
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_subscription_availability.test", "available_in_new_territories"),
				),
			},
		},
	})
}

func testAccSubscriptionAvailabilityDataSourceConfig(appID string, suffix int64, productID string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_subscription_group" "test" {
  app_id         = %[1]q
  reference_name = "tf_test_subavail_group_%[2]d"
}

resource "appleappstoreconnect_subscription" "test" {
  subscription_group_id = appleappstoreconnect_subscription_group.test.id
  product_id            = %[3]q
  name                  = "Availability Test"
  subscription_period   = "ONE_MONTH"
}

data "appleappstoreconnect_subscription_availability" "test" {
  subscription_id = appleappstoreconnect_subscription.test.id
}
`, appID, suffix, productID)
}
