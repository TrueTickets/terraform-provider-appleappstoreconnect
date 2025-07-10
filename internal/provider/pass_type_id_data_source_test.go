// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPassTypeIDDataSource(t *testing.T) {
	testIdentifier := fmt.Sprintf("pass.io.truetickets.test.datasource%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing using ID
			{
				Config: testAccPassTypeIDDataSourceConfigByID(testIdentifier),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_pass_type_id.test", "id"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_pass_type_id.test", "identifier", testIdentifier),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_pass_type_id.test", "description", "Test Pass Type"),
				),
			},
			// Read testing using filter
			{
				Config: testAccPassTypeIDDataSourceConfigByFilter(testIdentifier),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_pass_type_id.test", "id"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_pass_type_id.test", "identifier", testIdentifier),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_pass_type_id.test", "description", "Test Pass Type"),
				),
			},
		},
	})
}

func testAccPassTypeIDDataSourceConfigByID(identifier string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_pass_type_id" "test" {
  identifier  = %[1]q
  description = "Test Pass Type"
}

data "appleappstoreconnect_pass_type_id" "test" {
  id = appleappstoreconnect_pass_type_id.test.id
}
`, identifier)
}

func testAccPassTypeIDDataSourceConfigByFilter(identifier string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_pass_type_id" "test" {
  identifier  = %[1]q
  description = "Test Pass Type"
}

data "appleappstoreconnect_pass_type_id" "test" {
  filter = {
    identifier = appleappstoreconnect_pass_type_id.test.identifier
  }
}
`, identifier)
}
