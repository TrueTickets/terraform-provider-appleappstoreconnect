# List all certificates
data "appleappstoreconnect_certificates" "all" {
}

# Filter by certificate type
data "appleappstoreconnect_certificates" "pass_certs" {
  filter {
    certificate_type = "PASS_TYPE_ID"
  }
}

# Filter by display name
data "appleappstoreconnect_certificates" "production_certs" {
  filter {
    display_name = "Production"
  }
}

# Output certificate information
output "all_certificate_count" {
  value = length(data.appleappstoreconnect_certificates.all.certificates)
}

output "pass_certificates" {
  value = [for cert in data.appleappstoreconnect_certificates.pass_certs.certificates : {
    id           = cert.id
    display_name = cert.display_name
    expires      = cert.expiration_date
  }]
}

# Save certificates to local files
resource "local_file" "pass_certs" {
  for_each = { for cert in data.appleappstoreconnect_certificates.pass_certs.certificates : cert.id => cert }
  
  content  = each.value.certificate_content
  filename = "certificates/pass_${each.value.id}.pem"
}