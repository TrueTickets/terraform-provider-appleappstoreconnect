resource "appleappstoreconnect_pass_type_id" "membership" {
  identifier  = "pass.io.truetickets.test.membership"
  description = "Membership Cards"
}

resource "tls_private_key" "membership" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_cert_request" "membership" {
  private_key_pem = tls_private_key.membership.private_key_pem

  subject {
    common_name  = "Terraform Test Certificate"
    organization = "True Tickets"
  }
}

resource "appleappstoreconnect_certificate" "membership" {
  certificate_type = "PASS_TYPE_ID"
  csr_content      = tls_private_key.membership.private_key_pem

  relationships = {
    pass_type_id = appleappstoreconnect_pass_type_id.membership.id
  }
}

# Save certificates to local files
resource "local_file" "membership_cert_file" {
  content  = appleappstoreconnect_certificate.membership.certificate_content
  filename = "${path.module}/certificates/membership.pem"
}
