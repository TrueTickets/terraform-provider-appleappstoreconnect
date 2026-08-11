# Copyright (c) TrueTickets, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "bundle_id" {
  description = "The bundle ID of the app to manage In-App Purchases for."
  type        = string
}

variable "in_app_purchases" {
  description = <<-EOT
    The In-App Purchases to create in the app, keyed by a logical name. The same
    map is passed to every app instance so the product definition lives in one
    place.
  EOT
  type = map(object({
    product_id           = string
    name                 = string
    in_app_purchase_type = string
    base_territory       = string
    # Customer price per territory, e.g. { USA = "9.99" }.
    prices                       = map(string)
    available_territories        = list(string)
    available_in_new_territories = optional(bool, true)
    localizations = map(object({
      name        = string
      description = optional(string)
    }))
  }))
  default = {}
}
