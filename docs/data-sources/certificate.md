---
page_title: "appleappstoreconnect_certificate Data Source - terraform-provider-appleappstoreconnect"
subcategory: ""
description: |-
  Use this data source to retrieve information about an existing Certificate in App Store Connect.
---

# appleappstoreconnect_certificate (Data Source)

Use this data source to retrieve information about an existing Certificate in App Store Connect.

## Example Usage

### By ID

```terraform
data "appleappstoreconnect_certificate" "example" {
  id = "1234567890"
}
```

### By Certificate Type and Serial Number

```terraform
data "appleappstoreconnect_certificate" "example" {
  filter {
    certificate_type = "PASS_TYPE_ID"
    serial_number   = "1234567890ABCDEF"
  }
}
```

### Using Data Source for Certificate Download

```terraform
data "appleappstoreconnect_certificate" "pass_cert" {
  filter {
    certificate_type = "PASS_TYPE_ID"
  }
}

# Save certificate to local file
resource "local_file" "cert" {
  content  = data.appleappstoreconnect_certificate.pass_cert.certificate_content
  filename = "pass_certificate.pem"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) The unique identifier of the Certificate. Conflicts with `filter`.
* `filter` - (Optional) Filter criteria for finding a Certificate. Conflicts with `id`. This block supports:
  * `certificate_type` - (Optional) The certificate type to filter by. Valid values are: `IOS_DEVELOPMENT`, `IOS_DISTRIBUTION`, `MAC_APP_DEVELOPMENT`, `MAC_APP_DISTRIBUTION`, `MAC_INSTALLER_DISTRIBUTION`, `PASS_TYPE_ID`, `PASS_TYPE_ID_WITH_NFC`, `DEVELOPER_ID_KEXT`, `DEVELOPER_ID_APPLICATION`, `DEVELOPMENT_PUSH_SSL`, `PRODUCTION_PUSH_SSL`, `PUSH_SSL`.
  * `serial_number` - (Optional) The serial number to search for.

~> **Note:** You must provide either `id` or `filter`, but not both. When using `filter`, ensure your criteria return exactly one certificate.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `certificate_type` - The type of certificate.
* `certificate_content` - (Sensitive) The certificate content in PEM format.
* `display_name` - The display name of the certificate.
* `name` - The name of the certificate.
* `platform` - The platform for the certificate.
* `serial_number` - The serial number of the certificate.
* `expiration_date` - The expiration date of the certificate.
* `relationships` - The relationships for the certificate.
  * `pass_type_id` - The ID of the associated Pass Type ID (if applicable).