# Copyright (c) HashiCorp, Inc.

# Example of using data sources to retrieve existing resources

# Find a specific Pass Type ID by identifier
data "appleappstoreconnect_pass_type_id" "existing_membership" {
  filter {
    identifier = "pass.com.example.membership"
  }

  depends_on = [appleappstoreconnect_pass_type_id.membership]
}

# Find all Pass Type certificates
data "appleappstoreconnect_certificates" "all_pass_certs" {
  filter {
    certificate_type = "PASS_TYPE_ID"
  }

  depends_on = [
    appleappstoreconnect_certificate.membership_cert,
    appleappstoreconnect_certificate.loyalty_cert
  ]
}

# Find all NFC-enabled Pass Type certificates
data "appleappstoreconnect_certificates" "nfc_pass_certs" {
  filter {
    certificate_type = "PASS_TYPE_ID_WITH_NFC"
  }

  depends_on = [appleappstoreconnect_certificate.event_ticket_cert]
}

# Find a specific certificate by its serial number
data "appleappstoreconnect_certificate" "membership_cert_lookup" {
  filter {
    certificate_type = "PASS_TYPE_ID"
    serial_number    = appleappstoreconnect_certificate.membership_cert.serial_number
  }

  depends_on = [appleappstoreconnect_certificate.membership_cert]
}

# Output data source results
output "data_source_results" {
  description = "Results from data source lookups"
  value = {
    existing_membership_id  = data.appleappstoreconnect_pass_type_id.existing_membership.id
    total_pass_certificates = length(data.appleappstoreconnect_certificates.all_pass_certs.certificates)
    total_nfc_certificates  = length(data.appleappstoreconnect_certificates.nfc_pass_certs.certificates)
    membership_cert_found   = data.appleappstoreconnect_certificate.membership_cert_lookup.id == appleappstoreconnect_certificate.membership_cert.id
  }
}
