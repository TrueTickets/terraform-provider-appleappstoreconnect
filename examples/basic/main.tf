# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    appleappstoreconnect = {
      source  = "truetickets/appleappstoreconnect"
      version = "~> 0.1"
    }
  }
}

# Configure the provider using environment variables:
# export APP_STORE_CONNECT_ISSUER_ID="your-issuer-id"
# export APP_STORE_CONNECT_KEY_ID="your-key-id"
# export APP_STORE_CONNECT_PRIVATE_KEY="$(cat /path/to/AuthKey.p8)"
provider "appleappstoreconnect" {}

# Create a Pass Type ID
resource "appleappstoreconnect_pass_type_id" "example" {
  identifier  = "pass.com.example.mypass"
  description = "My Example Pass"
}

# Create a certificate for the Pass Type ID
resource "appleappstoreconnect_certificate" "example" {
  certificate_type = "PASS_TYPE_ID"
  csr_content      = file("example.csr")

  relationships {
    pass_type_id = appleappstoreconnect_pass_type_id.example.id
  }
}

# Output the certificate
output "certificate_content" {
  value     = appleappstoreconnect_certificate.example.certificate_content
  sensitive = true
}

output "certificate_expiration" {
  value = appleappstoreconnect_certificate.example.expiration_date
}
