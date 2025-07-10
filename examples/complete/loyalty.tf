resource "appleappstoreconnect_pass_type_id" "loyalty" {
  identifier  = "pass.io.truetickets.test.loyalty"
  description = "Loyalty Program Cards"
}

resource "tls_private_key" "loyalty" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_cert_request" "loyalty" {
  private_key_pem = tls_private_key.loyalty.private_key_pem

  subject {
    common_name  = "Terraform Test Certificate"
    organization = "True Tickets"
  }
}

resource "appleappstoreconnect_certificate" "loyalty" {
  certificate_type = "PASS_TYPE_ID"
  csr_content      = tls_private_key.loyalty.private_key_pem

  relationships = {
    pass_type_id = appleappstoreconnect_pass_type_id.loyalty.id
  }
}

resource "local_file" "loyalty_cert_file" {
  content  = appleappstoreconnect_certificate.loyalty.certificate_content
  filename = "${path.module}/certificates/loyalty.pem"
}
