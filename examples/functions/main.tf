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

# Example 1: Generate certificate and encode to PKCS12
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

# Create a self-signed certificate for testing
resource "tls_self_signed_cert" "example" {
  private_key_pem = tls_private_key.example.private_key_pem

  subject {
    common_name         = "Test Certificate"
    organization        = "Test Org"
    organizational_unit = "Engineering"
    country             = "US"
    locality            = "San Francisco"
    province            = "CA"
  }

  validity_period_hours = 8760 # 1 year

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
  ]
}

# Encode to PKCS12
locals {
  pkcs12_data = provider::appleappstoreconnect::pkcs12_encode(
    tls_self_signed_cert.example.cert_pem,
    tls_private_key.example.private_key_pem,
    "testpassword"
  )
}

# Save as P12 file
resource "local_file" "certificate_p12" {
  content         = base64decode(local.pkcs12_data)
  filename        = "test_certificate.p12"
  file_permission = "0600"
}

# Example 2: Decode the PKCS12 file we just created
locals {
  decoded_cert = provider::appleappstoreconnect::pkcs12_decode(
    local.pkcs12_data,
    "testpassword"
  )
}

output "original_cert" {
  value = tls_self_signed_cert.example.cert_pem
}

output "decoded_cert" {
  value = local.decoded_cert.certificate_pem
}

output "decoded_key" {
  value     = local.decoded_cert.private_key_pem
  sensitive = true
}

output "pkcs12_file_path" {
  value = local_file.certificate_p12.filename
}

# Example 3: Read an existing P12 file
# Uncomment this if you have an existing P12 file to test with
# locals {
#   existing_p12 = provider::appleappstoreconnect::pkcs12_decode(
#     filebase64("existing_certificate.p12"),
#     "password"
#   )
# }
#
# output "existing_cert_info" {
#   value = {
#     certificate = local.existing_p12.certificate_pem
#     has_key     = local.existing_p12.private_key_pem != ""
#   }
# }
