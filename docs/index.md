---
page_title: "appleappstoreconnect Provider"
subcategory: ""
description: |-
  The Apple App Store Connect provider allows you to manage App Store Connect resources such as Pass Type IDs and Certificates for Apple Wallet pass development.
---

# appleappstoreconnect Provider

The Apple App Store Connect provider allows you to manage App Store Connect resources such as Pass Type IDs and Certificates for Apple Wallet pass development.

## Example Usage

```terraform
terraform {
  required_providers {
    appleappstoreconnect = {
      source  = "truetickets/appleappstoreconnect"
      version = "~> 0.1"
    }
  }
}

# Configure the Apple App Store Connect Provider
provider "appleappstoreconnect" {
  issuer_id   = "69a6de70-example-1234-5678-b66ef836a256"
  key_id      = "D383SF739"
  private_key = file("~/.appstoreconnect/private_key.p8")
}

# Create a Pass Type ID
resource "appleappstoreconnect_pass_type_id" "example" {
  identifier  = "pass.com.example.membership"
  description = "Example Membership Pass"
}

# Create a Certificate for the Pass Type ID
resource "appleappstoreconnect_certificate" "example" {
  certificate_type = "PASS_TYPE_ID"
  csr_content     = file("path/to/certificate_signing_request.csr")
  
  relationships {
    pass_type_id = appleappstoreconnect_pass_type_id.example.id
  }
}
```

## Authentication

The Apple App Store Connect provider uses JWT authentication. You need to:

1. Create an API key in App Store Connect
2. Download the private key (.p8 file)
3. Note the Issuer ID and Key ID

## Schema

### Required

- `issuer_id` (String) The issuer ID from the API keys page in App Store Connect. Can also be set via the `APP_STORE_CONNECT_ISSUER_ID` environment variable.
- `key_id` (String) The key ID from the API keys page in App Store Connect. Can also be set via the `APP_STORE_CONNECT_KEY_ID` environment variable.
- `private_key` (String, Sensitive) The private key contents (.p8 file) for App Store Connect API authentication. Can also be set via the `APP_STORE_CONNECT_PRIVATE_KEY` environment variable.