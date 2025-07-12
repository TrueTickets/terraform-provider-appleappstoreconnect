---
page_title: "<no value> <no value> - <no value>"
subcategory: ""
description: |-
    Encode certificate and private key to PKCS12 format
---

# <no value> (<no value>)

Encode certificate and private key to PKCS12 format

The `pkcs12_encode` function encodes a certificate and private key pair
into PKCS12 (P12) format. This is useful when you need to create a P12
file for code signing or other purposes that require both the
certificate and private key bundled together.

## Example Usage

### Basic Usage

```hcl
# Assuming you have certificate and key data
locals {
  # Encode certificate and private key to PKCS12
  pkcs12_data = provider::appleappstoreconnect::pkcs12_encode(
    file("certificate.pem"),
    file("private_key.pem"),
    "mypassword"
  )
}

# Save the PKCS12 file
resource "local_file" "p12_file" {
  content  = base64decode(local.pkcs12_data)
  filename = "certificate.p12"
}
```

### With Certificate Resource

```hcl
# Generate a private key
resource "tls_private_key" "example" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

# Create CSR
resource "tls_cert_request" "example" {
  private_key_pem = tls_private_key.example.private_key_pem

  subject {
    common_name         = "Example Certificate"
    organization        = "Example Org"
    organizational_unit = "Engineering"
  }
}

# Create certificate in App Store Connect
resource "appleappstoreconnect_certificate" "example" {
  certificate_type = "PASS_TYPE_ID"
  csr_content     = tls_cert_request.example.cert_request_pem

  relationships {
    pass_type_id = appleappstoreconnect_pass_type_id.example.id
  }
}

# Encode to PKCS12
locals {
  pkcs12_data = provider::appleappstoreconnect::pkcs12_encode(
    base64decode(appleappstoreconnect_certificate.example.certificate_content_pem),
    tls_private_key.example.private_key_pem,
    var.p12_password
  )
}

# Save as P12 file
resource "local_file" "certificate_p12" {
  content         = base64decode(local.pkcs12_data)
  filename        = "pass_certificate.p12"
  file_permission = "0600"
}
```

## Signature

```text
pkcs12_encode(certificate_pem string, private_key_pem string, password string) string
```

## Arguments

1. `certificate_pem` (String) The certificate in PEM format
2. `private_key_pem` (String) The private key in PEM format. Supports
   RSA, EC, and PKCS8 key formats
3. `password` (String) Password to protect the PKCS12 file

The function returns a base64 encoded string containing the PKCS12 data.
Use Terraform's `base64decode()` function to get the raw binary data for
saving to a file.

## Notes

- The function supports RSA and EC private keys
- Private keys can be in PKCS1, PKCS8, or EC format
- The resulting PKCS12 file is compatible with standard tools like
  OpenSSL and macOS Keychain
- Always store the password securely, preferably using Terraform
  variables or a secret management system
