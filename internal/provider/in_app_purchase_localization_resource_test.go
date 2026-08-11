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

func TestAccInAppPurchaseLocalizationResource(t *testing.T) {
	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	productID := fmt.Sprintf("com.truetickets.test.iaploc%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: testAccInAppPurchaseLocalizationResourceConfig(appID, productID, "Premium", "Unlock all features."),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("appleappstoreconnect_in_app_purchase_localization.test", "id"),
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase_localization.test", "locale", "en-US"),
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase_localization.test", "name", "Premium"),
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase_localization.test", "description", "Unlock all features."),
					resource.TestCheckResourceAttrPair(
						"appleappstoreconnect_in_app_purchase_localization.test", "in_app_purchase_id",
						"appleappstoreconnect_in_app_purchase.test", "id",
					),
				),
			},
			// Update name and description in place.
			{
				Config: testAccInAppPurchaseLocalizationResourceConfig(appID, productID, "Premium Plus", "Unlock everything."),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase_localization.test", "name", "Premium Plus"),
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase_localization.test", "description", "Unlock everything."),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "appleappstoreconnect_in_app_purchase_localization.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccInAppPurchaseLocalizationResourceConfig(appID, productID, name, description string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_in_app_purchase" "test" {
  app_id               = %[1]q
  product_id           = %[2]q
  name                 = "Localization Test"
  in_app_purchase_type = "NON_CONSUMABLE"
}

resource "appleappstoreconnect_in_app_purchase_localization" "test" {
  in_app_purchase_id = appleappstoreconnect_in_app_purchase.test.id
  locale             = "en-US"
  name               = %[3]q
  description        = %[4]q
}
`, appID, productID, name, description)
}
