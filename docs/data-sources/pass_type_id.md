---
page_title: "appleappstoreconnect_pass_type_id Data Source - terraform-provider-appleappstoreconnect"
subcategory: ""
description: |-
  Use this data source to retrieve information about an existing Pass Type ID in App Store Connect.
---

# appleappstoreconnect_pass_type_id (Data Source)

Use this data source to retrieve information about an existing Pass Type ID in App Store Connect.

## Example Usage

### By ID

```terraform
data "appleappstoreconnect_pass_type_id" "example" {
  id = "1234567890"
}
```

### By Identifier

```terraform
data "appleappstoreconnect_pass_type_id" "example" {
  filter {
    identifier = "pass.com.example.membership"
  }
}
```

### Using Data Source Values

```terraform
data "appleappstoreconnect_pass_type_id" "membership" {
  filter {
    identifier = "pass.com.example.membership"
  }
}

resource "appleappstoreconnect_certificate" "pass_cert" {
  certificate_type = "PASS_TYPE_ID"
  csr_content     = file("path/to/pass.csr")
  
  relationships {
    pass_type_id = data.appleappstoreconnect_pass_type_id.membership.id
  }
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) The unique identifier of the Pass Type ID. Conflicts with `filter`.
* `filter` - (Optional) Filter criteria for finding a Pass Type ID. Conflicts with `id`. This block supports:
  * `identifier` - (Required) The identifier to search for (e.g., 'pass.com.example.membership').

~> **Note:** You must provide either `id` or `filter`, but not both.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `identifier` - The identifier for the Pass Type ID.
* `description` - The description of the Pass Type ID.
* `created_date` - The date when the Pass Type ID was created.