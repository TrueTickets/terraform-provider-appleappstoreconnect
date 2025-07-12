# Copyright (c) HashiCorp, Inc.

# Example of using data sources to retrieve existing resources

# Find a specific Pass Type ID by identifier
data "appleappstoreconnect_pass_type_id" "existing_membership" {
  filter = {
    identifier = "pass.io.truetickets.test.membership"
  }

  depends_on = [appleappstoreconnect_pass_type_id.membership]
}

# Find all Pass Type certificates
data "appleappstoreconnect_certificates" "all_pass_certs" {
  filter = {
    certificate_type = "PASS_TYPE_ID"
  }

  depends_on = [
    appleappstoreconnect_certificate.membership,
    appleappstoreconnect_certificate.loyalty,
  ]
}

# Find all NFC-enabled Pass Type certificates
data "appleappstoreconnect_certificates" "nfc_pass_certs" {
  filter = {
    certificate_type = "PASS_TYPE_ID_WITH_NFC"
  }

  depends_on = [
    appleappstoreconnect_certificate.event_ticket,
  ]
}

# Find a specific certificate by its serial number
data "appleappstoreconnect_certificate" "membership_lookup" {
  filter = {
    certificate_type = "PASS_TYPE_ID"
    serial_number    = appleappstoreconnect_certificate.membership.serial_number
  }

  depends_on = [
    appleappstoreconnect_certificate.membership,
  ]
}

# Output data source results
output "data_source_results" {
  description = "Results from data source lookups"
  value = {
    existing_membership_id  = data.appleappstoreconnect_pass_type_id.existing_membership.id
    total_pass_certificates = length(data.appleappstoreconnect_certificates.all_pass_certs.certificates)
    total_nfc_certificates  = length(data.appleappstoreconnect_certificates.nfc_pass_certs.certificates)
    membership_cert_found   = data.appleappstoreconnect_certificate.membership_lookup.id == appleappstoreconnect_certificate.membership.id
  }
}

# Test PEM conversion
output "certificate_pem_test" {
  description = "Test that PEM conversion works"
  sensitive   = true
  value = {
    has_pem_content = data.appleappstoreconnect_certificate.membership_lookup.certificate_content_pem != null
    pem_starts_with = substr(data.appleappstoreconnect_certificate.membership_lookup.certificate_content_pem, 0, 27)
    has_der_content = data.appleappstoreconnect_certificate.membership_lookup.certificate_content != null
  }
}
