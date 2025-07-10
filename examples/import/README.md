# Import Example - Apple App Store Connect Terraform Provider

This example demonstrates how to import existing App Store Connect
resources into Terraform management.

## Overview

When you have existing Pass Type IDs and Certificates in App Store
Connect that were created outside of Terraform, you can import them to
bring them under Terraform management.

## Step 1: Discover Existing Resources

First, use data sources to find the IDs of existing resources:

```bash
terraform init
terraform apply
```

This will output the discovered Pass Type IDs and Certificates.

## Step 2: Create Resource Blocks

Uncomment the resource blocks in `main.tf` and update them with the
correct values from your existing resources.

## Step 3: Import Resources

### Import a Pass Type ID

```bash
terraform import appleappstoreconnect_pass_type_id.imported XXXXXXXXXX
```

Where `XXXXXXXXXX` is the ID of the Pass Type ID (not the identifier
like `pass.io.truetickets.test`).

### Import a Certificate

```bash
terraform import appleappstoreconnect_certificate.imported YYYYYYYYYY
```

Where `YYYYYYYYYY` is the ID of the Certificate.

## Step 4: Verify Import

After importing, run:

```bash
terraform plan
```

You should see no changes required if the import was successful and your
resource configuration matches the actual state.

## Important Notes

1. **Pass Type ID Import**: The ID to use for import is the App Store
   Connect resource ID (like `6MXXXXXXXX`), not the pass identifier
   (like `pass.io.truetickets.test.mypass`).

2. **Certificate Import**:

    - The `csr_content` field is required in the configuration but
      ignored during import
    - Make sure the `certificate_type` and `relationships` match the
      actual certificate

3. **Finding IDs**: You can find resource IDs by:
    - Using the data sources as shown in this example
    - Checking the App Store Connect web interface
    - Using the App Store Connect API directly

## Example Import Commands

```bash
# Import a Pass Type ID
terraform import appleappstoreconnect_pass_type_id.membership 6M7XXXXXXX

# Import a Certificate
terraform import appleappstoreconnect_certificate.membership_cert 7N8YYYYYYY
```

## Troubleshooting

If import fails:

1. Verify the resource ID is correct
2. Ensure your API credentials have permission to access the resource
3. Check that the resource configuration matches the actual resource
   type
