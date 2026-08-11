# Copyright (c) TrueTickets, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    appleappstoreconnect = {
      source = "TrueTickets/appleappstoreconnect"
    }
  }
}

# Credentials are read from APP_STORE_CONNECT_ISSUER_ID, APP_STORE_CONNECT_KEY_ID
# and APP_STORE_CONNECT_PRIVATE_KEY.
provider "appleappstoreconnect" {}

locals {
  # The product catalog, defined ONCE. Editing this and running `terraform apply`
  # updates the products in every app below.
  in_app_purchases = {
    premium = {
      product_id                   = "com.example.app.premium"
      name                         = "Premium"
      in_app_purchase_type         = "NON_CONSUMABLE"
      base_territory               = "USA"
      prices                       = { USA = "9.99" }
      available_territories        = ["USA", "FRA", "GBR"]
      available_in_new_territories = true
      localizations = {
        "en-US" = { name = "Premium", description = "Unlock all premium features." }
        "fr-FR" = { name = "Premium", description = "Débloquez toutes les fonctionnalités premium." }
      }
    }
  }

  # The apps to apply the catalog to, by bundle ID. Apple binds an In-App
  # Purchase to one app, so the catalog is created independently in each.
  apps = {
    prod    = "com.example.app"
    staging = "com.example.app.staging"
  }
}

module "products" {
  source   = "../modules/app-products"
  for_each = local.apps

  bundle_id        = each.value
  in_app_purchases = local.in_app_purchases
}

output "app_ids" {
  value = { for k, m in module.products : k => m.app_id }
}
