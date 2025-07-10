# Terraform Provider for Apple App Store Connect

This Terraform provider enables management of Apple App Store Connect
resources, with initial support for Pass Type IDs and Certificates used
in Apple Wallet pass development.

## Features

### Resources

- **Pass Type IDs**: Create and manage Pass Type identifiers for Apple
  Wallet passes
- **Certificates**: Create and manage certificates with Pass Type ID
  relationships, including automatic recreation before expiration

### Data Sources

- **Pass Type ID**: Retrieve information about existing Pass Type IDs
- **Certificate**: Retrieve information about a single certificate with
  filtering
- **Certificates**: List multiple certificates with filtering by type
  and display name

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >=
  1.0
- [Go](https://golang.org/doc/install) >= 1.23 (for development)
- Apple Developer account with App Store Connect API access
- API Key with appropriate permissions

## Installation

### Terraform 0.13+

```hcl
terraform {
  required_providers {
    appleappstoreconnect = {
      source  = "truetickets/appleappstoreconnect"
      version = "~> 0.1"
    }
  }
}
```

## Configuration

### Provider Configuration

```hcl
provider "appleappstoreconnect" {
  issuer_id   = "YOUR_ISSUER_ID"
  key_id      = "YOUR_KEY_ID"
  private_key = file("path/to/your/private_key.p8")
}
```

### Environment Variables

The provider can also be configured using environment variables:

```bash
export APP_STORE_CONNECT_ISSUER_ID="YOUR_ISSUER_ID"
export APP_STORE_CONNECT_KEY_ID="YOUR_KEY_ID"
export APP_STORE_CONNECT_PRIVATE_KEY="$(cat path/to/your/private_key.p8)"
```

## Usage Examples

### Create a Pass Type ID

```hcl
resource "appleappstoreconnect_pass_type_id" "membership" {
  identifier  = "pass.io.truetickets.test.membership"
  description = "Membership Pass"
}
```

### Create a Certificate for a Pass Type ID

```hcl
resource "appleappstoreconnect_certificate" "pass_cert" {
  certificate_type = "PASS_TYPE_ID"
  csr_content     = file("path/to/your/csr.pem")

  # Automatically recreate certificate 30 days before expiration (default)
  # Set to 0 to disable automatic recreation
  recreate_threshold = 2592000  # 30 days in seconds

  relationships = {
    pass_type_id = appleappstoreconnect_pass_type_id.membership.id
  }
}
```

### Certificate Auto-Renewal

The certificate resource supports automatic recreation before
expiration:

```hcl
resource "appleappstoreconnect_certificate" "auto_renew" {
  certificate_type = "PASS_TYPE_ID"
  csr_content     = tls_cert_request.example.cert_request_pem

  # Recreate 60 days before expiration
  recreate_threshold = 5184000  # 60 days in seconds

  relationships = {
    pass_type_id = appleappstoreconnect_pass_type_id.membership.id
  }
}

# Disable auto-renewal
resource "appleappstoreconnect_certificate" "manual_only" {
  certificate_type = "PASS_TYPE_ID"
  csr_content     = tls_cert_request.example.cert_request_pem

  # Disable automatic recreation
  recreate_threshold = 0

  relationships = {
    pass_type_id = appleappstoreconnect_pass_type_id.membership.id
  }
}
```

### List All Pass Type Certificates

```hcl
data "appleappstoreconnect_certificates" "pass_certs" {
  filter {
    certificate_type = "PASS_TYPE_ID"
  }
}

output "certificate_count" {
  value = length(data.appleappstoreconnect_certificates.pass_certs.certificates)
}
```

## Development

### Building the Provider

```shell
go build -o terraform-provider-appleappstoreconnect
```

### Running Tests

```shell
# Unit tests
go test ./...

# Acceptance tests (requires valid API credentials)
make testacc
```

### Generating Documentation

```shell
make generate
```

### Code Quality

The project uses pre-commit hooks to enforce code quality standards:

```shell
# Install pre-commit hooks
pre-commit install

# Run all pre-commit hooks
pre-commit run --all-files

# Format code
make fmt

# Run linter
make lint
```

#### Pre-commit Hooks

- **go fmt**: Formats Go code
- **go vet**: Checks Go code for suspicious constructs
- **golangci-lint**: Runs comprehensive Go linting
- **prettier**: Formats YAML and other files
- **yamllint**: Validates YAML syntax and formatting

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This provider is distributed under the
[Mozilla Public License 2.0](LICENSE).
