---
page_title: "appleappstoreconnect_pass_type_id Resource - terraform-provider-appleappstoreconnect"
subcategory: ""
description: |-
  Manages a Pass Type ID in App Store Connect.
---

# appleappstoreconnect_pass_type_id (Resource)

Manages a Pass Type ID in App Store Connect.

Pass Type IDs are used to identify different types of passes in Apple Wallet (formerly Passbook). Each pass type represents a specific kind of pass, such as boarding passes, event tickets, store cards, or generic passes.

## Example Usage

```terraform
resource "appleappstoreconnect_pass_type_id" "membership" {
  identifier  = "pass.com.example.membership"
  description = "Membership Pass for Example Company"
}

resource "appleappstoreconnect_pass_type_id" "event_ticket" {
  identifier  = "pass.com.example.events.ticket"
  description = "Event Ticket Pass"
}
```

## Argument Reference

The following arguments are supported:

* `identifier` - (Required, Forces new resource) The identifier for the Pass Type ID. This must be unique and follow reverse-DNS format, starting with `pass.` (e.g., `pass.com.example.membership`).
* `description` - (Required) A description of the Pass Type ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the Pass Type ID in App Store Connect.
* `created_date` - The date when the Pass Type ID was created.

## Import

Pass Type IDs can be imported using the `id`, e.g.,

```
$ terraform import appleappstoreconnect_pass_type_id.membership 1234567890
```