# Find Certificate by type
data "appleappstoreconnect_certificate" "pass_cert" {
  filter {
    certificate_type = "PASS_TYPE_ID"
  }
}

# Save certificate to local file
resource "local_file" "cert" {
  content  = data.appleappstoreconnect_certificate.pass_cert.certificate_content
  filename = "pass_certificate.pem"
}

# Use certificate information
output "certificate_expiration" {
  value = data.appleappstoreconnect_certificate.pass_cert.expiration_date
}