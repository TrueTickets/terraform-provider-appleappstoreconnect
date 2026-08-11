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

func TestAccSubscriptionDataSource(t *testing.T) {
	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	suffix := time.Now().Unix()
	productID := fmt.Sprintf("com.truetickets.test.subds%d", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID.
			{
				Config: testAccSubscriptionDataSourceConfigByID(appID, suffix, productID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.appleappstoreconnect_subscription.test", "product_id", productID),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_subscription.test", "subscription_period", "ONE_MONTH"),
				),
			},
			// Read by filter (group + product ID).
			{
				Config: testAccSubscriptionDataSourceConfigByFilter(appID, suffix, productID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_subscription.test", "id"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_subscription.test", "product_id", productID),
				),
			},
		},
	})
}

func testAccSubscriptionDataSourceConfigByID(appID string, suffix int64, productID string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_subscription_group" "test" {
  app_id         = %[1]q
  reference_name = "tf_test_subds_group_%[2]d"
}

resource "appleappstoreconnect_subscription" "test" {
  subscription_group_id = appleappstoreconnect_subscription_group.test.id
  product_id            = %[3]q
  name                  = "Data Source Test"
  subscription_period   = "ONE_MONTH"
}

data "appleappstoreconnect_subscription" "test" {
  id = appleappstoreconnect_subscription.test.id
}
`, appID, suffix, productID)
}

func testAccSubscriptionDataSourceConfigByFilter(appID string, suffix int64, productID string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_subscription_group" "test" {
  app_id         = %[1]q
  reference_name = "tf_test_subds_group_%[2]d"
}

resource "appleappstoreconnect_subscription" "test" {
  subscription_group_id = appleappstoreconnect_subscription_group.test.id
  product_id            = %[3]q
  name                  = "Data Source Test"
  subscription_period   = "ONE_MONTH"
}

data "appleappstoreconnect_subscription" "test" {
  filter = {
    subscription_group_id = appleappstoreconnect_subscription_group.test.id
    product_id            = appleappstoreconnect_subscription.test.product_id
  }
}
`, appID, suffix, productID)
}
