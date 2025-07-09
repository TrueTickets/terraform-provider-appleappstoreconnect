---
page_title: "appleappstoreconnect_certificates Data Source - terraform-provider-appleappstoreconnect"
subcategory: ""
description: |-
  Use this data source to retrieve a list of Certificates from App Store Connect.
---

# appleappstoreconnect_certificates (Data Source)

Use this data source to retrieve a list of Certificates from App Store Connect.

## Example Usage

### List All Certificates

```terraform
data "appleappstoreconnect_certificates" "all" {
}
```

### Filter by Certificate Type

```terraform
data "appleappstoreconnect_certificates" "pass_certs" {
  filter {
    certificate_type = "PASS_TYPE_ID"
  }
}
```

### Filter by Display Name

```terraform
data "appleappstoreconnect_certificates" "test_certs" {
  filter {
    display_name = "Test"
  }
}
```

### Combined Filters

```terraform
data "appleappstoreconnect_certificates" "specific_certs" {
  filter {
    certificate_type = "PASS_TYPE_ID"
    display_name     = "Production"
  }
}
```

### Export Certificate List

```terraform
data "appleappstoreconnect_certificates" "all_pass_certs" {
  filter {
    certificate_type = "PASS_TYPE_ID"
  }
}

# Output certificate IDs
output "pass_certificate_ids" {
  value = [for cert in data.appleappstoreconnect_certificates.all_pass_certs.certificates : cert.id]
}

# Check for expiring certificates
output "expiring_certificates" {
  value = [
    for cert in data.appleappstoreconnect_certificates.all_pass_certs.certificates :
    {
      id              = cert.id
      display_name    = cert.display_name
      expiration_date = cert.expiration_date
    }
    if can(cert.expiration_date) && timeadd(timestamp(), "30d") > cert.expiration_date
  ]
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) Filter criteria for listing certificates. This block supports:
  * `certificate_type` - (Optional) Filter by certificate type. Valid values are: `IOS_DEVELOPMENT`, `IOS_DISTRIBUTION`, `MAC_APP_DEVELOPMENT`, `MAC_APP_DISTRIBUTION`, `MAC_INSTALLER_DISTRIBUTION`, `PASS_TYPE_ID`, `PASS_TYPE_ID_WITH_NFC`, `DEVELOPER_ID_KEXT`, `DEVELOPER_ID_APPLICATION`, `DEVELOPMENT_PUSH_SSL`, `PRODUCTION_PUSH_SSL`, `PUSH_SSL`.
  * `display_name` - (Optional) Filter by display name (partial match). The filter performs a substring match, returning certificates whose display name contains the specified value.

~> **Note:** The data source retrieves up to 200 certificates per request, which is the maximum allowed by the App Store Connect API. If you have more than 200 certificates, you may need to use filters to retrieve specific subsets.

## Attributes Reference

The following attributes are exported:

* `certificates` - A list of certificates matching the filter criteria. Each certificate in the list contains:
  * `id` - The unique identifier of the Certificate.
  * `certificate_type` - The type of certificate.
  * `certificate_content` - (Sensitive) The certificate content in PEM format.
  * `display_name` - The display name of the certificate.
  * `name` - The name of the certificate.
  * `platform` - The platform for the certificate.
  * `serial_number` - The serial number of the certificate.
  * `expiration_date` - The expiration date of the certificate.
  * `relationships` - The relationships for the certificate.
    * `pass_type_id` - The ID of the associated Pass Type ID (if applicable).