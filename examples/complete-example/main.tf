# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    appleappstoreconnect = {
      source  = "truetickets/appleappstoreconnect"
      version = "~> 0.1"
    }
  }
}

# Configure the Apple App Store Connect Provider
provider "appleappstoreconnect" {
  # These can also be set via environment variables:
  # APP_STORE_CONNECT_ISSUER_ID
  # APP_STORE_CONNECT_KEY_ID
  # APP_STORE_CONNECT_PRIVATE_KEY

  issuer_id   = var.app_store_connect_issuer_id
  key_id      = var.app_store_connect_key_id
  private_key = file(var.app_store_connect_private_key_path)
}

# Create Pass Type IDs for different pass types
resource "appleappstoreconnect_pass_type_id" "membership" {
  identifier  = "pass.com.example.membership"
  description = "Membership Cards"
}

resource "appleappstoreconnect_pass_type_id" "loyalty" {
  identifier  = "pass.com.example.loyalty"
  description = "Loyalty Program Cards"
}

resource "appleappstoreconnect_pass_type_id" "event_ticket" {
  identifier  = "pass.com.example.eventticket"
  description = "Event Tickets"
}

# Create certificates for each Pass Type ID
resource "appleappstoreconnect_certificate" "membership_cert" {
  certificate_type = "PASS_TYPE_ID"
  csr_content      = file("${path.module}/csr/membership.csr")

  relationships {
    pass_type_id = appleappstoreconnect_pass_type_id.membership.id
  }
}

resource "appleappstoreconnect_certificate" "loyalty_cert" {
  certificate_type = "PASS_TYPE_ID"
  csr_content      = file("${path.module}/csr/loyalty.csr")

  relationships {
    pass_type_id = appleappstoreconnect_pass_type_id.loyalty.id
  }
}

resource "appleappstoreconnect_certificate" "event_ticket_cert" {
  certificate_type = "PASS_TYPE_ID_WITH_NFC"
  csr_content      = file("${path.module}/csr/event_ticket.csr")

  relationships {
    pass_type_id = appleappstoreconnect_pass_type_id.event_ticket.id
  }
}

# Save certificates to local files
resource "local_file" "membership_cert_file" {
  content  = appleappstoreconnect_certificate.membership_cert.certificate_content
  filename = "${path.module}/certificates/membership.pem"
}

resource "local_file" "loyalty_cert_file" {
  content  = appleappstoreconnect_certificate.loyalty_cert.certificate_content
  filename = "${path.module}/certificates/loyalty.pem"
}

resource "local_file" "event_ticket_cert_file" {
  content  = appleappstoreconnect_certificate.event_ticket_cert.certificate_content
  filename = "${path.module}/certificates/event_ticket.pem"
}
