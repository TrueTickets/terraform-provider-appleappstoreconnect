# Examples

This directory contains various examples demonstrating how to use the
Apple App Store Connect Terraform provider.

## Example Categories

### Quick Start Examples

- **[basic/](basic/)** - A minimal example to get started quickly with
  creating a Pass Type ID and Certificate
- **[complete-example/](complete-example/)** - A comprehensive example
  showing multiple resources, data sources, and best practices
- **[import/](import/)** - Demonstrates how to import existing App Store
  Connect resources into Terraform

### Documentation Examples

The document generation tool looks for files in the following locations.
These examples are used in the provider documentation:

- **provider/provider.tf** - Example file for the provider index page
- **data-sources/`full data source name`/data-source.tf** - Example file
  for the named data source page
- **resources/`full resource name`/resource.tf** - Example file for the
  named resource page

## Getting Started

1. Choose an example directory based on your needs
2. Follow the README.md instructions in that directory
3. Set up your App Store Connect API credentials
4. Run `terraform init` and `terraform apply`

## Prerequisites

All examples require:

- An Apple Developer account with App Store Connect API access
- API Key created in App Store Connect (Issuer ID, Key ID, and Private
  Key .p8 file)
- Terraform >= 1.0
- For certificate creation: Certificate Signing Request (CSR) files

## API Credentials Setup

You can provide API credentials in two ways:

### Option 1: Environment Variables (Recommended)

```bash
export APP_STORE_CONNECT_ISSUER_ID="69a6de70-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
export APP_STORE_CONNECT_KEY_ID="XXXXXXXXXX"
export APP_STORE_CONNECT_PRIVATE_KEY="$(cat ~/path/to/AuthKey_XXXXXXXXXX.p8)"
```

### Option 2: Provider Configuration

```hcl
provider "appleappstoreconnect" {
  issuer_id   = var.issuer_id
  key_id      = var.key_id
  private_key = file("path/to/AuthKey.p8")
}
```

## Common Use Cases

- **Creating Pass Type IDs**: Required for Apple Wallet passes
- **Managing Certificates**: For signing passes and push notifications
- **Automation**: Integrate pass infrastructure with your CI/CD pipeline
- **Multi-environment**: Manage development and production passes
  separately
