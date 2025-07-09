# Find Pass Type ID by identifier
data "appleappstoreconnect_pass_type_id" "membership" {
  filter {
    identifier = "pass.com.example.membership"
  }
}

# Use the data source to create a certificate
resource "appleappstoreconnect_certificate" "pass_cert" {
  certificate_type = "PASS_TYPE_ID"
  csr_content     = file("path/to/pass.csr")
  
  relationships {
    pass_type_id = data.appleappstoreconnect_pass_type_id.membership.id
  }
}