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

func TestAccInAppPurchaseResource(t *testing.T) {
	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	productID := fmt.Sprintf("com.truetickets.test.iap%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: testAccInAppPurchaseResourceConfig(appID, productID, "Test IAP"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("appleappstoreconnect_in_app_purchase.test", "id"),
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase.test", "product_id", productID),
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase.test", "name", "Test IAP"),
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase.test", "in_app_purchase_type", "NON_CONSUMABLE"),
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase.test", "app_id", appID),
					resource.TestCheckResourceAttrSet("appleappstoreconnect_in_app_purchase.test", "state"),
				),
			},
			// Update the reference name in place.
			{
				Config: testAccInAppPurchaseResourceConfig(appID, productID, "Test IAP Updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase.test", "name", "Test IAP Updated"),
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase.test", "product_id", productID),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "appleappstoreconnect_in_app_purchase.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccInAppPurchaseResourceConfig(appID, productID, name string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_in_app_purchase" "test" {
  app_id               = %[1]q
  product_id           = %[2]q
  name                 = %[3]q
  in_app_purchase_type = "NON_CONSUMABLE"
}
`, appID, productID, name)
}
