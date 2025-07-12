# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    appleappstoreconnect = {
      source  = "truetickets/appleappstoreconnect"
      version = "1.0.3"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
  }
  required_version = ">= 1.8.0"
}

provider "appleappstoreconnect" {
  issuer_id   = var.issuer_id
  key_id      = var.key_id
  private_key = var.private_key
}

variable "issuer_id" {
  description = "App Store Connect API Issuer ID"
  type        = string
}

variable "key_id" {
  description = "App Store Connect API Key ID"
  type        = string
}

variable "private_key" {
  description = "App Store Connect API Private Key"
  type        = string
  sensitive   = true
}

# Example 1: Create Pass Type ID
resource "appleappstoreconnect_pass_type_id" "example" {
  identifier  = "pass.io.truetickets.test.membership"
  description = "Test Membership Pass"
}

# Example 2: Generate certificate with PKCS12 bundle
resource "tls_private_key" "example" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_cert_request" "example" {
  private_key_pem = tls_private_key.example.private_key_pem

  subject {
    common_name         = "Test Certificate"
    organization        = "Test Org"
    organizational_unit = "Engineering"
    country             = "US"
    locality            = "San Francisco"
    province            = "CA"
  }
}

# Create certificate with automatic PKCS12 bundle generation
resource "appleappstoreconnect_certificate" "example" {
  certificate_type = "PASS_TYPE_ID"
  csr_content      = tls_cert_request.example.cert_request_pem

  # Optional: provide private key and password for PKCS12 bundle generation
  private_key_pem        = tls_private_key.example.private_key_pem
  pkcs12_bundle_password = "testpassword"

  relationships = {
    pass_type_id = appleappstoreconnect_pass_type_id.example.id
  }
}

# Save PKCS12 bundle as P12 file if available
resource "local_file" "certificate_p12" {
  count = appleappstoreconnect_certificate.example.pkcs12_bundle_content != null ? 1 : 0

  content         = base64decode(appleappstoreconnect_certificate.example.pkcs12_bundle_content)
  filename        = "test_certificate.p12"
  file_permission = "0600"
}

# Outputs
output "certificate_id" {
  value = appleappstoreconnect_certificate.example.id
}

output "certificate_pem" {
  value     = appleappstoreconnect_certificate.example.certificate_content_pem
  sensitive = true
}

output "pkcs12_available" {
  value = appleappstoreconnect_certificate.example.pkcs12_bundle_content != null
}

output "pkcs12_file_path" {
  value = length(local_file.certificate_p12) > 0 ? local_file.certificate_p12[0].filename : null
}

# Example 3: Certificate without PKCS12 bundle
resource "appleappstoreconnect_certificate" "no_pkcs12" {
  certificate_type = "PASS_TYPE_ID"
  csr_content      = tls_cert_request.example.cert_request_pem

  relationships = {
    pass_type_id = appleappstoreconnect_pass_type_id.example.id
  }
}
