// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCertificateResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCertificateResourceConfig("PASS_TYPE_ID", testCSRContent),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appleappstoreconnect_certificate.test", "certificate_type", "PASS_TYPE_ID"),
					resource.TestCheckResourceAttrSet("appleappstoreconnect_certificate.test", "id"),
					resource.TestCheckResourceAttrSet("appleappstoreconnect_certificate.test", "certificate_content"),
					resource.TestCheckResourceAttrSet("appleappstoreconnect_certificate.test", "serial_number"),
					resource.TestCheckResourceAttrSet("appleappstoreconnect_certificate.test", "expiration_date"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "appleappstoreconnect_certificate.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"csr_content"}, // CSR is not returned by API
			},
		},
	})
}

func testAccCertificateResourceConfig(certType, csrContent string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_pass_type_id" "test" {
  identifier  = "pass.com.example.test"
  description = "Test Pass Type"
}

resource "appleappstoreconnect_certificate" "test" {
  certificate_type = %[1]q
  csr_content     = %[2]q

  relationships {
    pass_type_id = appleappstoreconnect_pass_type_id.test.id
  }
}
`, certType, csrContent)
}

// testCSRContent is a sample CSR for testing purposes
// In real usage, this would be generated using openssl or similar tools.
const testCSRContent = `-----BEGIN CERTIFICATE REQUEST-----
MIICijCCAXICAQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUx
ITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAKcRlk+nk2XGRq1ge1fRKc5VVV8W8RDVK7mDWWkA
-----END CERTIFICATE REQUEST-----`

func TestCertificateTypeValidation(t *testing.T) {
	validTypes := []string{
		CertificateTypeIOSDevelopment,
		CertificateTypeIOSDistribution,
		CertificateTypeMacAppDevelopment,
		CertificateTypeMacAppDistribution,
		CertificateTypeMacInstallerDistribution,
		CertificateTypePassTypeID,
		CertificateTypePassTypeIDWithNFC,
		CertificateTypeDeveloperIDKext,
		CertificateTypeDeveloperIDApplication,
		CertificateTypeDevelopmentPushSSL,
		CertificateTypeProductionPushSSL,
		CertificateTypePushSSL,
	}

	for _, certType := range validTypes {
		t.Run(certType, func(t *testing.T) {
			// Just verify the constant is defined
			if certType == "" {
				t.Errorf("Certificate type constant %s is empty", certType)
			}
		})
	}
}
