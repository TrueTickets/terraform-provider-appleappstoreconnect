# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    appleappstoreconnect = {
      source  = "truetickets/appleappstoreconnect"
      version = "~> 0.1"
    }
  }
}

provider "appleappstoreconnect" {
  # Configure via environment variables or provider block
}

# First, use data sources to discover existing resources
data "appleappstoreconnect_pass_type_id" "existing" {
  filter {
    identifier = "pass.io.truetickets.test.existing"
  }
}

data "appleappstoreconnect_certificates" "existing" {
  filter {
    certificate_type = "PASS_TYPE_ID"
  }
}

# Output discovered resources
output "discovered_pass_type_id" {
  value = data.appleappstoreconnect_pass_type_id.existing.id
}

output "discovered_certificates" {
  value = [for cert in data.appleappstoreconnect_certificates.existing.certificates : {
    id           = cert.id
    display_name = cert.display_name
    expires      = cert.expiration_date
  }]
}

# After discovering resources, you can import them:
#
# resource "appleappstoreconnect_pass_type_id" "imported" {
#   identifier  = "pass.io.truetickets.test.existing"
#   description = "Imported Pass Type"
# }
#
# resource "appleappstoreconnect_certificate" "imported" {
#   certificate_type = "PASS_TYPE_ID"
#   csr_content     = "dummy" # This will be ignored during import
#
#   relationships {
#     pass_type_id = "XXXXXXXXXX"
#   }
# }
