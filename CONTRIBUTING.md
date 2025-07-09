# Contributing to terraform-provider-appleappstoreconnect

Thank you for your interest in contributing to the Apple App Store
Connect Terraform Provider!

## Development Requirements

- Go 1.23+
- Terraform 1.0+ (or OpenTofu)
- Apple Developer account with App Store Connect API access
- Valid API credentials for testing

## Getting Started

1. Fork and clone the repository
2. Install dependencies:
    ```bash
    go mod download
    ```
3. Build the provider:
    ```bash
    go build -o terraform-provider-appleappstoreconnect
    ```

## Development Workflow

### 1. Make Your Changes

- Follow the existing code patterns
- Use the Terraform Plugin Framework (not SDK v2)
- Ensure thread-safety for parallel execution
- Add appropriate logging with `tflog`

### 2. Write Tests

- Add unit tests for new functionality
- Update acceptance tests if adding resources/data sources
- Ensure existing tests still pass

### 3. Run Quality Checks

```bash
# Format code
make fmt

# Run linter
make lint

# Run unit tests
go test ./...

# Run acceptance tests (requires API credentials)
TF_ACC=1 go test ./... -timeout 30m
```

### 4. Update Documentation

- Add/update templates in `templates/` directory
- Run `make generate` to regenerate documentation
- Update examples in `examples/` directory

## Code Style

- Follow standard Go conventions
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and testable

## Testing

### Unit Tests

- Test individual functions and methods
- Mock external API calls
- Aim for high code coverage

### Acceptance Tests

- Test actual API interactions
- Use `resource.Test` framework
- Clean up resources after tests

Set these environment variables for acceptance tests:

```bash
export APP_STORE_CONNECT_ISSUER_ID="your-issuer-id"
export APP_STORE_CONNECT_KEY_ID="your-key-id"
export APP_STORE_CONNECT_PRIVATE_KEY="$(cat path/to/key.p8)"
export TF_ACC=1
```

## Adding a New Resource

1. **Define Types** (`internal/provider/<resource>_types.go`):

    ```go
    type MyResourceModel struct {
        ID          types.String `tfsdk:"id"`
        Name        types.String `tfsdk:"name"`
        Description types.String `tfsdk:"description"`
    }
    ```

2. **Implement Resource** (`internal/provider/<resource>_resource.go`):

    - Implement CRUD operations
    - Add validation
    - Handle import functionality

3. **Add Tests** (`internal/provider/<resource>_resource_test.go`):

    - Unit tests for validation
    - Acceptance tests for CRUD operations

4. **Register Resource** (in `provider.go`):

    ```go
    func (p *AppleAppStoreConnectProvider) Resources(ctx context.Context) []func() resource.Resource {
        return []func() resource.Resource{
            // ... existing resources
            NewMyResource,
        }
    }
    ```

5. **Document** (`templates/resources/<resource>.md.tmpl`):
    - Add examples
    - Document all attributes
    - Include import instructions

## Commit Messages

Follow conventional commit format:

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Test additions/changes
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

Example:

```
feat: add support for push notification certificates

- Implement DEVELOPMENT_PUSH_SSL certificate type
- Add validation for certificate requirements
- Update documentation with examples
```

## Pull Request Process

1. Create a feature branch from `main`
2. Make your changes following the guidelines above
3. Ensure all tests pass and code is properly formatted
4. Update the CHANGELOG.md with your changes
5. Create a pull request with a clear description
6. Address any review feedback

## Common Issues

### Linting Errors

- Missing comments: Add godoc comments to exported types/functions
- Line length: Break long lines appropriately
- Formatting: Run `make fmt`

### Test Failures

- Check API credentials are valid
- Ensure test resources don't already exist
- Verify network connectivity to App Store Connect

### Documentation Generation

- Ensure Terraform is installed
- Templates must be valid Go template syntax
- Run `make generate` after template changes

## Questions?

If you have questions or need help:

1. Check existing issues and pull requests
2. Review the CLAUDE.md file for context
3. Open an issue for discussion

Thank you for contributing!
