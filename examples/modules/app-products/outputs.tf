# Copyright (c) TrueTickets, Inc.
# SPDX-License-Identifier: MPL-2.0

output "app_id" {
  description = "The resolved App Store Connect app ID."
  value       = data.appleappstoreconnect_app.this.id
}

output "in_app_purchase_ids" {
  description = "Map of logical product name to created In-App Purchase ID."
  value       = { for k, iap in appleappstoreconnect_in_app_purchase.this : k => iap.id }
}
