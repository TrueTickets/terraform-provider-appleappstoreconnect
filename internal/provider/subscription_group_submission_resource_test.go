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

// TestAccSubscriptionGroupSubmissionResource submits a subscription group for
// App Review, which is irreversible and cannot be cleaned up. It is gated
// behind an explicit opt-in beyond TF_ACC.
func TestAccSubscriptionGroupSubmissionResource(t *testing.T) {
	if os.Getenv("APP_STORE_CONNECT_TEST_ALLOW_SUBMISSION") == "" {
		t.Skip("set APP_STORE_CONNECT_TEST_ALLOW_SUBMISSION=1 to run the submission test (it submits a subscription group for review and cannot be undone)")
	}

	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	suffix := time.Now().Unix()
	productID := fmt.Sprintf("com.truetickets.test.subgsub%d", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriptionGroupSubmissionResourceConfig(appID, suffix, productID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("appleappstoreconnect_subscription_group_submission.test", "id"),
					resource.TestCheckResourceAttrPair(
						"appleappstoreconnect_subscription_group_submission.test", "subscription_group_id",
						"appleappstoreconnect_subscription_group.test", "id",
					),
				),
			},
		},
	})
}

func testAccSubscriptionGroupSubmissionResourceConfig(appID string, suffix int64, productID string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_subscription_group" "test" {
  app_id         = %[1]q
  reference_name = "tf_test_subgsub_group_%[2]d"
}

resource "appleappstoreconnect_subscription_group_localization" "test" {
  subscription_group_id = appleappstoreconnect_subscription_group.test.id
  locale                = "en-US"
  name                  = "Submission Test Group"
}

resource "appleappstoreconnect_subscription" "test" {
  subscription_group_id = appleappstoreconnect_subscription_group.test.id
  product_id            = %[3]q
  name                  = "Submission Test"
  subscription_period   = "ONE_MONTH"
}

resource "appleappstoreconnect_subscription_localization" "test" {
  subscription_id = appleappstoreconnect_subscription.test.id
  locale          = "en-US"
  name            = "Submission Test"
  description     = "Submission acceptance test subscription."
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

resource "appleappstoreconnect_subscription_group_submission" "test" {
  subscription_group_id = appleappstoreconnect_subscription_group.test.id

  depends_on = [
    appleappstoreconnect_subscription_group_localization.test,
    appleappstoreconnect_subscription_localization.test,
    appleappstoreconnect_subscription_price.test,
  ]
}
`, appID, suffix, productID)
}
