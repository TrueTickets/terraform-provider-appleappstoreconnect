# Basic Example - Apple App Store Connect Terraform Provider

This is a minimal example to get started with the Apple App Store
Connect Terraform provider.

## Prerequisites

1. Apple Developer account with App Store Connect API access
2. API Key created in App Store Connect with appropriate permissions
3. A Certificate Signing Request (CSR) file

## Quick Start

1. Set your credentials as environment variables:

    ```bash
    export APP_STORE_CONNECT_ISSUER_ID="69a6de70-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    export APP_STORE_CONNECT_KEY_ID="XXXXXXXXXX"
    export APP_STORE_CONNECT_PRIVATE_KEY="$(cat ~/path/to/AuthKey_XXXXXXXXXX.p8)"
    ```

2. Generate a Certificate Signing Request:

    ```bash
    # Generate a private key
    openssl genrsa -out example.key 2048

    # Generate the CSR
    openssl req -new -key example.key -out example.csr \
      -subj "/C=US/ST=State/L=City/O=Organization/CN=Example Pass"
    ```

3. Initialize and apply:
    ```bash
    terraform init
    terraform apply
    ```

## What This Creates

- A Pass Type ID with identifier `pass.io.truetickets.test.mypass`
- A certificate for signing passes with that Pass Type ID

## Next Steps

- Save the certificate output to a file for use in your pass signing
  process
- Check the `complete-example` directory for more advanced usage
- Refer to the provider documentation for all available options
