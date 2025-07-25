# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    appleappstoreconnect = {
      source  = "truetickets/appleappstoreconnect"
      version = "~> 0.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

provider "appleappstoreconnect" {
  # These can also be set via environment variables:
  # APP_STORE_CONNECT_ISSUER_ID
  # APP_STORE_CONNECT_KEY_ID
  # APP_STORE_CONNECT_PRIVATE_KEY

  issuer_id   = var.app_store_connect_issuer_id
  key_id      = var.app_store_connect_key_id
  private_key = var.app_store_connect_private_key
}

# Create a Pass Type ID
resource "appleappstoreconnect_pass_type_id" "tf_test" {
  identifier  = "pass.io.truetickets.test.tf-test-2"
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

  # Recreate the certificate 60 days before expiration (default is 30 days)
  recreate_threshold = 5184000 # 60 days in seconds

  # Optional: provide private key and password for PKCS12 bundle generation
  private_key_pem        = tls_private_key.tf_test.private_key_pem
  pkcs12_bundle_password = "testpassword"

  relationships = {
    pass_type_id = appleappstoreconnect_pass_type_id.tf_test.id
  }
}


data "tls_certificate" "tf_test" {
  content = base64decode(appleappstoreconnect_certificate.tf_test.certificate_content_pem)
}

resource "random_password" "tf_test" {
  length  = 16
  special = true
}

# Output the certificate
output "certificate_content" {
  value     = appleappstoreconnect_certificate.tf_test.certificate_content
  sensitive = true
}

output "certificate_content_pem" {
  value     = appleappstoreconnect_certificate.tf_test.certificate_content_pem
  sensitive = true
}

output "certificate_ca_issuers" {
  value = appleappstoreconnect_certificate.tf_test.certificate_ca_issuers
}

output "pkcs12_bundle_content" {
  value     = appleappstoreconnect_certificate.tf_test.pkcs12_bundle_content
  sensitive = true
}

output "pkcs12_password" {
  value     = random_password.tf_test.result
  sensitive = true
}
