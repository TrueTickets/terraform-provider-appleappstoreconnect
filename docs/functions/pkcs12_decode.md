---
page_title: "<no value> <no value> - <no value>"
subcategory: ""
description: |-
    Decode PKCS12 format to certificate and private key
---

# <no value> (<no value>)

Decode PKCS12 format to certificate and private key

The `pkcs12_decode` function extracts the certificate and private key
from a PKCS12 (P12) file. This is useful when you have an existing P12
file and need to extract its contents for use in Terraform
configurations.

## Example Usage

### Basic Usage

```hcl
# Decode an existing P12 file
locals {
  p12_contents = provider::appleappstoreconnect::pkcs12_decode(
    filebase64("certificate.p12"),
    var.p12_password
  )
}

# Access the certificate and private key
output "certificate_pem" {
  value = local.p12_contents.certificate_pem
}

output "private_key_pem" {
  value     = local.p12_contents.private_key_pem
  sensitive = true
}
```

### Extract and Save Components

```hcl
# Read and decode P12 file
locals {
  p12_data = provider::appleappstoreconnect::pkcs12_decode(
    filebase64("existing_certificate.p12"),
    var.p12_password
  )
}

# Save certificate to file
resource "local_file" "certificate" {
  content  = local.p12_data.certificate_pem
  filename = "certificate.pem"
}

# Save private key to file (with restricted permissions)
resource "local_file" "private_key" {
  content         = local.p12_data.private_key_pem
  filename        = "private_key.pem"
  file_permission = "0600"
}

# Use extracted certificate in other resources
data "tls_certificate" "extracted" {
  content = local.p12_data.certificate_pem
}

output "certificate_info" {
  value = {
    subject      = data.tls_certificate.extracted.certificates[0].subject
    issuer       = data.tls_certificate.extracted.certificates[0].issuer
    not_after    = data.tls_certificate.extracted.certificates[0].not_after
    serial_number = data.tls_certificate.extracted.certificates[0].serial_number
  }
}
```

## Signature

```text
pkcs12_decode(pkcs12_base64 string, password string) object
```

## Arguments

1. `pkcs12_base64` (String) The PKCS12 data in base64 encoded format
2. `password` (String) Password to decrypt the PKCS12 file

The function returns an object with the following attributes:

- `certificate_pem` (String): The certificate in PEM format
- `private_key_pem` (String): The private key in PEM format (PKCS8)

## Notes

- The input must be base64 encoded PKCS12 data. Use `filebase64()` to
  read P12 files
- The private key is returned in PKCS8 format, which is compatible with
  most tools
- Both the certificate and private key are returned in PEM format
- Handle the private key with care as it contains sensitive
  cryptographic material
