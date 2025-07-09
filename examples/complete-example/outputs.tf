# Copyright (c) HashiCorp, Inc.

output "pass_type_ids" {
  description = "Map of Pass Type IDs"
  value = {
    membership = {
      id         = appleappstoreconnect_pass_type_id.membership.id
      identifier = appleappstoreconnect_pass_type_id.membership.identifier
    }
    loyalty = {
      id         = appleappstoreconnect_pass_type_id.loyalty.id
      identifier = appleappstoreconnect_pass_type_id.loyalty.identifier
    }
    event_ticket = {
      id         = appleappstoreconnect_pass_type_id.event_ticket.id
      identifier = appleappstoreconnect_pass_type_id.event_ticket.identifier
    }
  }
}

output "certificate_info" {
  description = "Certificate information"
  value = {
    membership = {
      serial_number   = appleappstoreconnect_certificate.membership_cert.serial_number
      expiration_date = appleappstoreconnect_certificate.membership_cert.expiration_date
    }
    loyalty = {
      serial_number   = appleappstoreconnect_certificate.loyalty_cert.serial_number
      expiration_date = appleappstoreconnect_certificate.loyalty_cert.expiration_date
    }
    event_ticket = {
      serial_number   = appleappstoreconnect_certificate.event_ticket_cert.serial_number
      expiration_date = appleappstoreconnect_certificate.event_ticket_cert.expiration_date
      has_nfc         = appleappstoreconnect_certificate.event_ticket_cert.certificate_type == "PASS_TYPE_ID_WITH_NFC"
    }
  }
  sensitive = false
}

output "certificate_files" {
  description = "Paths to saved certificate files"
  value = {
    membership   = local_file.membership_cert_file.filename
    loyalty      = local_file.loyalty_cert_file.filename
    event_ticket = local_file.event_ticket_cert_file.filename
  }
}
