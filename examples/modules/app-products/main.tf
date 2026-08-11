# Copyright (c) TrueTickets, Inc.
# SPDX-License-Identifier: MPL-2.0

# Resolve the app by its bundle ID. This is what lets the same product
# definition target different apps (production, staging, ...).
data "appleappstoreconnect_app" "this" {
  bundle_id = var.bundle_id
}

locals {
  # Flatten product x locale into a single map for for_each.
  localizations = merge([
    for product_key, product in var.in_app_purchases : {
      for locale, loc in product.localizations :
      "${product_key}.${locale}" => {
        product     = product_key
        locale      = locale
        name        = loc.name
        description = loc.description
      }
    }
  ]...)

  # Flatten product x territory into a single map for price point lookups.
  prices = merge([
    for product_key, product in var.in_app_purchases : {
      for territory, customer_price in product.prices :
      "${product_key}.${territory}" => {
        product        = product_key
        territory      = territory
        customer_price = customer_price
      }
    }
  ]...)
}

resource "appleappstoreconnect_in_app_purchase" "this" {
  for_each = var.in_app_purchases

  app_id               = data.appleappstoreconnect_app.this.id
  product_id           = each.value.product_id
  name                 = each.value.name
  in_app_purchase_type = each.value.in_app_purchase_type
}

resource "appleappstoreconnect_in_app_purchase_localization" "this" {
  for_each = local.localizations

  in_app_purchase_id = appleappstoreconnect_in_app_purchase.this[each.value.product].id
  locale             = each.value.locale
  name               = each.value.name
  description        = each.value.description
}

# Resolve each (product, territory) customer price to a price point ID.
data "appleappstoreconnect_in_app_purchase_price_point" "this" {
  for_each = local.prices

  in_app_purchase_id = appleappstoreconnect_in_app_purchase.this[each.value.product].id
  territory          = each.value.territory
  customer_price     = each.value.customer_price
}

resource "appleappstoreconnect_in_app_purchase_price_schedule" "this" {
  for_each = var.in_app_purchases

  in_app_purchase_id = appleappstoreconnect_in_app_purchase.this[each.key].id
  base_territory     = each.value.base_territory

  manual_prices = [
    for territory, customer_price in each.value.prices : {
      price_point_id = data.appleappstoreconnect_in_app_purchase_price_point.this["${each.key}.${territory}"].id
      territory      = territory
    }
  ]
}

resource "appleappstoreconnect_in_app_purchase_availability" "this" {
  for_each = var.in_app_purchases

  in_app_purchase_id           = appleappstoreconnect_in_app_purchase.this[each.key].id
  available_in_new_territories = each.value.available_in_new_territories
  available_territories        = each.value.available_territories
}
