// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPassTypeIDDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing using ID
			{
				Config: testAccPassTypeIDDataSourceConfigByID(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_pass_type_id.test", "id"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_pass_type_id.test", "identifier", "pass.io.truetickets.test.test"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_pass_type_id.test", "description", "Test Pass Type"),
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_pass_type_id.test", "created_date"),
				),
			},
			// Read testing using filter
			{
				Config: testAccPassTypeIDDataSourceConfigByFilter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_pass_type_id.test", "id"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_pass_type_id.test", "identifier", "pass.io.truetickets.test.test"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_pass_type_id.test", "description", "Test Pass Type"),
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_pass_type_id.test", "created_date"),
				),
			},
		},
	})
}

func testAccPassTypeIDDataSourceConfigByID() string {
	return `
resource "appleappstoreconnect_pass_type_id" "test" {
  identifier  = "pass.io.truetickets.test.test"
  description = "Test Pass Type"
}

data "appleappstoreconnect_pass_type_id" "test" {
  id = appleappstoreconnect_pass_type_id.test.id
}
`
}

func testAccPassTypeIDDataSourceConfigByFilter() string {
	return `
resource "appleappstoreconnect_pass_type_id" "test" {
  identifier  = "pass.io.truetickets.test.test"
  description = "Test Pass Type"
}

data "appleappstoreconnect_pass_type_id" "test" {
  filter {
    identifier = appleappstoreconnect_pass_type_id.test.identifier
  }
}
`
}
