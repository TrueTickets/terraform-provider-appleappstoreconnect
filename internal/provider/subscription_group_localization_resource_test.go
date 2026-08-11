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

func TestAccSubscriptionGroupLocalizationResource(t *testing.T) {
	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	refName := fmt.Sprintf("tf_test_grouploc_%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: testAccSubscriptionGroupLocalizationResourceConfig(appID, refName, "Premium Club"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("appleappstoreconnect_subscription_group_localization.test", "id"),
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription_group_localization.test", "locale", "en-US"),
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription_group_localization.test", "name", "Premium Club"),
				),
			},
			// Update name in place.
			{
				Config: testAccSubscriptionGroupLocalizationResourceConfig(appID, refName, "Premium Club Plus"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appleappstoreconnect_subscription_group_localization.test", "name", "Premium Club Plus"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "appleappstoreconnect_subscription_group_localization.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSubscriptionGroupLocalizationResourceConfig(appID, refName, name string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_subscription_group" "test" {
  app_id         = %[1]q
  reference_name = %[2]q
}

resource "appleappstoreconnect_subscription_group_localization" "test" {
  subscription_group_id = appleappstoreconnect_subscription_group.test.id
  locale                = "en-US"
  name                  = %[3]q
}
`, appID, refName, name)
}
