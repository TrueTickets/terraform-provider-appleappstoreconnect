# Copilot Workspace Configuration

## Project Overview

Terraform provider for Apple App Store Connect API, focusing on Pass
Type IDs and Certificates management.

## Key Technologies

- Language: Go 1.23+
- Framework: Terraform Plugin Framework
- API: Apple App Store Connect REST API
- Authentication: JWT with ES256

## Code Conventions

### Naming

- Resources: `appleappstoreconnect_<resource_name>`
- Functions: CamelCase for exported, camelCase for internal
- Files: snake_case (e.g., `pass_type_id_resource.go`)

### Structure

- Models use `types.String`, not `string`
- All times in RFC3339 format
- Sensitive data marked in schema
- Thread-safe implementations required

### Testing

- Unit tests for all validations
- Acceptance tests for CRUD operations
- Mock external API calls in unit tests

## Common Patterns

### Resource Schema

```go
"field_name": schema.StringAttribute{
    MarkdownDescription: "Clear description",
    Required:            true,
    Validators: []validator.String{
        stringvalidator.LengthAtLeast(1),
    },
}
```

### API Calls

```go
resp, err := r.client.Do(ctx, Request{
    Method:   http.MethodPost,
    Endpoint: "/endpoint",
    Body:     requestBody,
})
```

### Error Handling

```go
if err != nil {
    resp.Diagnostics.AddError(
        "Client Error",
        fmt.Sprintf("Unable to perform action: %s", err),
    )
    return
}
```

## Quick Actions

### Before Committing

1. Run `make fmt`
2. Run `make lint`
3. Run `go test ./...`
4. Update documentation if needed

### Adding Features

1. Write tests first
2. Implement feature
3. Add examples
4. Generate documentation
5. Run all quality checks

## Useful Commands

- `make fmt` - Format code
- `make lint` - Run linter
- `make generate` - Generate docs
- `go test ./...` - Run tests
- `TF_LOG=DEBUG terraform apply` - Debug mode
