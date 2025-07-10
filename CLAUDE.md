# Claude Context for terraform-provider-appleappstoreconnect

## Overview

This is a Terraform provider for Apple App Store Connect API,
specifically focused on managing Pass Type IDs and Certificates for
Apple Wallet pass development.

## Key Components

### Provider Configuration

- Uses JWT authentication with Apple App Store Connect API
- Requires: Issuer ID, Key ID, and Private Key (.p8 file)
- Supports environment variables: `APP_STORE_CONNECT_ISSUER_ID`,
  `APP_STORE_CONNECT_KEY_ID`, `APP_STORE_CONNECT_PRIVATE_KEY`

### Resources

#### `appleappstoreconnect_pass_type_id`

- Manages Pass Type IDs required for Apple Wallet passes
- Attributes:
    - `identifier` (required): Reverse-DNS format (e.g.,
      `pass.io.truetickets.test.membership`)
    - `description` (required): Human-readable description
    - `id` (computed): API resource ID
    - `created_date` (computed): Creation timestamp

#### `appleappstoreconnect_certificate`

- Manages certificates for signing passes
- Attributes:
    - `certificate_type` (required): Type of certificate (e.g.,
      `PASS_TYPE_ID`, `PASS_TYPE_ID_WITH_NFC`)
    - `csr_content` (required): Certificate Signing Request in PEM
      format
    - `relationships.pass_type_id` (required for pass certificates):
      Associated Pass Type ID
    - Various computed attributes: `serial_number`, `expiration_date`,
      `certificate_content`, etc.

### Data Sources

- `appleappstoreconnect_pass_type_id`: Find Pass Type ID by ID or
  identifier
- `appleappstoreconnect_certificate`: Find single certificate by ID or
  filters
- `appleappstoreconnect_certificates`: List multiple certificates with
  filtering

## Development Guidelines

### Code Structure

- Uses Terraform Plugin Framework (not SDK v2)
- Main code in `internal/provider/`
- JWT client with automatic token refresh (20-minute expiration,
  5-minute buffer)
- Thread-safe implementation for parallel execution

### Testing

- Unit tests: `go test ./...`
- Acceptance tests: `TF_ACC=1 go test ./...` (requires valid API
  credentials)
- Linting: `make lint`
- Formatting: `make fmt`

### Building

```bash
go build -o terraform-provider-appleappstoreconnect
```

### Documentation

- Generated using `make generate`
- Templates in `templates/` directory
- Output in `docs/` directory

## Common Tasks

### Adding a New Resource

1. Create type definitions in `internal/provider/<resource>_types.go`
2. Implement resource in `internal/provider/<resource>_resource.go`
3. Add tests in `internal/provider/<resource>_resource_test.go`
4. Register in `provider.go`
5. Create documentation template in `templates/resources/`
6. Run `make generate` to generate docs

### Running Linter

Always run the linter after making changes:

```bash
make lint
```

### Certificate Types

The provider supports these certificate types:

- `PASS_TYPE_ID` - Standard Pass Type ID certificate
- `PASS_TYPE_ID_WITH_NFC` - Pass Type ID certificate with NFC
  capabilities
- `IOS_DEVELOPMENT`, `IOS_DISTRIBUTION` - iOS app certificates
- `MAC_APP_DEVELOPMENT`, `MAC_APP_DISTRIBUTION` - Mac app certificates
- `DEVELOPER_ID_KEXT`, `DEVELOPER_ID_APPLICATION` - Developer ID
  certificates
- `DEVELOPMENT_PUSH_SSL`, `PRODUCTION_PUSH_SSL`, `PUSH_SSL` - Push
  notification certificates

### Import Functionality

All resources support import:

```bash
terraform import appleappstoreconnect_pass_type_id.example <PASS_TYPE_ID>
terraform import appleappstoreconnect_certificate.example <CERTIFICATE_ID>
```

## API Limitations

- Maximum 200 certificates per API request
- Pass Type identifiers must be unique across all Apple Developer
  accounts
- Certificates expire after 1 year and must be renewed
- Rate limiting may apply to API requests

## Error Handling

- All API errors are returned with descriptive messages
- Validation occurs both client-side and server-side
- Network errors trigger automatic retries with exponential backoff

## Security Considerations

- Never commit private keys or API credentials
- Use environment variables or secure secret management
- Certificate content is marked as sensitive in Terraform state
- All API communication uses HTTPS with JWT authentication
