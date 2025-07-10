// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccCertificateResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCertificateResourceConfig("PASS_TYPE_ID", testCSRContent, time.Now().Unix()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appleappstoreconnect_certificate.test", "certificate_type", "PASS_TYPE_ID"),
					resource.TestCheckResourceAttrSet("appleappstoreconnect_certificate.test", "id"),
					resource.TestCheckResourceAttrSet("appleappstoreconnect_certificate.test", "certificate_content"),
					resource.TestCheckResourceAttrSet("appleappstoreconnect_certificate.test", "serial_number"),
					resource.TestCheckResourceAttrSet("appleappstoreconnect_certificate.test", "expiration_date"),
					// Test default recreate_threshold (30 days = 2592000 seconds)
					resource.TestCheckResourceAttr("appleappstoreconnect_certificate.test", "recreate_threshold", "2592000"),
				),
			},
			// Test with custom recreate_threshold (requires replacement since certificates are immutable)
			{
				Config: testAccCertificateResourceConfigWithThreshold("PASS_TYPE_ID", testCSRContent, 5184000, time.Now().Unix()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appleappstoreconnect_certificate.test", "certificate_type", "PASS_TYPE_ID"),
					resource.TestCheckResourceAttr("appleappstoreconnect_certificate.test", "recreate_threshold", "5184000"),
				),
				// Expect replacement since recreate_threshold requires recreation
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("appleappstoreconnect_certificate.test", plancheck.ResourceActionReplace),
					},
				},
			},
			// ImportState testing
			{
				ResourceName:            "appleappstoreconnect_certificate.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"csr_content", "recreate_threshold"}, // CSR and recreate_threshold are not returned by API
			},
		},
	})
}

func testAccCertificateResourceConfig(certType, csrContent string, timestamp int64) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_pass_type_id" "test" {
  identifier  = "pass.io.truetickets.test.test-%[1]d"
  description = "Test Pass Type"
}

resource "appleappstoreconnect_certificate" "test" {
  certificate_type = %[2]q
  csr_content     = %[3]q

  relationships = {
    pass_type_id = appleappstoreconnect_pass_type_id.test.id
  }
}
`, timestamp, certType, csrContent)
}

func testAccCertificateResourceConfigWithThreshold(certType, csrContent string, threshold int, timestamp int64) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_pass_type_id" "test" {
  identifier  = "pass.io.truetickets.test.test-%[1]d"
  description = "Test Pass Type"
}

resource "appleappstoreconnect_certificate" "test" {
  certificate_type = %[2]q
  csr_content     = %[3]q
  recreate_threshold = %[4]d

  relationships = {
    pass_type_id = appleappstoreconnect_pass_type_id.test.id
  }
}
`, timestamp, certType, csrContent, threshold)
}

// testCSRContent is a sample CSR for testing purposes
// In real usage, this would be generated using openssl or similar tools.
const testCSRContent = `-----BEGIN CERTIFICATE REQUEST-----
MIICgTCCAWkCAQAwPDEVMBMGA1UEChMMVHJ1ZSBUaWNrZXRzMSMwIQYDVQQDExpU
ZXJyYWZvcm0gVGVzdCBDZXJ0aWZpY2F0ZTCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAMBNAgIKKdSCFwwpVfP8IGfnEqhDDrS7t/FxU8Q3Tk+Lm58i/Bxy
bfZSansw8rfkuk9eb/mD4753oAbT9X7PYVVQnEon/6M1clNK0MMikkK4muFHzXjQ
0Ahbbm1HE42fYVOmjAhIwhJznRDuUPIpb5BLOmBe9uFR+VukGHtd9bLHlK0MMGtx
AmZmK4caxw9hIBYvl94TyXUT2epivf/JfQ2uuOKdPvWZ6rIP53ztYtU0DhrYNXzf
UQK12YytSmrp281vXTDM8ZGwMFiqHu6Frny+DuAb5UO+aatuR8cWiSb4j7HeMBY2
nR3SuA08BvfHTTI+S/MhCHwkUAJELb54Oc0CAwEAAaAAMA0GCSqGSIb3DQEBCwUA
A4IBAQBChLSYLjoNUgSyMk1uk5tRgocDhF6tNqoNZhQMaDDGDIWo4pTowQ2rzLaw
9l6sL9KGkoAc2+J3TMG0RM4NHbQu+SCZ7tDHVzHzU0mfK8oVfEru2DcRgSr0GB5t
4mActt8FK+pjEEgFdXO2RgXCMZHHdRLQW//gjUc3QAWVt9HFXZRX2qdbE/jCsT10
HDwc1FnL2tZDwvG7uEnMijQ1BZZ8mUpUbgcU6IAC4Yl7r6R/cwzFUhBdVUv+eYP0
SHo2I0sq4+K9vgupGZjw01WazfKv3krrpJMKIzg2x3cRxfw4cftn5n+iFk0DzeLO
lHf2Z+AlDD0ia5hoUIjmvq/INl8s
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
