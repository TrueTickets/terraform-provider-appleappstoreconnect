# Copyright IBM Corp. 2025, 2026

# Register a Bundle ID (App ID) for an iOS app
resource "appleappstoreconnect_bundle_id" "app" {
  identifier = "io.truetickets.test.app"
  name       = "True Tickets App"
  platform   = "IOS"
}

# Look up the Bundle ID by its reverse-DNS identifier
data "appleappstoreconnect_bundle_id" "app" {
  filter = {
    identifier = "io.truetickets.test.app"
  }

  depends_on = [appleappstoreconnect_bundle_id.app]
}

output "bundle_id" {
  description = "The App Store Connect resource ID of the Bundle ID"
  value       = appleappstoreconnect_bundle_id.app.id
}
