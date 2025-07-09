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

ENHANCEMENTS:

- Added pre-commit hooks for code quality enforcement
- Improved code formatting and linting compliance
- Added comprehensive test coverage for all components
- Enhanced documentation generation using OpenTofu instead of Terraform

NOTES:

- Initial release of the Apple App Store Connect Terraform provider
- Supports JWT authentication with automatic token refresh
- All resources support import functionality
- Provider configuration can be set via provider block or environment
  variables
- Code quality enforced through pre-commit hooks (go fmt, go vet,
  golangci-lint, prettier, yamllint)
