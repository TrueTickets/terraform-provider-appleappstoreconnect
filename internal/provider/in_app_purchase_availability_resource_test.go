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

func TestAccInAppPurchaseAvailabilityResource(t *testing.T) {
	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	productID := fmt.Sprintf("com.truetickets.test.iapavail%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInAppPurchaseAvailabilityResourceConfig(appID, productID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("appleappstoreconnect_in_app_purchase_availability.test", "id"),
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase_availability.test", "available_in_new_territories", "true"),
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase_availability.test", "available_territories.#", "2"),
				),
			},
			// ImportState testing (territories are reconstructed on read).
			{
				ResourceName:      "appleappstoreconnect_in_app_purchase_availability.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccInAppPurchaseAvailabilityResourceConfig(appID, productID string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_in_app_purchase" "test" {
  app_id               = %[1]q
  product_id           = %[2]q
  name                 = "Availability Test"
  in_app_purchase_type = "NON_CONSUMABLE"
}

resource "appleappstoreconnect_in_app_purchase_availability" "test" {
  in_app_purchase_id           = appleappstoreconnect_in_app_purchase.test.id
  available_in_new_territories = true
  available_territories        = ["USA", "FRA"]
}
`, appID, productID)
}
