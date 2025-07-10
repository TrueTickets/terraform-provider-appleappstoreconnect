# Complete Example - Apple App Store Connect Terraform Provider

This example demonstrates comprehensive usage of the Apple App Store
Connect Terraform provider, including:

- Creating multiple Pass Type IDs for different use cases
- Creating certificates with and without NFC capabilities
- Using data sources to retrieve existing resources
- Exporting certificates to local files
- Organizing outputs for easy reference

## Prerequisites

1. An Apple Developer account with App Store Connect API access
2. API credentials (Issuer ID, Key ID, and Private Key)
3. Certificate Signing Requests (CSR files) for each certificate

## Setup

1. Create the required directories:

    ```bash
    mkdir -p csr certificates
    ```

2. Generate Certificate Signing Requests for each pass type:

    ```bash
    # Generate private keys
    openssl genrsa -out csr/membership.key 2048
    openssl genrsa -out csr/loyalty.key 2048
    openssl genrsa -out csr/event_ticket.key 2048

    # Generate CSRs
    openssl req -new -key csr/membership.key -out csr/membership.csr \
      -subj "/C=US/ST=State/L=City/O=Organization/CN=Membership Pass"

    openssl req -new -key csr/loyalty.key -out csr/loyalty.csr \
      -subj "/C=US/ST=State/L=City/O=Organization/CN=Loyalty Pass"

    openssl req -new -key csr/event_ticket.key -out csr/event_ticket.csr \
      -subj "/C=US/ST=State/L=City/O=Organization/CN=Event Ticket Pass"
    ```

3. Create a `terraform.tfvars` file with your credentials:
    ```hcl
    app_store_connect_issuer_id = "YOUR_ISSUER_ID"
    app_store_connect_key_id    = "YOUR_KEY_ID"
    app_store_connect_private_key_path = "/path/to/your/AuthKey.p8"
    ```

## Usage

1. Initialize Terraform:

    ```bash
    terraform init
    ```

2. Review the planned changes:

    ```bash
    terraform plan
    ```

3. Apply the configuration:
    ```bash
    terraform apply
    ```

## What This Example Creates

### Resources

1. **Pass Type IDs**:

    - `membership`: For membership cards
    - `loyalty`: For loyalty program cards
    - `event_ticket`: For event tickets with NFC support

2. **Certificates**:

    - Standard certificate for membership Pass Type ID
    - Standard certificate for loyalty Pass Type ID
    - NFC-enabled certificate for event ticket Pass Type ID

3. **Local Files**:
    - Certificates are automatically saved to the `certificates/`
      directory

### Data Sources

The example also demonstrates how to use data sources to:

- Look up existing Pass Type IDs by identifier
- List all certificates of a specific type
- Find certificates by serial number

## Outputs

After applying, you'll see outputs including:

- Pass Type ID information (IDs and identifiers)
- Certificate details (serial numbers and expiration dates)
- Paths to saved certificate files
- Results from data source lookups

## Clean Up

To remove all created resources:

```bash
terraform destroy
```

## Important Notes

- Pass Type identifiers must follow the format
  `pass.io.truetickets.test.name`
- Certificates expire after one year and need to be renewed
- NFC-enabled certificates (`PASS_TYPE_ID_WITH_NFC`) are required for
  passes that use NFC features
- Keep your private keys and certificates secure
