// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTerritoriesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTerritoriesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// The full territory list is well over 150 entries; assert a
					// representative subset is present and populated.
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_territories.test", "territories.#"),
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_territories.test", "territories.0.id"),
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_territories.test", "territories.0.currency"),
				),
			},
		},
	})
}

func testAccTerritoriesDataSourceConfig() string {
	return `
data "appleappstoreconnect_territories" "test" {}
`
}
