// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBundleIDDataSource(t *testing.T) {
	testIdentifier := fmt.Sprintf("io.truetickets.test.bundleds%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing using ID.
			{
				Config: testAccBundleIDDataSourceConfigByID(testIdentifier),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_bundle_id.test", "id"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_bundle_id.test", "identifier", testIdentifier),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_bundle_id.test", "name", "Test Bundle ID"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_bundle_id.test", "platform", "IOS"),
				),
			},
			// Read testing using filter.
			{
				Config: testAccBundleIDDataSourceConfigByFilter(testIdentifier),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_bundle_id.test", "id"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_bundle_id.test", "identifier", testIdentifier),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_bundle_id.test", "name", "Test Bundle ID"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_bundle_id.test", "platform", "IOS"),
				),
			},
		},
	})
}

func testAccBundleIDDataSourceConfigByID(identifier string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_bundle_id" "test" {
  identifier = %[1]q
  name       = "Test Bundle ID"
  platform   = "IOS"
}

data "appleappstoreconnect_bundle_id" "test" {
  id = appleappstoreconnect_bundle_id.test.id
}
`, identifier)
}

func testAccBundleIDDataSourceConfigByFilter(identifier string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_bundle_id" "test" {
  identifier = %[1]q
  name       = "Test Bundle ID"
  platform   = "IOS"
}

data "appleappstoreconnect_bundle_id" "test" {
  filter = {
    identifier = appleappstoreconnect_bundle_id.test.identifier
  }
}
`, identifier)
}
