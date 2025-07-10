// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCertificatesDataSource(t *testing.T) {
	t.Skip("Skipping due to Apple API server error when listing certificates without filter")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing without filter
			{
				Config: testAccCertificatesDataSourceConfigNoFilter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_certificates.test", "certificates.#"),
				),
			},
			// Read testing with certificate type filter
			{
				Config: testAccCertificatesDataSourceConfigWithTypeFilter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_certificates.test", "certificates.#"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_certificates.test", "filter.certificate_type", "PASS_TYPE_ID"),
				),
			},
			// Read testing with display name filter
			{
				Config: testAccCertificatesDataSourceConfigWithDisplayNameFilter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_certificates.test", "certificates.#"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_certificates.test", "filter.display_name", "Test"),
				),
			},
			// Read testing with both filters
			{
				Config: testAccCertificatesDataSourceConfigWithBothFilters(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_certificates.test", "certificates.#"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_certificates.test", "filter.certificate_type", "PASS_TYPE_ID"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_certificates.test", "filter.display_name", "Test"),
				),
			},
		},
	})
}

func testAccCertificatesDataSourceConfigNoFilter() string {
	return `
data "appleappstoreconnect_certificates" "test" {
}
`
}

func testAccCertificatesDataSourceConfigWithTypeFilter() string {
	return `
data "appleappstoreconnect_certificates" "test" {
  filter = {
    certificate_type = "PASS_TYPE_ID"
  }
}
`
}

func testAccCertificatesDataSourceConfigWithDisplayNameFilter() string {
	return `
data "appleappstoreconnect_certificates" "test" {
  filter = {
    display_name = "Test"
  }
}
`
}

func testAccCertificatesDataSourceConfigWithBothFilters() string {
	return `
data "appleappstoreconnect_certificates" "test" {
  filter = {
    certificate_type = "PASS_TYPE_ID"
    display_name     = "Test"
  }
}
`
}
