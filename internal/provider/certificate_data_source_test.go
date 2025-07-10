// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCertificateDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing using ID
			{
				Config: testAccCertificateDataSourceConfigByID(time.Now().Unix()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_certificate.test", "id"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_certificate.test", "certificate_type", "PASS_TYPE_ID"),
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_certificate.test", "certificate_content"),
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_certificate.test", "serial_number"),
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_certificate.test", "expiration_date"),
				),
			},
			// Read testing using filter
			{
				Config: testAccCertificateDataSourceConfigByFilter(time.Now().Unix()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_certificate.test", "id"),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_certificate.test", "certificate_type", "PASS_TYPE_ID"),
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_certificate.test", "certificate_content"),
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_certificate.test", "serial_number"),
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_certificate.test", "expiration_date"),
				),
			},
		},
	})
}

func testAccCertificateDataSourceConfigByID(timestamp int64) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_pass_type_id" "test" {
  identifier  = "pass.io.truetickets.test.test-%d"
  description = "Test Pass Type"
}

resource "appleappstoreconnect_certificate" "test" {
  certificate_type = "PASS_TYPE_ID"
  csr_content     = "`+testCSRContent+`"

  relationships = {
    pass_type_id = appleappstoreconnect_pass_type_id.test.id
  }
}

data "appleappstoreconnect_certificate" "test" {
  id = appleappstoreconnect_certificate.test.id
}
`, timestamp)
}

func testAccCertificateDataSourceConfigByFilter(timestamp int64) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_pass_type_id" "test" {
  identifier  = "pass.io.truetickets.test.test-%d"
  description = "Test Pass Type"
}

resource "appleappstoreconnect_certificate" "test" {
  certificate_type = "PASS_TYPE_ID"
  csr_content     = "`+testCSRContent+`"

  relationships = {
    pass_type_id = appleappstoreconnect_pass_type_id.test.id
  }
}

data "appleappstoreconnect_certificate" "test" {
  filter {
    certificate_type = "PASS_TYPE_ID"
    serial_number   = appleappstoreconnect_certificate.test.serial_number
  }
}
`, timestamp)
}
