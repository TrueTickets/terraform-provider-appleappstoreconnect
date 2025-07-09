# Pass Type ID Certificate
resource "appleappstoreconnect_pass_type_id" "membership" {
  identifier  = "pass.com.example.membership"
  description = "Membership Pass"
}

resource "appleappstoreconnect_certificate" "pass_cert" {
  certificate_type = "PASS_TYPE_ID"
  csr_content     = file("path/to/pass.csr")
  
  relationships {
    pass_type_id = appleappstoreconnect_pass_type_id.membership.id
  }
}