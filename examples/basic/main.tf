# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    appleappstoreconnect = {
      source  = "truetickets/appleappstoreconnect"
      version = "~> 0.0"
    }
    corefunc = {
      source  = "northwood-labs/corefunc"
      version = "~> 1.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

# Configure the provider using environment variables:
# export APP_STORE_CONNECT_ISSUER_ID="your-issuer-id"
# export APP_STORE_CONNECT_KEY_ID="your-key-id"
# export APP_STORE_CONNECT_PRIVATE_KEY="$(cat /path/to/AuthKey.p8)"
provider "appleappstoreconnect" {}

# Create a Pass Type ID
resource "appleappstoreconnect_pass_type_id" "tf_test" {
  identifier  = "pass.io.truetickets.tf-test-2"
  description = "Terraform Test Pass Type ID"
}

resource "tls_private_key" "tf_test" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_cert_request" "tf_test" {
  private_key_pem = tls_private_key.tf_test.private_key_pem

  subject {
    common_name  = "Terraform Test Certificate"
    organization = "True Tickets"
  }
}

# Create a certificate for the Pass Type ID
resource "appleappstoreconnect_certificate" "tf_test" {
  certificate_type = "PASS_TYPE_ID"
  csr_content      = tls_cert_request.tf_test.cert_request_pem

  relationships = {
    pass_type_id = appleappstoreconnect_pass_type_id.tf_test.id
  }
}

# Output the certificate
output "certificate_content" {
  value     = appleappstoreconnect_certificate.tf_test.certificate_content
  sensitive = true
}

output "certificate_expiration" {
  value = appleappstoreconnect_certificate.tf_test.expiration_date
}
