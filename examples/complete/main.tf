# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    appleappstoreconnect = {
      source  = "truetickets/appleappstoreconnect"
      version = "~> 0.1"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

# Configure the Apple App Store Connect Provider
provider "appleappstoreconnect" {
  # These can also be set via environment variables:
  # APP_STORE_CONNECT_ISSUER_ID
  # APP_STORE_CONNECT_KEY_ID
  # APP_STORE_CONNECT_PRIVATE_KEY

  issuer_id   = var.app_store_connect_issuer_id
  key_id      = var.app_store_connect_key_id
  private_key = var.app_store_connect_private_key
}
