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

func TestAccSubscriptionGroupResource(t *testing.T) {
	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	refName := fmt.Sprintf("tf_test_group_%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: testAccSubscriptionGroupResourceConfig(appID, refName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("appleappstoreconnect_subscription_group.test", "id"),
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription_group.test", "reference_name", refName),
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription_group.test", "app_id", appID),
				),
			},
			// Update reference name in place.
			{
				Config: testAccSubscriptionGroupResourceConfig(appID, refName+"_updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription_group.test", "reference_name", refName+"_updated"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "appleappstoreconnect_subscription_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSubscriptionGroupResourceConfig(appID, refName string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_subscription_group" "test" {
  app_id         = %[1]q
  reference_name = %[2]q
}
`, appID, refName)
}
