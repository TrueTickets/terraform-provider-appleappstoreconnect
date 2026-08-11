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

func TestAccInAppPurchaseDataSource(t *testing.T) {
	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	productID := fmt.Sprintf("com.truetickets.test.iapds%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID.
			{
				Config: testAccInAppPurchaseDataSourceConfigByID(appID, productID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.appleappstoreconnect_in_app_purchase.test", "product_id", productID),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_in_app_purchase.test", "in_app_purchase_type", "NON_CONSUMABLE"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_in_app_purchase.test", "app_id", appID),
				),
			},
			// Read by filter (app_id + product_id).
			{
				Config: testAccInAppPurchaseDataSourceConfigByFilter(appID, productID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_in_app_purchase.test", "id"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_in_app_purchase.test", "product_id", productID),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_in_app_purchase.test", "app_id", appID),
				),
			},
		},
	})
}

func testAccInAppPurchaseDataSourceConfigByID(appID, productID string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_in_app_purchase" "test" {
  app_id               = %[1]q
  product_id           = %[2]q
  name                 = "Data Source Test"
  in_app_purchase_type = "NON_CONSUMABLE"
}

data "appleappstoreconnect_in_app_purchase" "test" {
  id = appleappstoreconnect_in_app_purchase.test.id
}
`, appID, productID)
}

func testAccInAppPurchaseDataSourceConfigByFilter(appID, productID string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_in_app_purchase" "test" {
  app_id               = %[1]q
  product_id           = %[2]q
  name                 = "Data Source Test"
  in_app_purchase_type = "NON_CONSUMABLE"
}

data "appleappstoreconnect_in_app_purchase" "test" {
  filter = {
    app_id     = %[1]q
    product_id = appleappstoreconnect_in_app_purchase.test.product_id
  }
}
`, appID, productID)
}
