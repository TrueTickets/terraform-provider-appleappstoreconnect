---
page_title: "appleappstoreconnect_certificate Resource - terraform-provider-appleappstoreconnect"
subcategory: ""
description: |-
  Manages a Certificate in App Store Connect.
---

# appleappstoreconnect_certificate (Resource)

Manages a Certificate in App Store Connect.

Certificates are used to sign your iOS or Mac apps for development and distribution. This resource supports creating various types of certificates including Pass Type ID certificates for Apple Wallet passes.

## Example Usage

### Basic iOS Development Certificate

```terraform
resource "appleappstoreconnect_certificate" "ios_dev" {
  certificate_type = "IOS_DEVELOPMENT"
  csr_content     = file("path/to/ios_dev.csr")
}
```

### Pass Type ID Certificate

```terraform
resource "appleappstoreconnect_pass_type_id" "membership" {
  identifier  = "pass.com.example.membership"
  description = "Membership Pass"
}

resource "appleappstoreconnect_certificate" "pass_cert" {
  certificate_type = "PASS_TYPE_ID"
  csr_content     = file("path/to/pass.csr")
  
  relationships {
    pass_type_id = appleappstoreconnect_pass_type_id.membership.id
  }
}
```

### Pass Type ID Certificate with NFC

```terraform
resource "appleappstoreconnect_certificate" "pass_nfc_cert" {
  certificate_type = "PASS_TYPE_ID_WITH_NFC"
  csr_content     = file("path/to/pass_nfc.csr")
  
  relationships {
    pass_type_id = appleappstoreconnect_pass_type_id.membership.id
  }
}
```

## Generating a Certificate Signing Request (CSR)

To create a certificate, you first need to generate a CSR. Here's an example using OpenSSL:

```bash
# Generate a private key
openssl genrsa -out private.key 2048

# Generate a CSR
openssl req -new -key private.key -out certificate.csr \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=CommonName"
```

## Argument Reference

The following arguments are supported:

* `certificate_type` - (Required, Forces new resource) The type of certificate to create. Valid values are:
  * `IOS_DEVELOPMENT` - iOS Development certificate
  * `IOS_DISTRIBUTION` - iOS Distribution certificate
  * `MAC_APP_DEVELOPMENT` - Mac App Development certificate
  * `MAC_APP_DISTRIBUTION` - Mac App Distribution certificate
  * `MAC_INSTALLER_DISTRIBUTION` - Mac Installer Distribution certificate
  * `PASS_TYPE_ID` - Pass Type ID certificate for Apple Wallet
  * `PASS_TYPE_ID_WITH_NFC` - Pass Type ID certificate with NFC capability
  * `DEVELOPER_ID_KEXT` - Developer ID KEXT certificate
  * `DEVELOPER_ID_APPLICATION` - Developer ID Application certificate
  * `DEVELOPMENT_PUSH_SSL` - Development Push SSL certificate
  * `PRODUCTION_PUSH_SSL` - Production Push SSL certificate
  * `PUSH_SSL` - Push SSL certificate

* `csr_content` - (Required, Sensitive, Forces new resource) The certificate signing request (CSR) content in PEM format.

* `relationships` - (Optional, Forces new resource) The relationships for the certificate. This block supports:
  * `pass_type_id` - (Optional) The ID of the Pass Type ID to associate with this certificate. Required for `PASS_TYPE_ID` and `PASS_TYPE_ID_WITH_NFC` certificate types.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the Certificate in App Store Connect.
* `certificate_content` - (Sensitive) The certificate content in PEM format.
* `display_name` - The display name of the certificate.
* `name` - The name of the certificate.
* `platform` - The platform for the certificate.
* `serial_number` - The serial number of the certificate.
* `expiration_date` - The expiration date of the certificate.

## Import

Certificates can be imported using the `id`, e.g.,

```
$ terraform import appleappstoreconnect_certificate.ios_dev 1234567890
```

Note: The CSR content cannot be recovered during import and will need to be provided in the configuration.