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

// TestAccInAppPurchaseSubmissionResource submits an In-App Purchase for App
// Review, which is an irreversible action that cannot be cleaned up. It is
// therefore gated behind an explicit opt-in beyond TF_ACC.
func TestAccInAppPurchaseSubmissionResource(t *testing.T) {
	if os.Getenv("APP_STORE_CONNECT_TEST_ALLOW_SUBMISSION") == "" {
		t.Skip("set APP_STORE_CONNECT_TEST_ALLOW_SUBMISSION=1 to run the submission test (it submits an In-App Purchase for review and cannot be undone)")
	}

	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	productID := fmt.Sprintf("com.truetickets.test.iapsub%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInAppPurchaseSubmissionResourceConfig(appID, productID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("appleappstoreconnect_in_app_purchase_submission.test", "id"),
					resource.TestCheckResourceAttrPair(
						"appleappstoreconnect_in_app_purchase_submission.test", "in_app_purchase_id",
						"appleappstoreconnect_in_app_purchase.test", "id",
					),
				),
			},
		},
	})
}

func testAccInAppPurchaseSubmissionResourceConfig(appID, productID string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_in_app_purchase" "test" {
  app_id               = %[1]q
  product_id           = %[2]q
  name                 = "Submission Test"
  in_app_purchase_type = "NON_CONSUMABLE"
}

resource "appleappstoreconnect_in_app_purchase_localization" "test" {
  in_app_purchase_id = appleappstoreconnect_in_app_purchase.test.id
  locale             = "en-US"
  name               = "Submission Test"
  description        = "Submission acceptance test product."
}

data "appleappstoreconnect_in_app_purchase_price_point" "usd_099" {
  in_app_purchase_id = appleappstoreconnect_in_app_purchase.test.id
  territory          = "USA"
  customer_price     = "0.99"
}

resource "appleappstoreconnect_in_app_purchase_price_schedule" "test" {
  in_app_purchase_id = appleappstoreconnect_in_app_purchase.test.id
  base_territory     = "USA"

  manual_prices = [
    {
      price_point_id = data.appleappstoreconnect_in_app_purchase_price_point.usd_099.id
      territory      = "USA"
    },
  ]
}

resource "appleappstoreconnect_in_app_purchase_submission" "test" {
  in_app_purchase_id = appleappstoreconnect_in_app_purchase.test.id

  depends_on = [
    appleappstoreconnect_in_app_purchase_localization.test,
    appleappstoreconnect_in_app_purchase_price_schedule.test,
  ]
}
`, appID, productID)
}
