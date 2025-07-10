resource "appleappstoreconnect_pass_type_id" "event_ticket" {
  identifier  = "pass.io.truetickets.test.eventticket"
  description = "Event Tickets"
}

resource "tls_private_key" "event_ticket" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_cert_request" "event_ticket" {
  private_key_pem = tls_private_key.event_ticket.private_key_pem

  subject {
    common_name  = "Terraform Test Certificate"
    organization = "True Tickets"
  }
}

resource "appleappstoreconnect_certificate" "event_ticket" {
  certificate_type = "PASS_TYPE_ID_WITH_NFC"
  csr_content      = tls_private_key.event_ticket.private_key_pem

  relationships = {
    pass_type_id = appleappstoreconnect_pass_type_id.event_ticket.id
  }
}

resource "local_file" "event_ticket_cert_file" {
  content  = appleappstoreconnect_certificate.event_ticket.certificate_content
  filename = "${path.module}/certificates/event_ticket.pem"
}
