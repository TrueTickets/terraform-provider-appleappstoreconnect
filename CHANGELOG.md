## 0.1.0 (Unreleased)

FEATURES:

- **New Resource:** `appleappstoreconnect_pass_type_id` - Manage Pass
  Type IDs for Apple Wallet passes
- **New Resource:** `appleappstoreconnect_certificate` - Manage
  certificates with Pass Type ID relationships
- **New Data Source:** `appleappstoreconnect_pass_type_id` - Retrieve
  information about a Pass Type ID
- **New Data Source:** `appleappstoreconnect_certificate` - Retrieve
  information about a certificate with filtering support
- **New Data Source:** `appleappstoreconnect_certificates` - List
  multiple certificates with filtering by type and display name
- **New Resource:** `appleappstoreconnect_in_app_purchase` - Manage
  In-App Purchases (consumable, non-consumable, non-renewing)
- **New Resource:**
  `appleappstoreconnect_in_app_purchase_localization` - Localized name
  and description for an In-App Purchase
- **New Resource:**
  `appleappstoreconnect_in_app_purchase_price_schedule` - Base territory
  and per-territory pricing for an In-App Purchase
- **New Resource:**
  `appleappstoreconnect_in_app_purchase_availability` - Territory
  availability for an In-App Purchase
- **New Resource:** `appleappstoreconnect_in_app_purchase_submission` -
  Submit an In-App Purchase for App Review
- **New Resource:** `appleappstoreconnect_subscription_group` - Manage
  a subscription group
- **New Resource:**
  `appleappstoreconnect_subscription_group_localization` - Localized
  name for a subscription group
- **New Resource:** `appleappstoreconnect_subscription` - Manage an
  auto-renewable subscription
- **New Resource:** `appleappstoreconnect_subscription_localization` -
  Localized name and description for a subscription
- **New Resource:** `appleappstoreconnect_subscription_price` -
  Per-territory price for a subscription
- **New Resource:**
  `appleappstoreconnect_subscription_group_submission` - Submit a
  subscription group for App Review
- **New Data Source:** `appleappstoreconnect_app` - Look up an app by
  ID or bundle ID
- **New Data Source:** `appleappstoreconnect_territories` - List App
  Store territories
- **New Data Source:** `appleappstoreconnect_in_app_purchase` -
  Retrieve an In-App Purchase by ID or app/product ID
- **New Data Source:**
  `appleappstoreconnect_in_app_purchase_price_point` - Resolve an
  In-App Purchase price point from a customer price and territory
- **New Data Source:** `appleappstoreconnect_subscription` - Retrieve a
  subscription by ID or group/product ID
- **New Data Source:** `appleappstoreconnect_subscription_price_point` -
  Resolve a subscription price point from a customer price and territory
- **New Data Source:**
  `appleappstoreconnect_subscription_availability` - Read the territory
  availability of a subscription

ENHANCEMENTS:

- **In-App Purchases & Subscriptions**: Full lifecycle management of
  In-App Purchases and auto-renewable subscriptions, plus an example
  module that applies a shared product catalog to multiple apps
  (e.g. production and staging) via `for_each`
- **API versioning**: The client now sends the API version (`/v1`,
  `/v2`) per endpoint, enabling the In-App Purchase/subscription
  endpoints that mix versions (no change to existing resources)
- **Pagination**: List reads follow `links.next` so results above the
  200-per-page cap (territories, price points) are fully retrieved
- **Drift detection**: A `404 Not Found` on read now removes the
  resource from state instead of erroring, so the next plan recreates it
- **Certificate Auto-Renewal**: Added `recreate_threshold` argument to
  `appleappstoreconnect_certificate` resource for automatic recreation
  before expiration
- Added pre-commit hooks for code quality enforcement
- Improved code formatting and linting compliance
- Added comprehensive test coverage for all components
- Enhanced documentation generation using OpenTofu instead of Terraform
- Certificate resource now properly handles Apple's API limitation for
  programmatic revocation

NOTES:

- Initial release of the Apple App Store Connect Terraform provider
- Supports JWT authentication with automatic token refresh
- All resources support import functionality
- Provider configuration can be set via provider block or environment
  variables
- Code quality enforced through pre-commit hooks (go fmt, go vet,
  golangci-lint, prettier, yamllint)
